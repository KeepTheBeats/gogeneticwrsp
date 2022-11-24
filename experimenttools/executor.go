package experimenttools

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/KeepTheBeats/routing-algorithms/random"
	"github.com/wcharczuk/go-chart"
	"go/build"
	"gogeneticwrsp/algorithms"
	"io/ioutil"
	"log"
	"math"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"gogeneticwrsp/model"
)

func getFilePaths(numCloud, numApp int, appSuffix string) (string, string) {
	return getCloudFilePaths(numCloud), getAppFilePaths(numApp, appSuffix)
}

// GenerateCloudsApps generates clouds and apps and writes them into files.
func GenerateCloudsApps(numCloud, numApp int, suffix string, taskProportion float64) {
	GenerateClouds(numCloud)
	GenerateApps(numApp, suffix, taskProportion)
}

func getCloudFilePaths(numCloud int) string {
	var cloudPath string
	if runtime.GOOS == "windows" {
		cloudPath = fmt.Sprintf("%s\\src\\gogeneticwrsp\\experimenttools\\cloud_%d.json", build.Default.GOPATH, numCloud)
	} else {
		cloudPath = fmt.Sprintf("%s/src/gogeneticwrsp/experimenttools/cloud_%d.json", build.Default.GOPATH, numCloud)
	}
	return cloudPath
}

func getAppFilePaths(numApp int, suffix string) string {
	var appPath string
	if runtime.GOOS == "windows" {
		appPath = fmt.Sprintf("%s\\src\\gogeneticwrsp\\experimenttools\\app_%d_%s.json", build.Default.GOPATH, numApp, suffix)
	} else {
		appPath = fmt.Sprintf("%s/src/gogeneticwrsp/experimenttools/app_%d_%s.json", build.Default.GOPATH, numApp, suffix)
	}
	return appPath
}

// GenerateClouds generates clouds and writes them into files.
func GenerateClouds(numCloud int) {
	log.Printf("generate %d clouds and write them into files\n", numCloud)

	var clouds []model.Cloud = make([]model.Cloud, numCloud)

	// generate clouds
	//cloudDiffTimes := 2.0 // give clouds different types
	for i := 0; i < numCloud; i++ {
		clouds[i].Capacity.CPU = chooseResCPU()
		clouds[i].Capacity.Memory = chooseResMem()
		clouds[i].Capacity.Storage = chooseResStor()

		//// give clouds different types
		//if i > numCloud/4 && i <= numCloud/2 {
		//	clouds[i].Capacity.Memory /= cloudDiffTimes
		//}
		//if i > numCloud/2 && i <= int(float64(numCloud)*0.75) {
		//	clouds[i].Capacity.Storage /= cloudDiffTimes
		//}

		// network conditions
		clouds[i].Capacity.NetCondClouds = make([]model.NetworkCondition, numCloud)
		for j := 0; j < numCloud; j++ {
			if i == j {
				// every cloud has infinite bandwidth and zero RTT between itself, so for some apps with very high requirements, they can be deployed on the same cloud.
				clouds[i].Capacity.NetCondClouds[j].RTT = 0
				clouds[i].Capacity.NetCondClouds[j].DownBw = math.MaxFloat64
			} else {
				clouds[i].Capacity.NetCondClouds[j].RTT = generateResourceRTT()
				clouds[i].Capacity.NetCondClouds[j].DownBw = generateResourceBW()
			}
		}
		clouds[i].Capacity.NetCondImage.RTT = generateResourceRTT()
		clouds[i].Capacity.NetCondImage.DownBw = generateResourceBW()
		clouds[i].Capacity.NetCondController.RTT = generateResourceRTT()
		clouds[i].Capacity.NetCondController.DownBw = generateResourceBW()
		clouds[i].Capacity.UpBwImage = generateResourceBW()
		clouds[i].Capacity.UpBwController = generateResourceBW()

		clouds[i].Allocatable = model.ResCopy(clouds[i].Capacity)
		clouds[i].TmpAlloc = model.ResCopy(clouds[i].Capacity)

		clouds[i].RunningApps = []model.Application{}
		clouds[i].UpdateTime = time.Now()
	}

	cloudsJson, err := json.Marshal(clouds)
	if err != nil {
		log.Fatalln("json.Marshal(clouds) error:", err.Error())
	}

	var cloudPath string = getCloudFilePaths(numCloud)

	err = ioutil.WriteFile(cloudPath, cloudsJson, 0777)
	if err != nil {
		log.Fatalln("ioutil.WriteFile(cloudPath, cloudsJson, 0777) error:", err.Error())
	}
}

// GenerateApps generates apps and writes them into files.
func GenerateApps(numApp int, suffix string, taskProportion float64) {

	log.Printf("generate %d applications and write them into files\n", numApp)

	var apps []model.Application = make([]model.Application, numApp)

	// generate applications
	//var taskProportion float64 = random.RandomFloat64(0.1, 0.9)
	var taskNum int = int(float64(numApp) * taskProportion)
	var svcNum int = numApp - taskNum
	log.Println("taskProportion", taskProportion)
	log.Println("taskNum", taskNum)
	log.Println("svcNum", svcNum)

	//appDiffTimes := 2.0 // give clouds different types
	var currentTaskNum, currentSvcNum int = 0, 0
	for i := 0; i < numApp; i++ {
		var isTask bool
		if currentTaskNum >= taskNum { // if tasks are enough, only generate services
			isTask = false
		} else if currentSvcNum >= svcNum { // if services are enough, only generate tasks
			isTask = true
		} else if random.RandomFloat64(0, 1) < taskProportion { // both tasks and services are not enough, generate them following the proportion
			isTask = true
		} else {
			isTask = false
		}

		if isTask {
			currentTaskNum++
			apps[i].IsTask = true
			apps[i].TaskReq.CPUCycle = generateTaskCPU()

			apps[i].TaskReq.Memory = chooseReqMem()
			apps[i].TaskReq.Storage = chooseReqStor()

			//// give applications different types
			//if i > numApp/4 && i <= numApp/2 {
			//	apps[i].TaskReq.Memory *= appDiffTimes
			//}
			//if i > numApp/2 && i <= int(float64(numApp)*0.75) {
			//	apps[i].TaskReq.Storage *= appDiffTimes
			//}

		} else {
			currentSvcNum++
			apps[i].IsTask = false
			apps[i].SvcReq.CPUClock = generateSvcCPU()
			apps[i].SvcReq.Memory = chooseReqMem()
			apps[i].SvcReq.Storage = chooseReqStor()

			//// give applications different types
			//if i > numApp/4 && i <= numApp/2 {
			//	apps[i].SvcReq.Memory *= appDiffTimes
			//}
			//if i > numApp/2 && i <= int(float64(numApp)*0.75) {
			//	apps[i].SvcReq.Storage *= appDiffTimes
			//}
		}

		apps[i].Priority = generatePriority(100, 65535.9, 150, 300)
		apps[i].InputDataSize = generateInputSize()
		apps[i].ImageSize = generateImageSize()
		apps[i].AppIdx = i
		apps[i].IsNew = true
		apps[i].SvcSuspensionTime = 0
	}

	// generate dependence
	var orderedApps []model.Application = model.AppsCopy(apps)
	sort.Sort(model.AppSlice(orderedApps))
	for i := 0; i < numApp; i++ {
		// randomly choose whether this app depends on others
		// according to the related work, in 14 apps there are 8 depending on others
		if random.RandomInt(1, 14) > 8 {
			continue
		}

		// An app can only depend on apps with higher priorities
		var CurOrderedIdx int
		for j := 0; j < len(orderedApps); j++ {
			// find current app in the ordered apps, and the apps before it can be dependent
			if orderedApps[j].AppIdx == i {
				CurOrderedIdx = j
				break
			}
		}

		// make sure that the priorities in orderedApps before CurOrderedIdx are higher than CurOrderedIdx's priority (cannot be equal)
		for CurOrderedIdx-1 >= 0 && orderedApps[CurOrderedIdx].Priority == orderedApps[CurOrderedIdx-1].Priority {
			CurOrderedIdx--
		}

		if CurOrderedIdx == 0 {
			continue // current app has the highest priority, so no dependence
		}
		// In orderedApps, the idx [0,CurOrderedIdx-1] can be dependent, because they have higher priorities than the current one

		// how many dependent apps
		depNum := chooseDepNum()
		if depNum > CurOrderedIdx {
			depNum = CurOrderedIdx
		}

		// picked depNum apps from orderedApps[:CurOrderedIdx]
		var tmpForPick []int = make([]int, CurOrderedIdx)
		var pickedOrderedIdxes []int = random.RandomPickN(tmpForPick, depNum)
		var depIdxes []int = make([]int, depNum) // These are dependent indexes
		for j := 0; j < depNum; j++ {
			depIdxes[j] = orderedApps[pickedOrderedIdxes[j]].AppIdx
		}

		// generate every dependence for the current app
		for j := 0; j < len(depIdxes); j++ {
			var thisDependence model.Dependence
			// only when the dependent app is a service, the dependence require network resources
			if !apps[depIdxes[j]].IsTask {
				thisDependence = model.Dependence{
					AppIdx: depIdxes[j],
					DownBw: chooseReqBW(),
					UpBw:   chooseReqBW(),
					RTT:    chooseReqRTT(),
				}
			} else {
				thisDependence = model.Dependence{
					AppIdx: depIdxes[j],
					DownBw: 0,
					UpBw:   0,
					RTT:    math.MaxFloat64,
				}
			}
			apps[i].Depend = append(apps[i].Depend, thisDependence)
		}

	}

	appsJson, err := json.Marshal(apps)
	if err != nil {
		log.Fatalln("json.Marshal(apps) error:", err.Error())
	}

	var appPath string = getAppFilePaths(numApp, suffix)

	err = ioutil.WriteFile(appPath, appsJson, 0777)
	if err != nil {
		log.Fatalln("ioutil.WriteFile(appPath, appsJson, 0777) error:", err.Error())
	}
}

