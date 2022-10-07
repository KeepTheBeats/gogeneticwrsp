package experimenttools

import (
	"encoding/json"
	"fmt"
	"github.com/wcharczuk/go-chart"
	"go/build"
	"gogeneticwrsp/algorithms"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"runtime"
	"strconv"

	"gogeneticwrsp/model"
)

func getFilePaths(numCloud, numApp int) (string, string) {
	var cloudPath, appPath string
	if runtime.GOOS == "windows" {
		cloudPath = fmt.Sprintf("%s\\src\\gogeneticwrsp\\experimenttools\\cloud_%d.json", build.Default.GOPATH, numCloud)
		appPath = fmt.Sprintf("%s\\src\\gogeneticwrsp\\experimenttools\\app_%d.json", build.Default.GOPATH, numApp)
	} else {
		cloudPath = fmt.Sprintf("%s/src/gogeneticwrsp/experimenttools/cloud_%d.json", build.Default.GOPATH, numCloud)
		appPath = fmt.Sprintf("%s/src/gogeneticwrsp/experimenttools/app_%d.json", build.Default.GOPATH, numApp)
	}
	return cloudPath, appPath
}

// GenerateCloudsApps generates clouds and apps and writes them into files.
func GenerateCloudsApps(numCloud, numApp int) {

	log.Printf("generate %d clouds and %d applications and write them into files\n", numCloud, numApp)

	var clouds []model.Cloud = make([]model.Cloud, numCloud)
	var apps []model.Application = make([]model.Application, numApp)

	// generate clouds
	cloudDiffTimes := 2.0 // give clouds different types
	for i := 0; i < numCloud; i++ {
		clouds[i].Capacity.CPU = generateResourceCPUCapacity(16, 132)
		clouds[i].Capacity.Memory = generateResourceMemoryStorageCapacity(math.Pow(2, 30), math.Pow(2, 36))
		clouds[i].Capacity.Storage = generateResourceMemoryStorageCapacity(math.Pow(2, 32), math.Pow(2, 39))
		clouds[i].Capacity.NetLatency = generateResourceNetLatency(0, 4, 2, 2)

		// give clouds different types
		if i <= numCloud/4 {
			clouds[i].Capacity.CPU /= cloudDiffTimes
		}
		if i > numCloud/4 && i <= numCloud/2 {
			clouds[i].Capacity.Memory /= cloudDiffTimes
		}
		if i > numCloud/2 && i <= int(float64(numCloud)*0.75) {
			clouds[i].Capacity.Storage /= cloudDiffTimes
		}
		if i > int(float64(numCloud)*0.75) {
			clouds[i].Capacity.NetLatency *= cloudDiffTimes
		}

		clouds[i].Allocatable = clouds[i].Capacity
	}

	// generate applications
	appDiffTimes := 2.0 // give clouds different types
	for i := 0; i < numApp; i++ {
		apps[i].Requests.CPU = generateResourceCPU(1, 10, 3, 5, false)
		apps[i].Requests.Memory = generateResourceMemoryStorageRequest(math.Pow(2, 26), math.Pow(2, 32), math.Pow(2, 30), 4000*math.Pow(2, 20))

		apps[i].Requests.Storage = generateResourceMemoryStorageRequest(0, math.Pow(2, 35), math.Pow(2, 33), 24*math.Pow(2, 30))
		apps[i].Requests.NetLatency = generateResourceNetLatency(1, 5, 3, 3)
		apps[i].Priority = generatePriority(100, 65535.9, 150, 300)

		// give applications different types
		if i <= numApp/4 {
			apps[i].Requests.CPU *= appDiffTimes
		}
		if i > numApp/4 && i <= numApp/2 {
			apps[i].Requests.Memory *= appDiffTimes
		}
		if i > numApp/2 && i <= int(float64(numApp)*0.75) {
			apps[i].Requests.Storage *= appDiffTimes
		}
		if i > int(float64(numApp)*0.75) {
			apps[i].Requests.NetLatency /= appDiffTimes
		}
	}

	cloudsJson, err := json.Marshal(clouds)
	if err != nil {
		log.Fatalln("json.Marshal(clouds) error:", err.Error())

	}
	appsJson, err := json.Marshal(apps)
	if err != nil {
		log.Fatalln("json.Marshal(apps) error:", err.Error())
	}

	var cloudPath, appPath string = getFilePaths(numCloud, numApp)

	err = ioutil.WriteFile(cloudPath, cloudsJson, 0777)
	if err != nil {
		log.Fatalln("ioutil.WriteFile(cloudPath, cloudsJson, 0777) error:", err.Error())
	}
	err = ioutil.WriteFile(appPath, appsJson, 0777)
	if err != nil {
		log.Fatalln("ioutil.WriteFile(appPath, appsJson, 0777) error:", err.Error())
	}
}

// ReadCloudsApps from files
func ReadCloudsApps(numCloud, numApp int) ([]model.Cloud, []model.Application) {
	var cloudPath, appPath string = getFilePaths(numCloud, numApp)
	var clouds []model.Cloud
	var apps []model.Application

	cloudsJson, err := ioutil.ReadFile(cloudPath)
	if err != nil {
		log.Fatalln("ioutil.ReadFile(cloudPath) error:", err.Error())
	}
	appsJson, err := ioutil.ReadFile(appPath)
	if err != nil {
		log.Fatalln("ioutil.ReadFile(appPath) error:", err.Error())
	}

	err = json.Unmarshal(cloudsJson, &clouds)
	if err != nil {
		log.Fatalln("json.Unmarshal(cloudsJson, &clouds) error:", err.Error())
	}
	err = json.Unmarshal(appsJson, &apps)
	if err != nil {
		log.Fatalln("json.Unmarshal(appsJson, &apps) error:", err.Error())
	}

	return clouds, apps
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
		log.Println(experimenter.Name+", fitness:", algorithms.Fitness(clouds, apps, experimenter.ExperimentSolution.SchedulingResult), "CPU Idle Rate:", algorithms.CPUIdleRate(clouds, apps, experimenter.ExperimentSolution.SchedulingResult), "Memory Idle Rate:", algorithms.MemoryIdleRate(clouds, apps, experimenter.ExperimentSolution.SchedulingResult), "Storage Idle Rate:", algorithms.StorageIdleRate(clouds, apps, experimenter.ExperimentSolution.SchedulingResult), "Total Accepted Priority:", algorithms.AcceptedPriority(clouds, apps, experimenter.ExperimentSolution.SchedulingResult))
		if v, ok := experimenter.ExperimentAlgorithm.(*algorithms.Genetic); ok {
			log.Println("Got Iteration:", v.BestUntilNowUpdateIterations[len(v.BestUntilNowUpdateIterations)-1], v.BestAcceptableUntilNowUpdateIterations[len(v.BestAcceptableUntilNowUpdateIterations)-1])
		}
	}
}

type ContinuousHelper struct {
	Name                        string
	CPUIdleRecords              []float64
	MemoryIdleRecords           []float64
	StorageIdleRecords          []float64
	AcceptedPriorityRateRecords []float64
}