// SetOriIdx set the OriIdx for a two-dimension app slice
// In continuous experiments, if an app is remaining, its completion time may change, so I need the original index to update it;
func SetOriIdx(appGroups [][]model.Application) {
	var oriIdx int = 0
	for i := 0; i < len(appGroups); i++ {
		for j := 0; j < len(appGroups[i]); j++ {
			appGroups[i][j].OriIdx = oriIdx
			oriIdx++
		}
	}
}

// SetGeneratedTime set the GeneratedTime for a two-dimension app slice
func SetGeneratedTime(appGroups [][]model.Application, appArrivalTimeIntervals []time.Duration) {
	var thisTime time.Duration = 0 * time.Second
	for i := 0; i < len(appGroups); i++ {
		thisTime += appArrivalTimeIntervals[i]
		for j := 0; j < len(appGroups[i]); j++ {
			appGroups[i][j].GeneratedTime = float64(thisTime) / float64(time.Second)
		}
	}
}

// ReadCloudsApps from files
func ReadCloudsApps(numCloud, numApp int, suffix string) ([]model.Cloud, []model.Application) {
	return ReadClouds(numCloud), ReadApps(numApp, suffix)
}

// ReadClouds from files
func ReadClouds(numCloud int) []model.Cloud {
	var cloudPath string = getCloudFilePaths(numCloud)
	var clouds []model.Cloud

	cloudsJson, err := ioutil.ReadFile(cloudPath)
	if err != nil {
		log.Fatalln("ioutil.ReadFile(cloudPath) error:", err.Error())
	}

	err = json.Unmarshal(cloudsJson, &clouds)
	if err != nil {
		log.Fatalln("json.Unmarshal(cloudsJson, &clouds) error:", err.Error())
	}

	return clouds
}

// ReadApps from files
func ReadApps(numApp int, suffix string) []model.Application {
	var appPath string = getAppFilePaths(numApp, suffix)
	var apps []model.Application

	appsJson, err := ioutil.ReadFile(appPath)
	if err != nil {
		log.Fatalln("ioutil.ReadFile(appPath) error:", err.Error())
	}

	err = json.Unmarshal(appsJson, &apps)
	if err != nil {
		log.Fatalln("json.Unmarshal(appsJson, &apps) error:", err.Error())
	}

	return apps
}

type NumTimeGroup struct {
	NumInGroup    []int           `json:"numInGroup"`
	TimeIntervals []time.Duration `json:"timeIntervals"`
}

// GenerateNumTimeGroup generates the number of apps in app groups and time intervals between groups
func GenerateNumTimeGroup(groupNum int) {
	var numTime NumTimeGroup = NumTimeGroup{
		NumInGroup:    make([]int, groupNum),
		TimeIntervals: make([]time.Duration, groupNum),
	}

	for i := 0; i < groupNum; i++ {
		numTime.NumInGroup[i] = genAppNumGroup()
		if i == 0 {
			numTime.TimeIntervals[i] = 0 * time.Second
		} else {
			// The generation time of the first group should be 0, which is the start of the experiment;
			numTime.TimeIntervals[i] = time.Duration(genTimeIntervalGroups()) * time.Second
		}
	}

	numTimeJson, err := json.Marshal(numTime)
	if err != nil {
		log.Fatalln("numTimeJson, err := json.Marshal(numTime) error:", err.Error())
	}

	var numTimePath string = getNumTimeFilePaths(groupNum)

	err = ioutil.WriteFile(numTimePath, numTimeJson, 0777)
	if err != nil {
		log.Fatalln("ioutil.WriteFile(numTimePath, numTimeJson, 0777) error:", err.Error())
	}

}

func getNumTimeFilePaths(groupNum int) string {
	var numTimePath string
	if runtime.GOOS == "windows" {
		numTimePath = fmt.Sprintf("%s\\src\\gogeneticwrsp\\experimenttools\\numtime_%d.json", build.Default.GOPATH, groupNum)
	} else {
		numTimePath = fmt.Sprintf("%s/src/gogeneticwrsp/experimenttools/numtime_%d.json", build.Default.GOPATH, groupNum)
	}
	return numTimePath
}

// ReadNumTimeGroup from the file
func ReadNumTimeGroup(groupNum int) NumTimeGroup {
	var numTimePath string = getNumTimeFilePaths(groupNum)
	var numTime NumTimeGroup

	numTimeJson, err := ioutil.ReadFile(numTimePath)
	if err != nil {
		log.Fatalln("ioutil.ReadFile(numTimePath) error:", err.Error())
	}

	err = json.Unmarshal(numTimeJson, &numTime)
	if err != nil {
		log.Fatalln("json.Unmarshal(numTimeJson, &numTime) error:", err.Error())
	}

	return numTime
}

type OneTimeHelper struct {
	Name                string
	ExperimentAlgorithm algorithms.SchedulingAlgorithm
	ExperimentSolution  model.Solution
}