// ContinuousExperiment is that the applications are deployed one by one. In one time, we only handle one application.
func ContinuousExperiment(clouds []model.Cloud, apps []model.Application) {

	var currentClouds []model.Cloud
	var currentApps []model.Application
	var currentSolution model.Solution

	var firstFitRecorder ContinuousHelper = ContinuousHelper{
		Name:                        "First Fit",
		CPUIdleRecords:              make([]float64, 0),
		MemoryIdleRecords:           make([]float64, 0),
		StorageIdleRecords:          make([]float64, 0),
		AcceptedPriorityRateRecords: make([]float64, 0),
	}
	var randomFitRecorder ContinuousHelper = ContinuousHelper{
		Name:                        "Random Fit",
		CPUIdleRecords:              make([]float64, 0),
		MemoryIdleRecords:           make([]float64, 0),
		StorageIdleRecords:          make([]float64, 0),
		AcceptedPriorityRateRecords: make([]float64, 0),
	}
	// Multi-cloud Applications Scheduling Genetic Algorithm (MCASGA)
	var MCASGARecorder ContinuousHelper = ContinuousHelper{
		Name:                        "MCASGA",
		CPUIdleRecords:              make([]float64, 0),
		MemoryIdleRecords:           make([]float64, 0),
		StorageIdleRecords:          make([]float64, 0),
		AcceptedPriorityRateRecords: make([]float64, 0),
	}

	// First Fit
	currentClouds = make([]model.Cloud, len(clouds))
	copy(currentClouds, clouds)
	currentApps = []model.Application{}
	currentSolution = model.Solution{}

	for i := 0; i < len(apps); i++ {
		// the ith applications request comes
		currentApps = append(currentApps, apps[i])
		thisApp := []model.Application{apps[i]}
		ff := algorithms.NewFirstFit(currentClouds, thisApp)
		solution, err := ff.Schedule(currentClouds, thisApp)
		if err != nil {
			log.Printf("Error, app %d. Error message: %s", i, err.Error())
		}
		// deploy this app in current clouds (subtract the resources)
		currentClouds = algorithms.SimulateDeploy(currentClouds, thisApp, solution)
		// add the solution of this app to current solution
		currentSolution.SchedulingResult = append(currentSolution.SchedulingResult, solution.SchedulingResult...)
		// evaluate current solution, current cloud, current apps
		firstFitRecorder.CPUIdleRecords = append(firstFitRecorder.CPUIdleRecords, algorithms.CPUIdleRate(clouds, currentApps, currentSolution.SchedulingResult))
		firstFitRecorder.MemoryIdleRecords = append(firstFitRecorder.MemoryIdleRecords, algorithms.MemoryIdleRate(clouds, currentApps, currentSolution.SchedulingResult))
		firstFitRecorder.StorageIdleRecords = append(firstFitRecorder.StorageIdleRecords, algorithms.StorageIdleRate(clouds, currentApps, currentSolution.SchedulingResult))
		firstFitRecorder.AcceptedPriorityRateRecords = append(firstFitRecorder.AcceptedPriorityRateRecords, float64(algorithms.AcceptedPriority(clouds, currentApps, currentSolution.SchedulingResult))/float64(algorithms.TotalPriority(clouds, currentApps, currentSolution.SchedulingResult)))
	}

	// Random Fit
	currentClouds = make([]model.Cloud, len(clouds))
	copy(currentClouds, clouds)
	currentApps = []model.Application{}
	currentSolution = model.Solution{}

	for i := 0; i < len(apps); i++ {
		// the ith applications request comes
		currentApps = append(currentApps, apps[i])
		thisApp := []model.Application{apps[i]}
		rf := algorithms.NewRandomFit(currentClouds, thisApp)
		solution, err := rf.Schedule(currentClouds, thisApp)
		if err != nil {
			log.Printf("Error, app %d. Error message: %s", i, err.Error())
		}
		// deploy this app in current clouds (subtract the resources)
		currentClouds = algorithms.SimulateDeploy(currentClouds, thisApp, solution)
		// add the solution of this app to current solution
		currentSolution.SchedulingResult = append(currentSolution.SchedulingResult, solution.SchedulingResult...)
		// evaluate current solution, current cloud, current apps
		randomFitRecorder.CPUIdleRecords = append(randomFitRecorder.CPUIdleRecords, algorithms.CPUIdleRate(clouds, currentApps, currentSolution.SchedulingResult))
		randomFitRecorder.MemoryIdleRecords = append(randomFitRecorder.MemoryIdleRecords, algorithms.MemoryIdleRate(clouds, currentApps, currentSolution.SchedulingResult))
		randomFitRecorder.StorageIdleRecords = append(randomFitRecorder.StorageIdleRecords, algorithms.StorageIdleRate(clouds, currentApps, currentSolution.SchedulingResult))
		randomFitRecorder.AcceptedPriorityRateRecords = append(randomFitRecorder.AcceptedPriorityRateRecords, float64(algorithms.AcceptedPriority(clouds, currentApps, currentSolution.SchedulingResult))/float64(algorithms.TotalPriority(clouds, currentApps, currentSolution.SchedulingResult)))
	}

	// MCASGA
	currentClouds = make([]model.Cloud, len(clouds))
	copy(currentClouds, clouds)
	currentApps = []model.Application{}
	currentSolution = model.Solution{}

	for i := 0; i < len(apps); i++ {
		// the ith applications request comes
		currentApps = append(currentApps, apps[i])
		//ga := algorithms.NewGenetic(100, 5000, 0.7, 0.007, 5000, algorithms.InitializeUndeployedChromosome, clouds, currentApps)
		ga := algorithms.NewGenetic(100, 5000, 0.7, 0.007, 200, algorithms.RandomFitSchedule, clouds, currentApps)
		solution, err := ga.Schedule(clouds, currentApps)
		if err != nil {
			log.Printf("Error, app %d. Error message: %s", i, err.Error())
		}
		// deploy this app in current clouds (subtract the resources)
		currentClouds = algorithms.SimulateDeploy(clouds, currentApps, solution)
		// add the solution of this app to current solution
		currentSolution = solution
		// evaluate current solution, current cloud, current apps
		MCASGARecorder.CPUIdleRecords = append(MCASGARecorder.CPUIdleRecords, algorithms.CPUIdleRate(clouds, currentApps, currentSolution.SchedulingResult))
		MCASGARecorder.MemoryIdleRecords = append(MCASGARecorder.MemoryIdleRecords, algorithms.MemoryIdleRate(clouds, currentApps, currentSolution.SchedulingResult))
		MCASGARecorder.StorageIdleRecords = append(MCASGARecorder.StorageIdleRecords, algorithms.StorageIdleRate(clouds, currentApps, currentSolution.SchedulingResult))
		MCASGARecorder.AcceptedPriorityRateRecords = append(MCASGARecorder.AcceptedPriorityRateRecords, float64(algorithms.AcceptedPriority(clouds, currentApps, currentSolution.SchedulingResult))/float64(algorithms.TotalPriority(clouds, currentApps, currentSolution.SchedulingResult)))
	}

	// draw line charts
	var CPUChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
		var appNumbers []float64
		for i, _ := range MCASGARecorder.CPUIdleRecords {
			appNumbers = append(appNumbers, float64(i+1))
		}

		graph := chart.Chart{
			Title: "CPU Idle Rate",
			XAxis: chart.XAxis{
				Name:      "Number of Applications",
				NameStyle: chart.StyleShow(),
				Style:     chart.StyleShow(),
				ValueFormatter: func(v interface{}) string {
					return strconv.FormatInt(int64(v.(float64)), 10)
				},
			},
			YAxis: chart.YAxis{
				AxisType:  chart.YAxisSecondary,
				Name:      "CPU Idle Rate",
				NameStyle: chart.StyleShow(),
				Style:     chart.StyleShow(),
			},
			Background: chart.Style{
				Padding: chart.Box{
					Top:  50,
					Left: 20,
				},
			},
			Series: []chart.Series{
				chart.ContinuousSeries{
					Name:    firstFitRecorder.Name,
					XValues: appNumbers,
					YValues: firstFitRecorder.CPUIdleRecords,
				},
				chart.ContinuousSeries{
					Name:    randomFitRecorder.Name,
					XValues: appNumbers,
					YValues: randomFitRecorder.CPUIdleRecords,
				},
				chart.ContinuousSeries{
					Name:    MCASGARecorder.Name,
					XValues: appNumbers,
					YValues: MCASGARecorder.CPUIdleRecords,
				},
			},
		}

		graph.Elements = []chart.Renderable{
			chart.LegendThin(&graph),
		}

		res.Header().Set("Content-Type", "image/png")
		err := graph.Render(chart.PNG, res)
		if err != nil {
			log.Println("Error: graph.Render(chart.PNG, res)", err)
		}
	}

	var memoryChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
		var appNumbers []float64
		for i, _ := range MCASGARecorder.MemoryIdleRecords {
			appNumbers = append(appNumbers, float64(i+1))
		}

		graph := chart.Chart{
			Title: "Memory Idle Rate",
			XAxis: chart.XAxis{
				Name:      "Number of Applications",
				NameStyle: chart.StyleShow(),
				Style:     chart.StyleShow(),
				ValueFormatter: func(v interface{}) string {
					return strconv.FormatInt(int64(v.(float64)), 10)
				},
			},
			YAxis: chart.YAxis{
				AxisType:  chart.YAxisSecondary,
				Name:      "Memory Idle Rate",
				NameStyle: chart.StyleShow(),
				Style:     chart.StyleShow(),
			},
			Background: chart.Style{
				Padding: chart.Box{
					Top:  50,
					Left: 20,
				},
			},
			Series: []chart.Series{
				chart.ContinuousSeries{
					Name:    firstFitRecorder.Name,
					XValues: appNumbers,
					YValues: firstFitRecorder.MemoryIdleRecords,
				},
				chart.ContinuousSeries{
					Name:    randomFitRecorder.Name,
					XValues: appNumbers,
					YValues: randomFitRecorder.MemoryIdleRecords,
				},
				chart.ContinuousSeries{
					Name:    MCASGARecorder.Name,
					XValues: appNumbers,
					YValues: MCASGARecorder.MemoryIdleRecords,
				},
			},
		}

		graph.Elements = []chart.Renderable{
			chart.LegendThin(&graph),
		}

		res.Header().Set("Content-Type", "image/png")
		err := graph.Render(chart.PNG, res)
		if err != nil {
			log.Println("Error: graph.Render(chart.PNG, res)", err)
		}
	}

	var storageChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
		var appNumbers []float64
		for i, _ := range MCASGARecorder.MemoryIdleRecords {
			appNumbers = append(appNumbers, float64(i+1))
		}

		graph := chart.Chart{
			Title: "Storage Idle Rate",
			XAxis: chart.XAxis{
				Name:      "Number of Applications",
				NameStyle: chart.StyleShow(),
				Style:     chart.StyleShow(),
				ValueFormatter: func(v interface{}) string {
					return strconv.FormatInt(int64(v.(float64)), 10)
				},
			},
			YAxis: chart.YAxis{
				AxisType:  chart.YAxisSecondary,
				Name:      "Storage Idle Rate",
				NameStyle: chart.StyleShow(),
				Style:     chart.StyleShow(),
			},
			Background: chart.Style{
				Padding: chart.Box{
					Top:  50,
					Left: 20,
				},
			},
			Series: []chart.Series{
				chart.ContinuousSeries{
					Name:    firstFitRecorder.Name,
					XValues: appNumbers,
					YValues: firstFitRecorder.StorageIdleRecords,
				},
				chart.ContinuousSeries{
					Name:    randomFitRecorder.Name,
					XValues: appNumbers,
					YValues: randomFitRecorder.StorageIdleRecords,
				},
				chart.ContinuousSeries{
					Name:    MCASGARecorder.Name,
					XValues: appNumbers,
					YValues: MCASGARecorder.StorageIdleRecords,
				},
			},
		}

		graph.Elements = []chart.Renderable{
			chart.LegendThin(&graph),
		}

		res.Header().Set("Content-Type", "image/png")
		err := graph.Render(chart.PNG, res)
		if err != nil {
			log.Println("Error: graph.Render(chart.PNG, res)", err)
		}
	}

	var priorityChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
		var appNumbers []float64
		for i, _ := range MCASGARecorder.MemoryIdleRecords {
			appNumbers = append(appNumbers, float64(i+1))
		}

		graph := chart.Chart{
			Title: "Application Acceptance Rate",
			XAxis: chart.XAxis{
				Name:      "Number of Applications",
				NameStyle: chart.StyleShow(),
				Style:     chart.StyleShow(),
				ValueFormatter: func(v interface{}) string {
					return strconv.FormatInt(int64(v.(float64)), 10)
				},
			},
			YAxis: chart.YAxis{
				AxisType:  chart.YAxisSecondary,
				Name:      "Application Acceptance Rate",
				NameStyle: chart.StyleShow(),
				Style:     chart.StyleShow(),
			},
			Background: chart.Style{
				Padding: chart.Box{
					Top:  50,
					Left: 20,
				},
			},
			Series: []chart.Series{
				chart.ContinuousSeries{
					Name:    firstFitRecorder.Name,
					XValues: appNumbers,
					YValues: firstFitRecorder.AcceptedPriorityRateRecords,
				},
				chart.ContinuousSeries{
					Name:    randomFitRecorder.Name,
					XValues: appNumbers,
					YValues: randomFitRecorder.AcceptedPriorityRateRecords,
				},
				chart.ContinuousSeries{
					Name:    MCASGARecorder.Name,
					XValues: appNumbers,
					YValues: MCASGARecorder.AcceptedPriorityRateRecords,
				},
			},
		}

		graph.Elements = []chart.Renderable{
			chart.LegendThin(&graph),
		}

		res.Header().Set("Content-Type", "image/png")
		err := graph.Render(chart.PNG, res)
		if err != nil {
			log.Println("Error: graph.Render(chart.PNG, res)", err)
		}
	}

	http.HandleFunc("/CPUIdleRate", CPUChartFunc)
	http.HandleFunc("/memoryIdleRate", memoryChartFunc)
	http.HandleFunc("/storageIdleRate", storageChartFunc)
	http.HandleFunc("/acceptedPriority", priorityChartFunc)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("Error: http.ListenAndServe(\":8080\", nil)", err)
	}

}