// OneTimeExperiment is that all applications are deployed in one time and handled together
func OneTimeExperiment(clouds []model.Cloud, apps []model.Application) {
	gaRandomInit := algorithms.NewGenetic(100, 5000, 0.7, 0.007, 200, algorithms.RandomFitSchedule, clouds, apps)
	gaUndeployedInit := algorithms.NewGenetic(100, 5000, 0.7, 0.007, 200, algorithms.InitializeUndeployedChromosome, clouds, apps)
	//gaRandomInit := algorithms.NewGenetic(1000, 5000, 0.7, 0.007, 5000, algorithms.RandomFitSchedule, clouds, apps)
	//gaUndeployedInit := algorithms.NewGenetic(1000, 5000, 0.7, 0.007, 5000, algorithms.InitializeUndeployedChromosome, clouds, apps)
	ff := algorithms.NewFirstFit(clouds, apps)
	rf := algorithms.NewRandomFit(clouds, apps)

	var experimenters []OneTimeHelper = []OneTimeHelper{
		{
			Name:                "Genetic Random Init",
			ExperimentAlgorithm: gaRandomInit,
		},
		{
			Name:                "Genetic Undeployed Init",
			ExperimentAlgorithm: gaUndeployedInit,
		},
		{
			Name:                "First Fit",
			ExperimentAlgorithm: ff,
		},
		{
			Name:                "Random Fit",
			ExperimentAlgorithm: rf,
		},
	}

	for i := 0; i < len(experimenters); i++ {
		solution, err := experimenters[i].ExperimentAlgorithm.Schedule(clouds, apps)
		if err != nil {
			log.Printf("Error in [%s], %s", experimenters[i].Name, err.Error())
		}
		experimenters[i].ExperimentSolution = solution
	}

	for _, experimenter := range experimenters {
		log.Println(experimenter.Name, "CPUClock Idle Rate:", algorithms.CPUIdleRate(clouds, apps, experimenter.ExperimentSolution.SchedulingResult), "Memory Idle Rate:", algorithms.MemoryIdleRate(clouds, apps, experimenter.ExperimentSolution.SchedulingResult), "Storage Idle Rate:", algorithms.StorageIdleRate(clouds, apps, experimenter.ExperimentSolution.SchedulingResult), "Bandwidth Idle Rate:", algorithms.BwIdleRate(clouds, apps, experimenter.ExperimentSolution.SchedulingResult), "Total Accepted Priority:", algorithms.AcceptedPriority(clouds, apps, experimenter.ExperimentSolution.SchedulingResult))
		if v, ok := experimenter.ExperimentAlgorithm.(*algorithms.Genetic); ok {
			log.Println("Got Iteration:", v.BestUntilNowUpdateIterations[len(v.BestUntilNowUpdateIterations)-1], v.BestAcceptableUntilNowUpdateIterations[len(v.BestAcceptableUntilNowUpdateIterations)-1])
		}
	}
}

type ContinuousHelper struct {
	Name               string
	CPUIdleRecords     []float64
	MemoryIdleRecords  []float64
	StorageIdleRecords []float64
	BwIdleRecords      []float64

	AcceptedPriorityRateRecords []float64
	AcceptedSvcPriRateRecords   []float64
	AcceptedTaskPriRateRecords  []float64

	CloudsWithTime        [][]float64
	AllAppComplTime       []float64
	AllAppComplTimePerPri []float64

	SvcSusTime       []float64 // record the weighted suspension time of every service
	TaskComplTime    []float64 // record the weighted time since generated until completed of every task
	maxSvcSusTime    float64
	maxTaskComplTime float64
}

// we use the priority of each app as the weight of each time value
func (ch *ContinuousHelper) setSvcSusTaskComplTime(clouds []model.Cloud, totalApps []model.Application, currentSolution model.Solution) {
	ch.maxTaskComplTime, ch.maxSvcSusTime = 0, 0
	for i := 0; i < len(totalApps); i++ {
		if totalApps[i].IsTask {
			if currentSolution.SchedulingResult[i] != len(clouds) {
				ch.TaskComplTime = append(ch.TaskComplTime, (totalApps[i].TaskFinalComplTime-totalApps[i].GeneratedTime)*float64(totalApps[i].Priority)*0.00001) // priority*0.00001 is the weight
				if ch.TaskComplTime[len(ch.TaskComplTime)-1] > ch.maxTaskComplTime {
					ch.maxTaskComplTime = ch.TaskComplTime[len(ch.TaskComplTime)-1]
				}
			} else { // if rejected set it as -1, calculating later
				ch.TaskComplTime = append(ch.TaskComplTime, -1)
			}
		} else {
			if currentSolution.SchedulingResult[i] != len(clouds) {
				ch.SvcSusTime = append(ch.SvcSusTime, totalApps[i].SvcSuspensionTime*float64(totalApps[i].Priority)*0.00001) // priority*0.00001 is the weight
				if ch.SvcSusTime[len(ch.SvcSusTime)-1] > ch.maxSvcSusTime {
					ch.maxSvcSusTime = ch.SvcSusTime[len(ch.SvcSusTime)-1]
				}
			} else { // if rejected set it as -1, calculating later
				ch.SvcSusTime = append(ch.SvcSusTime, -1)
			}
		}
	}
}

// set the service suspension time and task completion time of rejected apps as very high
func setRejectedSvcTask(recorders ...*ContinuousHelper) {
	var totalMaxSvcSus, totalMaxTaskCompl float64 = 0, 0
	for _, recorder := range recorders {
		if (*recorder).maxSvcSusTime > totalMaxSvcSus {
			totalMaxSvcSus = (*recorder).maxSvcSusTime
		}
		if (*recorder).maxTaskComplTime > totalMaxTaskCompl {
			totalMaxTaskCompl = (*recorder).maxTaskComplTime
		}
	}

	for _, recorder := range recorders {
		for i := 0; i < len((*recorder).SvcSusTime); i++ {
			if (*recorder).SvcSusTime[i] == -1 {
				(*recorder).SvcSusTime[i] = totalMaxSvcSus * 1.1
			}
		}
		for i := 0; i < len((*recorder).TaskComplTime); i++ {
			if (*recorder).TaskComplTime[i] == -1 {
				(*recorder).TaskComplTime[i] = totalMaxTaskCompl * 1.1
			}
		}
	}
}

// ContinuousExperiment is that the applications are deployed one by one. In one time, we only handle one application.
func ContinuousExperiment(clouds []model.Cloud, apps [][]model.Application, appArrivalTimeIntervals []time.Duration) {
	// if an app is remaining, its completion time may change, so I need the original index to update it;
	SetOriIdx(apps)
	SetGeneratedTime(apps, appArrivalTimeIntervals)
	//fmt.Println()
	//for i := 0; i < len(apps); i++ {
	//	for j := 0; j < len(apps[i]); j++ {
	//		fmt.Printf("%0.2f ", apps[i][j].GeneratedTime)
	//	}
	//	fmt.Println()
	//}
	//time.Sleep(time.Second * 1000)

	var currentClouds []model.Cloud
	var totalApps []model.Application
	var currentSolution model.Solution

	var totalSolution model.Solution

	var lastApps []model.Application // apps in last round
	var lastSolution model.Solution  // solution of last round

	var currentTime time.Duration

	var firstFitRecorder ContinuousHelper = ContinuousHelper{
		Name:                        "First Fit",
		CPUIdleRecords:              make([]float64, 0),
		MemoryIdleRecords:           make([]float64, 0),
		StorageIdleRecords:          make([]float64, 0),
		BwIdleRecords:               make([]float64, 0),
		AcceptedPriorityRateRecords: make([]float64, 0),
		AcceptedSvcPriRateRecords:   make([]float64, 0),
		AcceptedTaskPriRateRecords:  make([]float64, 0),
		CloudsWithTime:              make([][]float64, 0),
		AllAppComplTime:             make([]float64, 0),
		AllAppComplTimePerPri:       make([]float64, 0),
		SvcSusTime:                  make([]float64, 0),
		TaskComplTime:               make([]float64, 0),
	}
	var randomFitRecorder ContinuousHelper = ContinuousHelper{
		Name:                        "Random Fit",
		CPUIdleRecords:              make([]float64, 0),
		MemoryIdleRecords:           make([]float64, 0),
		StorageIdleRecords:          make([]float64, 0),
		BwIdleRecords:               make([]float64, 0),
		AcceptedPriorityRateRecords: make([]float64, 0),
		AcceptedSvcPriRateRecords:   make([]float64, 0),
		AcceptedTaskPriRateRecords:  make([]float64, 0),
		CloudsWithTime:              make([][]float64, 0),
		AllAppComplTime:             make([]float64, 0),
		AllAppComplTimePerPri:       make([]float64, 0),
		SvcSusTime:                  make([]float64, 0),
		TaskComplTime:               make([]float64, 0),
	}
	// Multi-cloud Applications Scheduling Genetic Algorithm (MCASGA)
	var MCASGARecorder ContinuousHelper = ContinuousHelper{
		Name:                        "MCASGA",
		CPUIdleRecords:              make([]float64, 0),
		MemoryIdleRecords:           make([]float64, 0),
		StorageIdleRecords:          make([]float64, 0),
		BwIdleRecords:               make([]float64, 0),
		AcceptedPriorityRateRecords: make([]float64, 0),
		AcceptedSvcPriRateRecords:   make([]float64, 0),
		AcceptedTaskPriRateRecords:  make([]float64, 0),
		CloudsWithTime:              make([][]float64, 0),
		AllAppComplTime:             make([]float64, 0),
		AllAppComplTimePerPri:       make([]float64, 0),
		SvcSusTime:                  make([]float64, 0),
		TaskComplTime:               make([]float64, 0),
	}

	// First Fit
	currentClouds = model.CloudsCopy(clouds)
	totalApps = []model.Application{}
	currentSolution = model.Solution{}

	currentTime = 0 * time.Second

	for i := 0; i < len(apps); i++ {
		currentTime += appArrivalTimeIntervals[i]
		log.Println("group", i, "currentTime", float64(currentTime)/float64(time.Second))

		// the ith applications group request comes
		totalApps = model.CombApps(totalApps, apps[i])
		thisAppGroup := apps[i]

		ff := algorithms.NewFirstFit(currentClouds, thisAppGroup)
		solution, err := ff.Schedule(currentClouds, thisAppGroup)
		if err != nil {
			log.Printf("Error, app %d. Error message: %s", i, err.Error())
		}

		// get apps with all time related attributes
		tmpClouds4Time := model.CloudsCopy(currentClouds)
		tmpAppsToDeploy4Time := model.AppsCopy(thisAppGroup)
		tmpSolution4Time := model.SolutionCopy(solution)
		tmpClouds4Time = algorithms.SimulateDeploy(tmpClouds4Time, tmpAppsToDeploy4Time, tmpSolution4Time)

		timeApps := algorithms.CalcStartComplTime(tmpClouds4Time, tmpAppsToDeploy4Time, tmpSolution4Time.SchedulingResult)

		for j := 0; j < len(timeApps); j++ {
			if tmpSolution4Time.SchedulingResult[j] == len(tmpClouds4Time) { // only record the time of accepted apps
				continue
			}
			if timeApps[j].IsTask { // set final completion time for tasks
				totalApps[timeApps[j].OriIdx].TaskFinalComplTime = timeApps[j].TaskCompletionTime + float64(currentTime)/float64(time.Second)
			} else { // set suspension time for services
				totalApps[timeApps[j].OriIdx].SvcSuspensionTime += timeApps[j].StartTime
			}
		}

		// if last group does not finish when this group generate, this group need to wait for last one
		if i != 0 {
			for j := 0; j < len(tmpClouds4Time); j++ {
				if wait := firstFitRecorder.CloudsWithTime[i-1][j] - float64(currentTime)/float64(time.Second); wait > 0 {
					tmpClouds4Time[j].TotalTaskComplTime += wait
					for k := 0; k < len(tmpClouds4Time[j].RunningApps); k++ {
						if totalApps[tmpClouds4Time[j].RunningApps[k].OriIdx].IsTask {
							totalApps[tmpClouds4Time[j].RunningApps[k].OriIdx].TaskFinalComplTime += wait
						} else {
							totalApps[tmpClouds4Time[j].RunningApps[k].OriIdx].SvcSuspensionTime += wait
						}
					}
				}
			}
		}

		// record TotalTaskComplTime of clouds
		var thisTimeRecord []float64 = make([]float64, len(tmpClouds4Time))
		var longestTime float64 = 0
		for j := 0; j < len(tmpClouds4Time); j++ {
			thisTimeRecord[j] = tmpClouds4Time[j].TotalTaskComplTime + float64(currentTime)/float64(time.Second)
			if thisTimeRecord[j] > longestTime {
				longestTime = thisTimeRecord[j]
			}
		}
		firstFitRecorder.CloudsWithTime = append(firstFitRecorder.CloudsWithTime, thisTimeRecord)
		firstFitRecorder.AllAppComplTime = append(firstFitRecorder.AllAppComplTime, longestTime)
		log.Println("thisTimeRecord", thisTimeRecord)
		log.Println("firstFitRecorder.AllAppComplTime", firstFitRecorder.AllAppComplTime)

		// add the solution of this app to current solution
		currentSolution.SchedulingResult = append(currentSolution.SchedulingResult, solution.SchedulingResult...)

		// deploy this app in current clouds (subtract the resources)
		//currentClouds = algorithms.TrulyDeploy(clouds, totalApps, currentSolution)
		currentClouds = algorithms.TrulyDeploy(currentClouds, thisAppGroup, solution)

		// RunningApps should be empty before the scheduling of the next round
		for j := 0; j < len(currentClouds); j++ {
			currentClouds[j].RunningApps = []model.Application{}
		}

		// evaluate current solution, current cloud, current apps
		firstFitRecorder.CPUIdleRecords = append(firstFitRecorder.CPUIdleRecords, algorithms.CPUIdleRate(clouds, totalApps, currentSolution.SchedulingResult))
		firstFitRecorder.MemoryIdleRecords = append(firstFitRecorder.MemoryIdleRecords, algorithms.MemoryIdleRate(clouds, totalApps, currentSolution.SchedulingResult))
		firstFitRecorder.StorageIdleRecords = append(firstFitRecorder.StorageIdleRecords, algorithms.StorageIdleRate(clouds, totalApps, currentSolution.SchedulingResult))
		firstFitRecorder.BwIdleRecords = append(firstFitRecorder.BwIdleRecords, algorithms.BwIdleRate(clouds, totalApps, currentSolution.SchedulingResult))

		firstFitRecorder.AcceptedPriorityRateRecords = append(firstFitRecorder.AcceptedPriorityRateRecords, float64(algorithms.AcceptedPriority(clouds, totalApps, currentSolution.SchedulingResult))/float64(algorithms.TotalPriority(clouds, totalApps, currentSolution.SchedulingResult)))
		firstFitRecorder.AcceptedSvcPriRateRecords = append(firstFitRecorder.AcceptedSvcPriRateRecords, algorithms.AcceptedSvcPriRate(clouds, totalApps, currentSolution.SchedulingResult))
		firstFitRecorder.AcceptedTaskPriRateRecords = append(firstFitRecorder.AcceptedTaskPriRateRecords, algorithms.AcceptedTaskPriRate(clouds, totalApps, currentSolution.SchedulingResult))

		firstFitRecorder.AllAppComplTimePerPri = append(firstFitRecorder.AllAppComplTimePerPri, longestTime/float64(algorithms.AcceptedPriority(clouds, totalApps, currentSolution.SchedulingResult)))
	}
	// For firstFit, record the service suspension time and task completion time
	firstFitRecorder.setSvcSusTaskComplTime(clouds, totalApps, currentSolution)

	// Random Fit
	currentClouds = model.CloudsCopy(clouds)
	totalApps = []model.Application{}
	currentSolution = model.Solution{}

	currentTime = 0 * time.Second

	for i := 0; i < len(apps); i++ {
		currentTime += appArrivalTimeIntervals[i]
		log.Println("group", i, "currentTime", float64(currentTime)/float64(time.Second))

		// the ith applications group request comes
		totalApps = model.CombApps(totalApps, apps[i])
		thisAppGroup := apps[i]

		rf := algorithms.NewRandomFit(currentClouds, thisAppGroup)
		solution, err := rf.Schedule(currentClouds, thisAppGroup)
		if err != nil {
			log.Printf("Error, app %d. Error message: %s", i, err.Error())
		}

		// get apps with all time related attributes
		tmpClouds4Time := model.CloudsCopy(currentClouds)
		tmpAppsToDeploy4Time := model.AppsCopy(thisAppGroup)
		tmpSolution4Time := model.SolutionCopy(solution)
		tmpClouds4Time = algorithms.SimulateDeploy(tmpClouds4Time, tmpAppsToDeploy4Time, tmpSolution4Time)
		timeApps := algorithms.CalcStartComplTime(tmpClouds4Time, tmpAppsToDeploy4Time, tmpSolution4Time.SchedulingResult)

		for j := 0; j < len(timeApps); j++ {
			if tmpSolution4Time.SchedulingResult[j] == len(tmpClouds4Time) { // only record the time of accepted apps
				continue
			}
			if timeApps[j].IsTask { // set final completion time for tasks
				totalApps[timeApps[j].OriIdx].TaskFinalComplTime = timeApps[j].TaskCompletionTime + float64(currentTime)/float64(time.Second)
			} else { // set suspension time for services
				totalApps[timeApps[j].OriIdx].SvcSuspensionTime += timeApps[j].StartTime
			}
		}

		// if last group does not finish when this group generate, this group need to wait for last one
		if i != 0 {
			for j := 0; j < len(tmpClouds4Time); j++ {
				if wait := randomFitRecorder.CloudsWithTime[i-1][j] - float64(currentTime)/float64(time.Second); wait > 0 {
					tmpClouds4Time[j].TotalTaskComplTime += wait
					for k := 0; k < len(tmpClouds4Time[j].RunningApps); k++ {
						if totalApps[tmpClouds4Time[j].RunningApps[k].OriIdx].IsTask {
							totalApps[tmpClouds4Time[j].RunningApps[k].OriIdx].TaskFinalComplTime += wait
						} else {
							totalApps[tmpClouds4Time[j].RunningApps[k].OriIdx].SvcSuspensionTime += wait
						}
					}
				}
			}
		}

		// record TotalTaskComplTime of clouds
		var thisTimeRecord []float64 = make([]float64, len(tmpClouds4Time))
		var longestTime float64 = 0
		for j := 0; j < len(tmpClouds4Time); j++ {
			thisTimeRecord[j] = tmpClouds4Time[j].TotalTaskComplTime + float64(currentTime)/float64(time.Second)
			if thisTimeRecord[j] > longestTime {
				longestTime = thisTimeRecord[j]
			}
		}
		randomFitRecorder.CloudsWithTime = append(randomFitRecorder.CloudsWithTime, thisTimeRecord)
		randomFitRecorder.AllAppComplTime = append(randomFitRecorder.AllAppComplTime, longestTime)
		log.Println("thisTimeRecord", thisTimeRecord)
		log.Println("randomFitRecorder.AllAppComplTime", randomFitRecorder.AllAppComplTime)

		// add the solution of this app to current solution
		currentSolution.SchedulingResult = append(currentSolution.SchedulingResult, solution.SchedulingResult...)

		// deploy this app in current clouds (subtract the resources)
		//currentClouds = algorithms.TrulyDeploy(clouds, totalApps, currentSolution)
		currentClouds = algorithms.TrulyDeploy(currentClouds, thisAppGroup, solution)

		// RunningApps should be empty before the scheduling of the next round
		for j := 0; j < len(currentClouds); j++ {
			currentClouds[j].RunningApps = []model.Application{}
		}

		// evaluate current solution, current cloud, current apps
		randomFitRecorder.CPUIdleRecords = append(randomFitRecorder.CPUIdleRecords, algorithms.CPUIdleRate(clouds, totalApps, currentSolution.SchedulingResult))
		randomFitRecorder.MemoryIdleRecords = append(randomFitRecorder.MemoryIdleRecords, algorithms.MemoryIdleRate(clouds, totalApps, currentSolution.SchedulingResult))
		randomFitRecorder.StorageIdleRecords = append(randomFitRecorder.StorageIdleRecords, algorithms.StorageIdleRate(clouds, totalApps, currentSolution.SchedulingResult))
		randomFitRecorder.BwIdleRecords = append(randomFitRecorder.BwIdleRecords, algorithms.BwIdleRate(clouds, totalApps, currentSolution.SchedulingResult))

		randomFitRecorder.AcceptedPriorityRateRecords = append(randomFitRecorder.AcceptedPriorityRateRecords, float64(algorithms.AcceptedPriority(clouds, totalApps, currentSolution.SchedulingResult))/float64(algorithms.TotalPriority(clouds, totalApps, currentSolution.SchedulingResult)))
		randomFitRecorder.AcceptedSvcPriRateRecords = append(randomFitRecorder.AcceptedSvcPriRateRecords, algorithms.AcceptedSvcPriRate(clouds, totalApps, currentSolution.SchedulingResult))
		randomFitRecorder.AcceptedTaskPriRateRecords = append(randomFitRecorder.AcceptedTaskPriRateRecords, algorithms.AcceptedTaskPriRate(clouds, totalApps, currentSolution.SchedulingResult))

		randomFitRecorder.AllAppComplTimePerPri = append(randomFitRecorder.AllAppComplTimePerPri, longestTime/float64(algorithms.AcceptedPriority(clouds, totalApps, currentSolution.SchedulingResult)))
	}
	// For randomFit, record the service suspension time and task completion time
	randomFitRecorder.setSvcSusTaskComplTime(clouds, totalApps, currentSolution)

	// MCASGA
	currentClouds = model.CloudsCopy(clouds)
	totalApps = []model.Application{}
	currentSolution = model.Solution{}

	totalSolution = model.Solution{}

	lastApps = []model.Application{}
	lastSolution = model.Solution{}

	currentTime = 0 * time.Second

	for i := 0; i < len(apps); i++ {
		currentTime += appArrivalTimeIntervals[i]
		log.Println("group", i, "currentTime", float64(currentTime)/float64(time.Second))

		var appsToDeploy []model.Application

		// the ith applications group request comes
		appsToDeploy = model.CombApps(appsToDeploy, apps[i]) // this group of apps + remaining apps

		if len(lastApps) != 0 { // not the first group
			tmpClouds := model.CloudsCopy(clouds)
			tmpApps := model.AppsCopy(lastApps)
			tmpSolution := model.SolutionCopy(lastSolution)

			timeClouds := algorithms.SimulateDeploy(tmpClouds, tmpApps, tmpSolution)
			algorithms.CalcStartComplTime(timeClouds, tmpApps, tmpSolution.SchedulingResult)

			//for j := 0; j < len(timeClouds); j++ {
			//	fmt.Println(i, timeClouds[j].TotalTaskComplTime, len(timeClouds[j].RunningApps))
			//}

			resClouds := model.CloudsCopy(currentClouds)

			var timeIntervalSec float64 = float64(appArrivalTimeIntervals[i]) / float64(time.Second) // unit second

			// here:
			// resClouds have the information: 1. resource usage; 2. update time of clouds in last round 3. deployed apps
			// timeClouds have the information: 1. execution time; 2. deployed apps
			// I need to use them to calculate: at this time (this round), what applications are still running on each cloud, and how many cycles of them still need to be executed

			remainingApps := algorithms.CalcRemainingApps(resClouds, timeClouds, timeIntervalSec)
			appsToDeploy = model.CombApps(appsToDeploy, remainingApps) // this group of apps + remaining apps

		}

		totalApps = model.CombApps(totalApps, apps[i])
		lastApps = model.AppsCopy(appsToDeploy)

		//ga := algorithms.NewGenetic(100, 5000, 0.7, 0.007, 2000, algorithms.InitializeUndeployedChromosome, clouds, totalApps)
		ga := algorithms.NewGenetic(200, 5000, 0.3, 0.004, 200, algorithms.RandomFitSchedule, clouds, appsToDeploy)
		solution, err := ga.Schedule(clouds, appsToDeploy)
		if err != nil {
			log.Printf("Error, app %d. Error message: %s", i, err.Error())
		}

		lastSolution = model.SolutionCopy(solution)

		// get apps with all time related attributes
		tmpClouds4Time := model.CloudsCopy(clouds)
		tmpAppsToDeploy4Time := model.AppsCopy(appsToDeploy)
		tmpSolution4Time := model.SolutionCopy(solution)
		tmpClouds4Time = algorithms.SimulateDeploy(tmpClouds4Time, tmpAppsToDeploy4Time, tmpSolution4Time)
		timeApps := algorithms.CalcStartComplTime(tmpClouds4Time, tmpAppsToDeploy4Time, tmpSolution4Time.SchedulingResult)

		for j := 0; j < len(timeApps); j++ {
			if tmpSolution4Time.SchedulingResult[j] == len(tmpClouds4Time) { // only record the time of accepted apps
				continue
			}
			if timeApps[j].IsTask { // set final completion time for tasks
				totalApps[timeApps[j].OriIdx].TaskFinalComplTime = timeApps[j].TaskCompletionTime + float64(currentTime)/float64(time.Second)
			} else { // set suspension time for services
				totalApps[timeApps[j].OriIdx].SvcSuspensionTime += timeApps[j].StartTime
			}
		}

		// record TotalTaskComplTime of clouds
		var thisTimeRecord []float64 = make([]float64, len(tmpClouds4Time))
		var longestTime float64 = 0
		for j := 0; j < len(tmpClouds4Time); j++ {
			thisTimeRecord[j] = tmpClouds4Time[j].TotalTaskComplTime + float64(currentTime)/float64(time.Second)
			if thisTimeRecord[j] > longestTime {
				longestTime = thisTimeRecord[j]
			}
		}
		MCASGARecorder.CloudsWithTime = append(MCASGARecorder.CloudsWithTime, thisTimeRecord)
		MCASGARecorder.AllAppComplTime = append(MCASGARecorder.AllAppComplTime, longestTime)
		log.Println("thisTimeRecord", thisTimeRecord)
		log.Println("MCASGARecorder.AllAppComplTime", MCASGARecorder.AllAppComplTime)

		lastSolution = model.SolutionCopy(solution)

		// deploy this app in current clouds (subtract the resources)
		currentClouds = algorithms.TrulyDeploy(clouds, appsToDeploy, solution)
		for j := 0; j < len(currentClouds); j++ {
			currentClouds[j].RefreshTime(appArrivalTimeIntervals[i])
		}

		// add the solution of this app to current solution
		currentSolution = model.SolutionCopy(solution)
		tmpSolution := model.SolutionCopy(solution)
		tmpSolution.SchedulingResult = tmpSolution.SchedulingResult[:len(apps[i])]
		totalSolution.SchedulingResult = append(totalSolution.SchedulingResult, tmpSolution.SchedulingResult...)
		//log.Println("i:", i, "------------------")
		//log.Println("currentSolution:", currentSolution)
		//log.Println("len(apps[i]):", len(apps[i]), "tmpSolution:", tmpSolution)
		//log.Println("totalSolution", totalSolution)

		// evaluate current solution, current cloud, current apps
		MCASGARecorder.CPUIdleRecords = append(MCASGARecorder.CPUIdleRecords, algorithms.CPUIdleRate(clouds, appsToDeploy, currentSolution.SchedulingResult))
		MCASGARecorder.MemoryIdleRecords = append(MCASGARecorder.MemoryIdleRecords, algorithms.MemoryIdleRate(clouds, appsToDeploy, currentSolution.SchedulingResult))
		MCASGARecorder.StorageIdleRecords = append(MCASGARecorder.StorageIdleRecords, algorithms.StorageIdleRate(clouds, appsToDeploy, currentSolution.SchedulingResult))
		MCASGARecorder.BwIdleRecords = append(MCASGARecorder.BwIdleRecords, algorithms.BwIdleRate(clouds, appsToDeploy, currentSolution.SchedulingResult))

		MCASGARecorder.AcceptedPriorityRateRecords = append(MCASGARecorder.AcceptedPriorityRateRecords, float64(algorithms.AcceptedPriority(clouds, totalApps, totalSolution.SchedulingResult))/float64(algorithms.TotalPriority(clouds, totalApps, totalSolution.SchedulingResult)))
		MCASGARecorder.AcceptedSvcPriRateRecords = append(MCASGARecorder.AcceptedSvcPriRateRecords, algorithms.AcceptedSvcPriRate(clouds, totalApps, totalSolution.SchedulingResult))
		MCASGARecorder.AcceptedTaskPriRateRecords = append(MCASGARecorder.AcceptedTaskPriRateRecords, algorithms.AcceptedTaskPriRate(clouds, totalApps, totalSolution.SchedulingResult))

		MCASGARecorder.AllAppComplTimePerPri = append(MCASGARecorder.AllAppComplTimePerPri, longestTime/float64(algorithms.AcceptedPriority(clouds, totalApps, totalSolution.SchedulingResult)))
	}
	// For MCASGA, record the service suspension time and task completion time
	MCASGARecorder.setSvcSusTaskComplTime(clouds, totalApps, totalSolution)
	// after the service suspension time and task completion time in 3 recorders are set, we handle the rejected apps
	setRejectedSvcTask(&firstFitRecorder, &randomFitRecorder, &MCASGARecorder)

	log.Println("MCASGA solution:", totalSolution.SchedulingResult)

	currentTime = 0 * time.Second
	// output csv files
	generateCsvFunc := func(recorder ContinuousHelper) [][]string {
		var csvContent [][]string
		csvContent = append(csvContent, []string{"Number of Applications", "Number of New Applications", "Time", "CPUClock Idle Rate", "Memory Idle Rate", "Storage Idle Rate", "Bandwidth Idle Rate", "Application Acceptance Rate", "Service Acceptance Rate", "Task Acceptance Rate", "Completion Time", "Completion Time Per Priority"})
		appNum := 0
		currentTime = 0 * time.Second
		for i := 0; i < len(apps); i++ {
			appNum += len(apps[i])
			currentTime += appArrivalTimeIntervals[i]
			csvContent = append(csvContent, []string{fmt.Sprintf("%d", appNum), fmt.Sprintf("%d", len(apps[i])), fmt.Sprintf("%.0f", float64(currentTime)/float64(time.Second)), fmt.Sprintf("%f", recorder.CPUIdleRecords[i]), fmt.Sprintf("%f", recorder.MemoryIdleRecords[i]), fmt.Sprintf("%f", recorder.StorageIdleRecords[i]), fmt.Sprintf("%f", recorder.BwIdleRecords[i]), fmt.Sprintf("%f", recorder.AcceptedPriorityRateRecords[i]), fmt.Sprintf("%f", recorder.AcceptedSvcPriRateRecords[i]), fmt.Sprintf("%f", recorder.AcceptedTaskPriRateRecords[i]), fmt.Sprintf("%f", recorder.AllAppComplTime[i]), fmt.Sprintf("%f", recorder.AllAppComplTimePerPri[i])})
		}
		return csvContent
	}
	var ffCsvContent [][]string = generateCsvFunc(firstFitRecorder)
	var rfCsvContent [][]string = generateCsvFunc(randomFitRecorder)
	var MCASGACsvContent [][]string = generateCsvFunc(MCASGARecorder)

	csvPathFunc := func(name string) string {
		name = strings.Replace(name, " ", "_", -1)
		var csvpath string
		if runtime.GOOS == "windows" {
			csvpath = fmt.Sprintf("%s\\src\\gogeneticwrsp\\experimenttools\\%s.csv", build.Default.GOPATH, name)
		} else {
			csvpath = fmt.Sprintf("%s/src/gogeneticwrsp/experimenttools/%s.csv", build.Default.GOPATH, name)
		}
		return csvpath
	}

	writeFileFunc := func(fileName string, csvContent [][]string) {
		f, err := os.Create(fileName)
		defer f.Close()
		if err != nil {
			log.Fatalln("Fatal: ", err)
		}
		w := csv.NewWriter(f)
		defer w.Flush()

		for _, record := range csvContent {
			if err := w.Write(record); err != nil {
				log.Fatalf("write record %v, error: %s", record, err.Error())
			}
		}
	}
	writeFileFunc(csvPathFunc(firstFitRecorder.Name), ffCsvContent)
	writeFileFunc(csvPathFunc(randomFitRecorder.Name), rfCsvContent)
	writeFileFunc(csvPathFunc(MCASGARecorder.Name), MCASGACsvContent)

	// write cdf csv file of service suspension time and task completion time
	// output csv files
	genSvcCdfCsvFunc := func() [][]string {
		var csvContent [][]string
		csvContent = append(csvContent, []string{"First Fit Weighted Service Suspension Time", "Random Fit Weighted Service Suspension Time", "MCASGA Weighted Service Suspension Time"})
		for i := 0; i < len(firstFitRecorder.SvcSusTime); i++ {
			csvContent = append(csvContent, []string{fmt.Sprintf("%g", firstFitRecorder.SvcSusTime[i]), fmt.Sprintf("%g", randomFitRecorder.SvcSusTime[i]), fmt.Sprintf("%g", MCASGARecorder.SvcSusTime[i])})
		}
		return csvContent
	}
	genTaskCdfCsvFunc := func() [][]string {
		var csvContent [][]string
		csvContent = append(csvContent, []string{"First Fit Weighted Task Completion Time", "Random Fit Weighted Task Completion Time", "MCASGA Weighted Task Completion Time"})
		for i := 0; i < len(firstFitRecorder.TaskComplTime); i++ {
			csvContent = append(csvContent, []string{fmt.Sprintf("%g", firstFitRecorder.TaskComplTime[i]), fmt.Sprintf("%g", randomFitRecorder.TaskComplTime[i]), fmt.Sprintf("%g", MCASGARecorder.TaskComplTime[i])})
		}
		return csvContent
	}
	writeFileFunc(csvPathFunc("svc_cdf"), genSvcCdfCsvFunc())
	writeFileFunc(csvPathFunc("task_cdf"), genTaskCdfCsvFunc())

	// draw line charts
	var ticks []chart.Tick
	var xValues []float64
	for i := 0; i < len(apps); i++ {
		if i == 0 {
			xValues = append(xValues, float64(appArrivalTimeIntervals[i])/float64(time.Second))
		} else {
			xValues = append(xValues, xValues[i-1]+float64(appArrivalTimeIntervals[i])/float64(time.Second))
		}

		ticks = append(ticks, chart.Tick{
			Value: xValues[i],
			Label: fmt.Sprintf("%.0fs, %d apps", xValues[i], len(apps[i])),
		})
	}

	////// draw line charts
	////var appNumbers []float64
	////var ticks []chart.Tick
	////appNumbers = append(appNumbers, float64(len(apps[0])))
	////ticks = append(ticks, chart.Tick{
	////	Value: appNumbers[0],
	////	Label: fmt.Sprintf("%d", int(appNumbers[0])),
	////})
	////for i := 1; i < len(apps); i++ {
	////	appNumbers = append(appNumbers, appNumbers[i-1]+float64(len(apps[i])))
	////	ticks = append(ticks, chart.Tick{
	////		Value: appNumbers[i],
	////		Label: fmt.Sprintf("%d", int(appNumbers[i])),
	////	})
	////}
	//
	//var CPUChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
	//	graph := chart.Chart{
	//		Title: "CPUClock Idle Rate",
	//		XAxis: chart.XAxis{
	//			Name:      "Time, Number of New Applications",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//			ValueFormatter: func(v interface{}) string {
	//				return strconv.FormatInt(int64(v.(float64)), 10)
	//			},
	//			Ticks: ticks,
	//		},
	//		YAxis: chart.YAxis{
	//			AxisType:  chart.YAxisSecondary,
	//			Name:      "CPUClock Idle Rate",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//		},
	//		Background: chart.Style{
	//			Padding: chart.Box{
	//				Top:  50,
	//				Left: 20,
	//			},
	//		},
	//		Series: []chart.Series{
	//			chart.ContinuousSeries{
	//				Name:    firstFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: firstFitRecorder.CPUIdleRecords,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    randomFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: randomFitRecorder.CPUIdleRecords,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    MCASGARecorder.Name,
	//				XValues: xValues,
	//				YValues: MCASGARecorder.CPUIdleRecords,
	//			},
	//		},
	//	}
	//
	//	graph.Elements = []chart.Renderable{
	//		chart.LegendThin(&graph),
	//	}
	//
	//	res.Header().Set("Content-Type", "image/png")
	//	err := graph.Render(chart.PNG, res)
	//	if err != nil {
	//		log.Println("Error: graph.Render(chart.PNG, res)", err)
	//	}
	//}
	//
	//var memoryChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
	//	graph := chart.Chart{
	//		Title: "Memory Idle Rate",
	//		XAxis: chart.XAxis{
	//			Name:      "Time, Number of New Applications",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//			ValueFormatter: func(v interface{}) string {
	//				return strconv.FormatInt(int64(v.(float64)), 10)
	//			},
	//			Ticks: ticks,
	//		},
	//		YAxis: chart.YAxis{
	//			AxisType:  chart.YAxisSecondary,
	//			Name:      "Memory Idle Rate",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//		},
	//		Background: chart.Style{
	//			Padding: chart.Box{
	//				Top:  50,
	//				Left: 20,
	//			},
	//		},
	//		Series: []chart.Series{
	//			chart.ContinuousSeries{
	//				Name:    firstFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: firstFitRecorder.MemoryIdleRecords,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    randomFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: randomFitRecorder.MemoryIdleRecords,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    MCASGARecorder.Name,
	//				XValues: xValues,
	//				YValues: MCASGARecorder.MemoryIdleRecords,
	//			},
	//		},
	//	}
	//
	//	graph.Elements = []chart.Renderable{
	//		chart.LegendThin(&graph),
	//	}
	//
	//	res.Header().Set("Content-Type", "image/png")
	//	err := graph.Render(chart.PNG, res)
	//	if err != nil {
	//		log.Println("Error: graph.Render(chart.PNG, res)", err)
	//	}
	//}
	//
	//var storageChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
	//	graph := chart.Chart{
	//		Title: "Storage Idle Rate",
	//		XAxis: chart.XAxis{
	//			Name:      "Time, Number of New Applications",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//			ValueFormatter: func(v interface{}) string {
	//				return strconv.FormatInt(int64(v.(float64)), 10)
	//			},
	//			Ticks: ticks,
	//		},
	//		YAxis: chart.YAxis{
	//			AxisType:  chart.YAxisSecondary,
	//			Name:      "Storage Idle Rate",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//		},
	//		Background: chart.Style{
	//			Padding: chart.Box{
	//				Top:  50,
	//				Left: 20,
	//			},
	//		},
	//		Series: []chart.Series{
	//			chart.ContinuousSeries{
	//				Name:    firstFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: firstFitRecorder.StorageIdleRecords,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    randomFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: randomFitRecorder.StorageIdleRecords,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    MCASGARecorder.Name,
	//				XValues: xValues,
	//				YValues: MCASGARecorder.StorageIdleRecords,
	//			},
	//		},
	//	}
	//
	//	graph.Elements = []chart.Renderable{
	//		chart.LegendThin(&graph),
	//	}
	//
	//	res.Header().Set("Content-Type", "image/png")
	//	err := graph.Render(chart.PNG, res)
	//	if err != nil {
	//		log.Println("Error: graph.Render(chart.PNG, res)", err)
	//	}
	//}
	//
	//var bwChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
	//	graph := chart.Chart{
	//		Title: "Bandwidth Idle Rate",
	//		XAxis: chart.XAxis{
	//			Name:      "Time, Number of New Applications",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//			ValueFormatter: func(v interface{}) string {
	//				return strconv.FormatInt(int64(v.(float64)), 10)
	//			},
	//			Ticks: ticks,
	//		},
	//		YAxis: chart.YAxis{
	//			AxisType:  chart.YAxisSecondary,
	//			Name:      "Bandwidth Idle Rate",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//		},
	//		Background: chart.Style{
	//			Padding: chart.Box{
	//				Top:  50,
	//				Left: 20,
	//			},
	//		},
	//		Series: []chart.Series{
	//			chart.ContinuousSeries{
	//				Name:    firstFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: firstFitRecorder.BwIdleRecords,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    randomFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: randomFitRecorder.BwIdleRecords,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    MCASGARecorder.Name,
	//				XValues: xValues,
	//				YValues: MCASGARecorder.BwIdleRecords,
	//			},
	//		},
	//	}
	//
	//	graph.Elements = []chart.Renderable{
	//		chart.LegendThin(&graph),
	//	}
	//
	//	res.Header().Set("Content-Type", "image/png")
	//	err := graph.Render(chart.PNG, res)
	//	if err != nil {
	//		log.Println("Error: graph.Render(chart.PNG, res)", err)
	//	}
	//}
	//
	//var priorityChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
	//
	//	graph := chart.Chart{
	//		Title: "Application Acceptance Rate",
	//		XAxis: chart.XAxis{
	//			Name:      "Time, Number of New Applications",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//			ValueFormatter: func(v interface{}) string {
	//				return strconv.FormatInt(int64(v.(float64)), 10)
	//			},
	//			Ticks: ticks,
	//		},
	//		YAxis: chart.YAxis{
	//			AxisType:  chart.YAxisSecondary,
	//			Name:      "Application Acceptance Rate",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//		},
	//		Background: chart.Style{
	//			Padding: chart.Box{
	//				Top:  50,
	//				Left: 20,
	//			},
	//		},
	//		Series: []chart.Series{
	//			chart.ContinuousSeries{
	//				Name:    firstFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: firstFitRecorder.AcceptedPriorityRateRecords,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    randomFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: randomFitRecorder.AcceptedPriorityRateRecords,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    MCASGARecorder.Name,
	//				XValues: xValues,
	//				YValues: MCASGARecorder.AcceptedPriorityRateRecords,
	//			},
	//		},
	//	}
	//
	//	graph.Elements = []chart.Renderable{
	//		chart.LegendThin(&graph),
	//	}
	//
	//	res.Header().Set("Content-Type", "image/png")
	//	err := graph.Render(chart.PNG, res)
	//	if err != nil {
	//		log.Println("Error: graph.Render(chart.PNG, res)", err)
	//	}
	//}
	//
	//var svcPriChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
	//
	//	graph := chart.Chart{
	//		Title: "Service Acceptance Rate",
	//		XAxis: chart.XAxis{
	//			Name:      "Time, Number of New Applications",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//			ValueFormatter: func(v interface{}) string {
	//				return strconv.FormatInt(int64(v.(float64)), 10)
	//			},
	//			Ticks: ticks,
	//		},
	//		YAxis: chart.YAxis{
	//			AxisType:  chart.YAxisSecondary,
	//			Name:      "Service Acceptance Rate",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//		},
	//		Background: chart.Style{
	//			Padding: chart.Box{
	//				Top:  50,
	//				Left: 20,
	//			},
	//		},
	//		Series: []chart.Series{
	//			chart.ContinuousSeries{
	//				Name:    firstFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: firstFitRecorder.AcceptedSvcPriRateRecords,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    randomFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: randomFitRecorder.AcceptedSvcPriRateRecords,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    MCASGARecorder.Name,
	//				XValues: xValues,
	//				YValues: MCASGARecorder.AcceptedSvcPriRateRecords,
	//			},
	//		},
	//	}
	//
	//	graph.Elements = []chart.Renderable{
	//		chart.LegendThin(&graph),
	//	}
	//
	//	res.Header().Set("Content-Type", "image/png")
	//	err := graph.Render(chart.PNG, res)
	//	if err != nil {
	//		log.Println("Error: graph.Render(chart.PNG, res)", err)
	//	}
	//}
	//
	//var taskPriChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
	//
	//	graph := chart.Chart{
	//		Title: "Task Acceptance Rate",
	//		XAxis: chart.XAxis{
	//			Name:      "Time, Number of New Applications",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//			ValueFormatter: func(v interface{}) string {
	//				return strconv.FormatInt(int64(v.(float64)), 10)
	//			},
	//			Ticks: ticks,
	//		},
	//		YAxis: chart.YAxis{
	//			AxisType:  chart.YAxisSecondary,
	//			Name:      "Task Acceptance Rate",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//		},
	//		Background: chart.Style{
	//			Padding: chart.Box{
	//				Top:  50,
	//				Left: 20,
	//			},
	//		},
	//		Series: []chart.Series{
	//			chart.ContinuousSeries{
	//				Name:    firstFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: firstFitRecorder.AcceptedTaskPriRateRecords,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    randomFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: randomFitRecorder.AcceptedTaskPriRateRecords,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    MCASGARecorder.Name,
	//				XValues: xValues,
	//				YValues: MCASGARecorder.AcceptedTaskPriRateRecords,
	//			},
	//		},
	//	}
	//
	//	graph.Elements = []chart.Renderable{
	//		chart.LegendThin(&graph),
	//	}
	//
	//	res.Header().Set("Content-Type", "image/png")
	//	err := graph.Render(chart.PNG, res)
	//	if err != nil {
	//		log.Println("Error: graph.Render(chart.PNG, res)", err)
	//	}
	//}
	//
	//var complTimeChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
	//	graph := chart.Chart{
	//		Title: "Completion Time",
	//		XAxis: chart.XAxis{
	//			Name:      "Time, Number of New Applications",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//			ValueFormatter: func(v interface{}) string {
	//				return strconv.FormatInt(int64(v.(float64)), 10)
	//			},
	//			Ticks: ticks,
	//		},
	//		YAxis: chart.YAxis{
	//			AxisType:  chart.YAxisSecondary,
	//			Name:      "Completion Time (s)",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//		},
	//		Background: chart.Style{
	//			Padding: chart.Box{
	//				Top:  50,
	//				Left: 20,
	//			},
	//		},
	//		Series: []chart.Series{
	//			chart.ContinuousSeries{
	//				Name:    firstFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: firstFitRecorder.AllAppComplTime,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    randomFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: randomFitRecorder.AllAppComplTime,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    MCASGARecorder.Name,
	//				XValues: xValues,
	//				YValues: MCASGARecorder.AllAppComplTime,
	//			},
	//		},
	//	}
	//
	//	graph.Elements = []chart.Renderable{
	//		chart.LegendThin(&graph),
	//	}
	//
	//	res.Header().Set("Content-Type", "image/png")
	//	err := graph.Render(chart.PNG, res)
	//	if err != nil {
	//		log.Println("Error: graph.Render(chart.PNG, res)", err)
	//	}
	//}
	//
	//var complTimePerPriChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
	//	graph := chart.Chart{
	//		Title: "Completion Time Per Priority",
	//		XAxis: chart.XAxis{
	//			Name:      "Time, Number of New Applications",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//			//ValueFormatter: func(v interface{}) string {
	//			//	return strconv.FormatInt(int64(v.(float64)), 10)
	//			//},
	//			Ticks: ticks,
	//		},
	//		YAxis: chart.YAxis{
	//			AxisType:  chart.YAxisSecondary,
	//			Name:      "Completion Time Per Priority (s)",
	//			NameStyle: chart.StyleShow(),
	//			Style:     chart.StyleShow(),
	//		},
	//		Background: chart.Style{
	//			Padding: chart.Box{
	//				Top:  50,
	//				Left: 20,
	//			},
	//		},
	//		Series: []chart.Series{
	//			chart.ContinuousSeries{
	//				Name:    firstFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: firstFitRecorder.AllAppComplTimePerPri,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    randomFitRecorder.Name,
	//				XValues: xValues,
	//				YValues: randomFitRecorder.AllAppComplTimePerPri,
	//			},
	//			chart.ContinuousSeries{
	//				Name:    MCASGARecorder.Name,
	//				XValues: xValues,
	//				YValues: MCASGARecorder.AllAppComplTimePerPri,
	//			},
	//		},
	//	}
	//
	//	graph.Elements = []chart.Renderable{
	//		chart.LegendThin(&graph),
	//	}
	//
	//	res.Header().Set("Content-Type", "image/png")
	//	err := graph.Render(chart.PNG, res)
	//	if err != nil {
	//		log.Println("Error: graph.Render(chart.PNG, res)", err)
	//	}
	//}
	//
	//http.HandleFunc("/CPUIdleRate", CPUChartFunc)
	//http.HandleFunc("/memoryIdleRate", memoryChartFunc)
	//http.HandleFunc("/storageIdleRate", storageChartFunc)
	//http.HandleFunc("/bwIdleRate", bwChartFunc)
	//http.HandleFunc("/acceptedPriority", priorityChartFunc)
	//http.HandleFunc("/acceptedSvcPri", svcPriChartFunc)
	//http.HandleFunc("/acceptedTaskPri", taskPriChartFunc)
	//http.HandleFunc("/complTime", complTimeChartFunc)
	//http.HandleFunc("/complTimePerPri", complTimePerPriChartFunc)
	//err := http.ListenAndServe(":8080", nil)
	//if err != nil {
	//	log.Println("Error: http.ListenAndServe(\":8080\", nil)", err)
	//}

}
