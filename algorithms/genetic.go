package algorithms

import (
	"fmt"
	"github.com/KeepTheBeats/routing-algorithms/random"
	"github.com/wcharczuk/go-chart"
	"gogeneticwrsp/model"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
)

type Chromosome []int
type Population []Chromosome

// PopulationCopy deep copy a population
func PopulationCopy(src Population) Population {
	var dst Population = make(Population, len(src))
	for i := 0; i < len(dst); i++ {
		dst[i] = make(Chromosome, len(src[i]))
		copy(dst[i], src[i])
	}
	return dst
}

type Genetic struct {
	ChromosomesCount       int
	IterationCount         int
	BestUntilNow           Chromosome
	BestAcceptableUntilNow Chromosome
	CrossoverProbability   float64
	MutationProbability    float64
	StopNoUpdateIteration  int

	FitnessRecordIterationBest   []float64
	FitnessRecordBestUntilNow    []float64
	BestUntilNowUpdateIterations []float64

	FitnessRecordIterationBestAcceptable   []float64
	FitnessRecordBestAcceptableUntilNow    []float64
	BestAcceptableUntilNowUpdateIterations []float64

	SelectableCloudsForApps [][]int

	InitFunc func([]model.Cloud, []model.Application) []int // the function to initialize populations

	RejectExecTime float64 // We set this time as the start time of rejected services and completion time of rejected tasks unit second
}

func NewGenetic(chromosomesCount int, iterationCount int, crossoverProbability float64, mutationProbability float64, stopNoUpdateIteration int, initFunc func([]model.Cloud, []model.Application) []int, clouds []model.Cloud, apps []model.Application) *Genetic {
	if err := model.DependencyValid(apps); err != nil {
		log.Panicf("model.DependencyValid(apps), err: %s", err.Error())
	}

	selectableCloudsForApps := make([][]int, len(apps))
	for i := 0; i < len(apps); i++ {
		if !apps[i].IsNew { // remaining apps, cannot be rejected
			if !apps[i].CanMigrate { // executing tasks and their dependent apps cannot be migrated
				selectableCloudsForApps[i] = append(selectableCloudsForApps[i], apps[i].CloudRemainingOn)
			} else {
				for j := 0; j < len(clouds); j++ {
					if CloudMeetApp(clouds[j], apps[i]) {
						selectableCloudsForApps[i] = append(selectableCloudsForApps[i], j)
					}
				}
			}
			continue
		}

		// new apps
		selectableCloudsForApps[i] = append(selectableCloudsForApps[i], len(clouds))
		for j := 0; j < len(clouds); j++ {
			if CloudMeetApp(clouds[j], apps[i]) {
				selectableCloudsForApps[i] = append(selectableCloudsForApps[i], j)
			}
		}
		// increase the possibility of rejecting
		originalLen := len(selectableCloudsForApps[i]) - 1 // how many selectable clouds except for "rejecting"
		for j := 0; j < originalLen-1; j++ {
			selectableCloudsForApps[i] = append(selectableCloudsForApps[i], len(clouds))
		}
	}

	// make sure the two variables acceptable
	var bestUntilNow, bestAcceptableUntilNow = make(Chromosome, len(apps)), make(Chromosome, len(apps))
	for i := 0; i < len(apps); i++ {
		bestUntilNow[i] = len(clouds)
		bestAcceptableUntilNow[i] = len(clouds)
	}

	return &Genetic{
		ChromosomesCount:                       chromosomesCount,
		IterationCount:                         iterationCount,
		BestUntilNow:                           bestUntilNow,
		BestAcceptableUntilNow:                 bestAcceptableUntilNow,
		CrossoverProbability:                   crossoverProbability,
		MutationProbability:                    mutationProbability,
		StopNoUpdateIteration:                  stopNoUpdateIteration,
		FitnessRecordBestUntilNow:              []float64{-1},
		FitnessRecordBestAcceptableUntilNow:    []float64{-1},
		BestUntilNowUpdateIterations:           []float64{-1}, // We define that the first BestUntilNow is set in the No. -1 iteration
		BestAcceptableUntilNowUpdateIterations: []float64{-1},
		SelectableCloudsForApps:                selectableCloudsForApps,
		InitFunc:                               initFunc,
	}
}

// Fitness calculate the fitness value of this scheduling result, the fitness values is possibly less than 0
func (g *Genetic) Fitness(clouds []model.Cloud, apps []model.Application, chromosome Chromosome) float64 {
	var deployedClouds []model.Cloud = SimulateDeploy(clouds, apps, model.Solution{SchedulingResult: chromosome})
	var appsCopy []model.Application = model.AppsCopy(apps)
	CalcStartComplTime(deployedClouds, appsCopy, chromosome)
	//for i := 0; i < len(deployedClouds); i++ {
	//	sort.Sort(model.AppSlice(deployedClouds[i].RunningApps))
	//	fmt.Println("Cloud:", i)
	//	for j := 0; j < len(deployedClouds[i].RunningApps); j++ {
	//		fmt.Println(deployedClouds[i].RunningApps[j].IsTask, deployedClouds[i].RunningApps[j].StartTime, deployedClouds[i].RunningApps[j].TaskCompletionTime)
	//	}
	//}
	//time.Sleep(101 * time.Second)

	var fitnessValue float64
	appsCopy = model.AppsCopy(apps)
	// the fitnessValue is based on each application
	for appIndex := 0; appIndex < len(chromosome); appIndex++ {
		fitnessValue += g.fitnessOneApp(deployedClouds, appsCopy[appIndex], chromosome[appIndex])
	}

	return fitnessValue
}

// CalcStartComplTime calculate the completion time of all tasks on all clouds, and the start time of all applications
// slice of Golang is a reference (address/pointer), so we can change the contents in the function
func CalcStartComplTime(clouds []model.Cloud, apps []model.Application, chromosome Chromosome) []model.Application {
	// initialization
	for i := 0; i < len(clouds); i++ {
		clouds[i].TotalTaskComplTime = 0
		clouds[i].TmpAlloc.CPU.LogicalCores = clouds[i].Allocatable.CPU.LogicalCores
	}
	// save the original order of apps
	unorderedApps := model.AppsCopy(apps)
	// traverse apps from high priority to low priority
	sort.Sort(model.AppSlice(apps))
	for k := 0; k < len(apps); k++ {
		// In this chromosome, this app is scheduled on this cloud
		cloudIndex := chromosome[apps[k].AppIdx]
		if cloudIndex == len(clouds) {
			continue // this app is rejected
		}
		for i := 0; i < len(clouds[cloudIndex].RunningApps); i++ {
			// find the app in the RunningApps of this cloud
			if clouds[cloudIndex].RunningApps[i].AppIdx == apps[k].AppIdx {
				// the start time of every app should be after all its dependent apps
				// apps is sorted by priority, and an app can only depend on others with higher priorities, so all of this app's dependence in unorderedApps already have the StartTime and TaskCompletionTime
				latestStartTime := clouds[cloudIndex].TotalTaskComplTime
				for j := 0; j < len(clouds[cloudIndex].RunningApps[i].Depend); j++ {
					if unorderedApps[clouds[cloudIndex].RunningApps[i].Depend[j].AppIdx].IsTask { // should be after the completion time of every dependent task
						if unorderedApps[clouds[cloudIndex].RunningApps[i].Depend[j].AppIdx].TaskCompletionTime > latestStartTime {
							latestStartTime = unorderedApps[clouds[cloudIndex].RunningApps[i].Depend[j].AppIdx].TaskCompletionTime
						}
					} else { // should be after the start time of every dependent service
						if unorderedApps[clouds[cloudIndex].RunningApps[i].Depend[j].AppIdx].StableTime > latestStartTime {
							latestStartTime = unorderedApps[clouds[cloudIndex].RunningApps[i].Depend[j].AppIdx].StableTime
						}
					}
				}
				clouds[cloudIndex].TotalTaskComplTime = latestStartTime
				clouds[cloudIndex].RunningApps[i].StartTime = latestStartTime
				unorderedApps[apps[k].AppIdx].StartTime = latestStartTime

				// calculate image pulling time, 1 Byte = 8 bits
				imagePullTime := (clouds[cloudIndex].RunningApps[i].ImageSize*8)/(clouds[cloudIndex].TmpAlloc.NetCondImage.DownBw*1024*1024) + (clouds[cloudIndex].TmpAlloc.NetCondImage.RTT / 1000) // unit: second

				// An old app with image pulling done on the old cloud, image will already exist
				if !clouds[cloudIndex].RunningApps[i].IsNew && clouds[cloudIndex].RunningApps[i].ImagePullDone && cloudIndex == clouds[cloudIndex].RunningApps[i].CloudRemainingOn {
					imagePullTime = 0
				}

				// calculate time of transmitting data from Architecture Controller to this cloud, 1 Byte = 8 bits
				dataInputTime := (clouds[cloudIndex].RunningApps[i].InputDataSize*8)/(clouds[cloudIndex].TmpAlloc.NetCondController.DownBw*1024*1024) + (clouds[cloudIndex].TmpAlloc.NetCondController.RTT / 1000) // unit: second

				// calculate the startup time of this application
				startUpTime := clouds[cloudIndex].RunningApps[i].StartUpCPUCycle / (clouds[cloudIndex].TmpAlloc.CPU.LogicalCores * clouds[cloudIndex].TmpAlloc.CPU.BaseClock * 1024 * 1024 * 1024) // unit: second

				// For an old app already stable on the old cloud, we can simply pause/unpause it through cgroup freezer, no need to input data or start up
				if !clouds[cloudIndex].RunningApps[i].IsNew && clouds[cloudIndex].RunningApps[i].AlreadyStable && cloudIndex == clouds[cloudIndex].RunningApps[i].CloudRemainingOn {
					dataInputTime = 0
					startUpTime = 0
				}

				// set image pull done time
				clouds[cloudIndex].RunningApps[i].ImagePullDoneTime = clouds[cloudIndex].RunningApps[i].StartTime + imagePullTime
				unorderedApps[apps[k].AppIdx].ImagePullDoneTime = unorderedApps[apps[k].AppIdx].StartTime + imagePullTime

				// set data input done time
				clouds[cloudIndex].RunningApps[i].DataInputDoneTime = clouds[cloudIndex].RunningApps[i].StartTime + imagePullTime + dataInputTime
				unorderedApps[apps[k].AppIdx].DataInputDoneTime = unorderedApps[apps[k].AppIdx].StartTime + imagePullTime + dataInputTime

				// set stable time
				clouds[cloudIndex].RunningApps[i].StableTime = clouds[cloudIndex].RunningApps[i].StartTime + imagePullTime + dataInputTime + startUpTime
				unorderedApps[apps[k].AppIdx].StableTime = unorderedApps[apps[k].AppIdx].StartTime + imagePullTime + dataInputTime + startUpTime

				if clouds[cloudIndex].RunningApps[i].IsTask { // Tasks do not take up the resources, but use all remaining resources to finish this task before handling other applications
					// task execution time
					execTime := clouds[cloudIndex].RunningApps[i].TaskReq.CPUCycle / (clouds[cloudIndex].TmpAlloc.CPU.LogicalCores * clouds[cloudIndex].TmpAlloc.CPU.BaseClock * 1024 * 1024 * 1024) // unit: second
					// a task should consume the three parts of time
					clouds[cloudIndex].TotalTaskComplTime += imagePullTime + dataInputTime + startUpTime + execTime
					clouds[cloudIndex].RunningApps[i].TaskCompletionTime = clouds[cloudIndex].TotalTaskComplTime
					unorderedApps[apps[k].AppIdx].TaskCompletionTime = clouds[cloudIndex].TotalTaskComplTime
				} else { // Services take up the resource
					// take up cpu
					clouds[cloudIndex].TmpAlloc.CPU.LogicalCores -= clouds[cloudIndex].RunningApps[i].SvcReq.CPUClock / clouds[cloudIndex].TmpAlloc.CPU.BaseClock

					// a service should consume the two parts of time
					clouds[cloudIndex].TotalTaskComplTime += imagePullTime + dataInputTime + startUpTime
				}

				break
			}
		}
	}

	// restore
	for i := 0; i < len(clouds); i++ {
		clouds[i].TmpAlloc.CPU.LogicalCores = clouds[i].Allocatable.CPU.LogicalCores
	}

	//for i := 0; i < len(unorderedApps); i++ {
	//	fmt.Printf("app: %d, isTask: %t, accept: %t, priority: %d, startTime: %g, completionTime: %g \ndepend on [ \n, ", i, unorderedApps[i].IsTask, chromosome[i] != len(clouds), unorderedApps[i].Priority, unorderedApps[i].StartTime, unorderedApps[i].TaskCompletionTime)
	//	for j := 0; j < len(unorderedApps[i].Depend); j++ {
	//		fmt.Printf("(app: %d, isTask: %t, accept: %t, priority: %d, startTime: %g, completionTime: %g) \n", unorderedApps[i].Depend[j].AppIdx, unorderedApps[unorderedApps[i].Depend[j].AppIdx].IsTask, chromosome[unorderedApps[i].Depend[j].AppIdx] != len(clouds), unorderedApps[unorderedApps[i].Depend[j].AppIdx].Priority, unorderedApps[unorderedApps[i].Depend[j].AppIdx].StartTime, unorderedApps[unorderedApps[i].Depend[j].AppIdx].TaskCompletionTime)
	//	}
	//	fmt.Println("]")
	//}
	//time.Sleep(100 * time.Second)
	return unorderedApps
}

// fitness of a single application, the fitness values is >= 0
func (g *Genetic) fitnessOneApp(clouds []model.Cloud, app model.Application, chosenCloudIndex int) float64 {
	if chosenCloudIndex == len(clouds) {
		return 0
	}

	if app.IsTask { // Task: minimize execution time
		for i := 0; i < len(clouds[chosenCloudIndex].RunningApps); i++ {
			// find this app in RunningApps of this cloud
			if clouds[chosenCloudIndex].RunningApps[i].AppIdx == app.AppIdx {
				thisTaskCompleTime := clouds[chosenCloudIndex].RunningApps[i].TaskCompletionTime
				// Maxmize (X - ta) * priority (if ta > X, we set ta = X)
				thisFitness := (g.RejectExecTime - thisTaskCompleTime) * float64(app.Priority)
				if thisFitness < 0 {
					thisFitness = 0
				}
				return thisFitness
			}
		}
		log.Panicln("Task, cannot find clouds[chosenCloudIndex].RunningApps[i].AppIdx == app.AppIdx")
	} else { // Service: maximize execution time
		for i := 0; i < len(clouds[chosenCloudIndex].RunningApps); i++ {
			// find this app in RunningApps of this cloud
			if clouds[chosenCloudIndex].RunningApps[i].AppIdx == app.AppIdx {
				thisSvcStableTime := clouds[chosenCloudIndex].RunningApps[i].StableTime
				//Maximize (X - tb) *priority (if tb > X, we set tb = X)
				thisFitness := (g.RejectExecTime - thisSvcStableTime) * float64(app.Priority)
				if thisFitness < 0 {
					thisFitness = 0
				}
				return thisFitness
			}
		}
		log.Panicln("Service, cannot find clouds[chosenCloudIndex].RunningApps[i].AppIdx == app.AppIdx")
	}
	log.Panicln("unreachable return 0")
	return 0
}

// old-version fitness of a single application
func fitnessOneAppOld(clouds []model.Cloud, app model.Application, chosenCloudIndex int) float64 {
	// The application is rejected, only set a small fitness, because in some aspects being rejected is better than being deployed incorrectly
	if chosenCloudIndex == len(clouds) {
		return 0
	}

	var subFitness []float64
	var overflow bool

	var cpuFitness float64 // fitness about CPUClock
	if clouds[chosenCloudIndex].TmpAlloc.CPU.LogicalCores >= 0 {
		cpuFitness = 1
	} else {
		// CPUClock is compressible resource in Kubernetes
		//overflowRate := clouds[chosenCloudIndex].Capacity.CPUClock / ((0 - clouds[chosenCloudIndex].TmpAlloc.CPUClock) + clouds[chosenCloudIndex].Capacity.CPUClock)
		cpuFitness = -1
		overflow = true
		// even CPUClock is compressible, I think we still should not allow it to overflow
	}
	subFitness = append(subFitness, cpuFitness)

	var memoryFitness float64 // fitness about memory
	if clouds[chosenCloudIndex].TmpAlloc.Memory >= 0 {
		memoryFitness = 1
	} else {
		// Memory is incompressible resource in Kubernetes, but the fitness is used for the selection in evolution
		// After evolution, we will only output the solution with no overflow resources
		//memoryFitness = 0
		//overflowRate := clouds[chosenCloudIndex].Capacity.Memory / ((0 - clouds[chosenCloudIndex].TmpAlloc.Memory) + clouds[chosenCloudIndex].Capacity.Memory)
		memoryFitness = -1
		overflow = true
	}
	subFitness = append(subFitness, memoryFitness)

	var storageFitness float64 // fitness about storage
	if clouds[chosenCloudIndex].TmpAlloc.Storage >= 0 {
		storageFitness = 1
	} else {
		// Storage is incompressible resource in Kubernetes, but the fitness is used for the selection in evolution
		// After evolution, we will only output the solution with no overflow resources
		//storageFitness = 0
		//overflowRate := clouds[chosenCloudIndex].Capacity.Storage / ((0 - clouds[chosenCloudIndex].TmpAlloc.Storage) + clouds[chosenCloudIndex].Capacity.Storage)
		storageFitness = -1
		overflow = true
	}
	subFitness = append(subFitness, storageFitness)

	// if overflow, do not add fitness
	if overflow {
		for i := 0; i < len(subFitness); i++ {
			if subFitness[i] > 0 {
				subFitness[i] = 0
			}
		}
	}

	var fitness float64
	for i := 0; i < len(subFitness); i++ {
		fitness += subFitness[i]
	}
	// the higher priority, the more important the application, the more its fitness should be scaled up
	fitness *= float64(app.Priority)

	return fitness
}

func (g *Genetic) initialize(clouds []model.Cloud, apps []model.Application) Population {
	var initPopulation Population
	// in a population, there are g.ChromosomesCount chromosomes (individuals)
	for i := 0; i < g.ChromosomesCount; i++ {
		//var chromosome Chromosome = InitializeUndeployedChromosome(clouds, apps)
		//var chromosome Chromosome = InitializeAcceptableChromosome(clouds, apps)
		//var chromosome Chromosome = RandomFitSchedule(clouds, apps)
		//var chromosome Chromosome = FirstFitSchedule(clouds, apps)
		var chromosome Chromosome = g.InitFunc(clouds, apps)
		initPopulation = append(initPopulation, chromosome)
	}
	return initPopulation
}

// initialize an Undeployed chromosome
func InitializeUndeployedChromosome(clouds []model.Cloud, apps []model.Application) []int {
	var chromosome Chromosome = make(Chromosome, len(apps))
	for i := 0; i < len(apps); i++ {
		chromosome[i] = len(clouds)
	}

	return chromosome
}

// initialize an acceptable chromosome
func InitializeAcceptableChromosome(clouds []model.Cloud, apps []model.Application) []int {
	var chromosome Chromosome = make(Chromosome, len(apps))
	for i := 0; i < len(apps); i++ {
		chromosome[i] = len(clouds)
	}

	var undeployed []int = make([]int, len(apps))
	for i := 0; i < len(apps); i++ {
		undeployed[i] = i
	}
	for len(undeployed) > 0 {
		appIndex := random.RandomInt(0, len(undeployed)-1)
		for i := 0; i < len(clouds); i++ {
			if !CloudMeetApp(clouds[i], apps[undeployed[appIndex]]) {
				continue
			}
			chromosome[undeployed[appIndex]] = i
			if Acceptable(clouds, apps, chromosome) {
				break
			}
			chromosome[undeployed[appIndex]] = len(clouds)
		}
		undeployed = append(undeployed[:appIndex], undeployed[appIndex+1:]...)
	}

	return chromosome
}

// When we need to randomly select a cloud for an app, we use this function to limit the range to g.SelectableCloudsForApps
func (g *Genetic) randomSelect(appIndex int) int {
	selectedIndex := random.RandomInt(0, len(g.SelectableCloudsForApps[appIndex])-1)
	gene := g.SelectableCloudsForApps[appIndex][selectedIndex]
	return gene
}

func (g *Genetic) selectionOperator(clouds []model.Cloud, apps []model.Application, population Population) Population {
	fitnesses := make([]float64, len(population))
	cumulativeFitnesses := make([]float64, len(population))

	// calculate the fitness of each chromosome in this population
	var maxFitness, minFitness float64 = -math.MaxFloat64, math.MaxFloat64 // record the max and min for standardization
	for i := 0; i < len(population); i++ {
		fitness := g.Fitness(clouds, apps, population[i])

		fitnesses[i] = fitness
		if fitness > maxFitness {
			maxFitness = fitness
		}
		if fitness < minFitness {
			minFitness = fitness
		}
	}
	//log.Println("fitnesses:", fitnesses)

	// standardization
	// If all fitnesses are big, there differences will be relatively small, and they will be very close in the aspect of proportion,
	// so in roulette-wheel selection, the possibility of choosing big fitness will be only a little higher than the possibility of choosing small fitness.
	// Therefore, we need to add an offset to all fitnesses to standardize them, to give them proper difference in the aspect of proportion.
	//log.Println("Original fitness", fitnesses)

	// this value means that "How many times the maximum value is the minimum value"
	// which is equivalent to "How many times the possibility of choosing maximum value is that of choosing the minimum value"
	maxTimesMin := 10.0
	maxMinDiff := maxFitness - minFitness
	standardizedMinFitness := maxMinDiff / (maxTimesMin - 1)
	offset := standardizedMinFitness - minFitness
	//log.Printf("maxFitness %f, minFitness %f, standardizedMinFitness %f, offset %f", maxFitness, minFitness, standardizedMinFitness, offset)
	for i := 0; i < len(fitnesses); i++ {
		fitnesses[i] += offset
	}
	//log.Println("Standardized fitness", fitnesses)

	// make unacceptable chromosomes harder to be selected
	for i := 0; i < len(fitnesses); i++ {
		if !Acceptable(clouds, apps, population[i]) {
			fitnesses[i] /= 3
		}
	}

	// calculate the cumulative fitnesses of the chromosomes in this population
	cumulativeFitnesses[0] = fitnesses[0]
	for i := 1; i < len(population); i++ {
		cumulativeFitnesses[i] = cumulativeFitnesses[i-1] + fitnesses[i]
	}

	// select g.ChromosomesCount chromosomes to generate a new population
	var newPopulation Population
	var bestFitnessInThisIteration float64 = -1
	var bestFitnessInThisIterationIndex int
	var bestAcceptableFitnessInThisIteration float64 = -1
	var bestAcceptableFitnessInThisIterationIndex int

	for i := 0; i < g.ChromosomesCount; i++ {

		// roulette-wheel selection
		// selectionThreshold is a random float64 in [0, biggest cumulative fitness])
		selectionThreshold := random.RandomFloat64(0, cumulativeFitnesses[len(cumulativeFitnesses)-1])
		// selected the smallest cumulativeFitnesses that is begger than selectionThreshold
		selectedChromosomeIndex := sort.SearchFloat64s(cumulativeFitnesses, selectionThreshold)
		//log.Printf("selectionThreshold %f,selectedChromosomeIndex %d", selectionThreshold, selectedChromosomeIndex)

		// cannot directly append population[selectedChromosomeIndex], otherwise, the new population may be changed by strange reasons
		newChromosome := make(Chromosome, len(population[selectedChromosomeIndex]))
		copy(newChromosome, population[selectedChromosomeIndex])
		newPopulation = append(newPopulation, newChromosome)

		// record the best fitness in this iteration
		chosenFitness := g.Fitness(clouds, apps, population[selectedChromosomeIndex])
		//log.Printf("selectedChromosomeIndex %d, population[selectedChromosomeIndex] %d, chosenFitness %f", selectedChromosomeIndex, population[selectedChromosomeIndex], chosenFitness)

		if chosenFitness > bestFitnessInThisIteration {
			bestFitnessInThisIteration = chosenFitness
			bestFitnessInThisIterationIndex = selectedChromosomeIndex
			if Acceptable(clouds, apps, population[selectedChromosomeIndex]) {
				bestAcceptableFitnessInThisIteration = chosenFitness
				bestAcceptableFitnessInThisIterationIndex = selectedChromosomeIndex
			}

		}
	}

	//if Fitness(clouds, apps, g.BestUntilNow) != g.FitnessRecordBestUntilNow[len(g.FitnessRecordBestUntilNow)-1] {
	//	log.Println("not equal:", g.FitnessRecordBestUntilNow[len(g.FitnessRecordBestUntilNow)-1], Fitness(clouds, apps, g.BestUntilNow))
	//}

	// record:
	// the best chromosome until now;
	// the best fitness record;
	// the best fitness in every iteration;
	// the iterations during which the g.BestUntilNow is updated;

	// the best fitness in every iteration;
	g.FitnessRecordIterationBest = append(g.FitnessRecordIterationBest, bestFitnessInThisIteration)
	if bestFitnessInThisIteration > g.FitnessRecordBestUntilNow[len(g.FitnessRecordBestUntilNow)-1] {

		//log.Printf("In iteration %d, update g.FitnessRecordBestUntilNow from %f to %f, update g.BestUntilNow from %v, to %v\n", len(g.FitnessRecordIterationBest)-1, g.FitnessRecordBestUntilNow[len(g.FitnessRecordBestUntilNow)-1], bestFitnessInThisIteration, g.BestUntilNow, population[bestFitnessInThisIterationIndex])
		// the best chromosome until now;
		// if we directly use "=", g.BestUntilNow may be changed by strange reasons.
		copy(g.BestUntilNow, population[bestFitnessInThisIterationIndex])
		// the best fitness record in all iterations until now;
		g.FitnessRecordBestUntilNow = append(g.FitnessRecordBestUntilNow, bestFitnessInThisIteration)
		// the iteration during which the g.BestUntilNow is updated
		// len(g.FitnessRecordIterationBest)-1 is the current iteration number, there are IterationCount+1 iterations in total
		g.BestUntilNowUpdateIterations = append(g.BestUntilNowUpdateIterations, float64(len(g.FitnessRecordIterationBest)-1))
	}

	g.FitnessRecordIterationBestAcceptable = append(g.FitnessRecordIterationBestAcceptable, bestAcceptableFitnessInThisIteration)
	//fmt.Println("bestAcceptableFitnessInThisIteration", bestAcceptableFitnessInThisIteration)
	//fmt.Println("Fitness(clouds, apps, g.BestAcceptableUntilNow)", Fitness(clouds, apps, g.BestAcceptableUntilNow))
	//fmt.Println("g.BestAcceptableUntilNow)", g.BestAcceptableUntilNow)
	if bestAcceptableFitnessInThisIteration >= 0 && bestAcceptableFitnessInThisIteration > g.FitnessRecordBestAcceptableUntilNow[len(g.FitnessRecordBestAcceptableUntilNow)-1] {
		copy(g.BestAcceptableUntilNow, population[bestAcceptableFitnessInThisIterationIndex])
		g.FitnessRecordBestAcceptableUntilNow = append(g.FitnessRecordBestAcceptableUntilNow, bestAcceptableFitnessInThisIteration)
		g.BestAcceptableUntilNowUpdateIterations = append(g.BestAcceptableUntilNowUpdateIterations, float64(len(g.FitnessRecordIterationBestAcceptable)-1))
	}

	return newPopulation
}

func (g *Genetic) crossoverOperator(clouds []model.Cloud, apps []model.Application, population Population) Population {
	if len(apps) <= 1 { // only with at least 2 genes in a chromosome, can we do crossover
		return population
	}
	// avoid changing the original population, maybe not needed but for security
	var copyPopulation Population = PopulationCopy(population)

	// traverse all chromosomes in this population, use random to judge whether a chromosome needs crossover
	var indexesNeedCrossover []int
	for i := 0; i < len(copyPopulation); i++ {
		if random.RandomFloat64(0, 1) < g.CrossoverProbability {
			indexesNeedCrossover = append(indexesNeedCrossover, i)
		}
	}

	//log.Println("indexesNeedCrossover:", indexesNeedCrossover)

	var newPopulation Population

	// randomly choose pairs of chromosomes to do crossover
	var whetherCrossover []bool = make([]bool, len(population))
	for len(indexesNeedCrossover) > 1 { // if len(indexesNeedCrossover) <= 1, stop crossover
		// choose two indexes of chromosomes for crossover;
		// delete them from indexesNeedCrossover;
		// mark them in whetherCrossover
		// first index
		first := random.RandomInt(0, len(indexesNeedCrossover)-1)
		firstIndex := indexesNeedCrossover[first]
		whetherCrossover[firstIndex] = true                                                            // mark
		indexesNeedCrossover = append(indexesNeedCrossover[:first], indexesNeedCrossover[first+1:]...) // delete
		// second index
		second := random.RandomInt(0, len(indexesNeedCrossover)-1)
		secondIndex := indexesNeedCrossover[second]
		whetherCrossover[secondIndex] = true                                                             // mark
		indexesNeedCrossover = append(indexesNeedCrossover[:second], indexesNeedCrossover[second+1:]...) // delete

		firstChromosome := copyPopulation[firstIndex]
		secondChromosome := copyPopulation[secondIndex]

		// randomly choose a gene after which the genes are exchanged
		startPosition := random.RandomInt(1, len(apps)-1) // in each chromosome, there are len(apps) genes

		//log.Println("firstIndex:", firstIndex, "secondIndex:", secondIndex, "startPosition:", startPosition)

		// cut the firstChromosome and secondChromosome at the startPosition
		firstHead := make([]int, len(firstChromosome[:startPosition]))
		copy(firstHead, firstChromosome[:startPosition])

		firstTail := make([]int, len(firstChromosome[startPosition:]))
		copy(firstTail, firstChromosome[startPosition:])

		secondHead := make([]int, len(secondChromosome[:startPosition]))
		copy(secondHead, secondChromosome[:startPosition])

		secondTail := make([]int, len(secondChromosome[startPosition:]))
		copy(secondTail, secondChromosome[startPosition:])

		// generate new chromosomes
		newFirstChromosome := append(firstHead, secondTail...)
		newSecondChromosome := append(secondHead, firstTail...)

		// append the two new chromosomes in newPopulation
		newPopulation = append(newPopulation, newFirstChromosome, newSecondChromosome)
	}

	//log.Println("whetherCrossover:", whetherCrossover)

	// directly put the chromosomes with no crossover to the new population
	for i := 0; i < len(copyPopulation); i++ {
		if !whetherCrossover[i] {
			newPopulation = append(newPopulation, copyPopulation[i])
		}
	}

	return newPopulation
}

func (g *Genetic) mutationOperator(clouds []model.Cloud, apps []model.Application, population Population) Population {
	// avoid changing the original population, maybe not needed but for security
	var copyPopulation Population = PopulationCopy(population)

	// Traverse every gene. Every gene has a probability of g.MutationProbability that it do mutation
	for i := 0; i < len(copyPopulation); i++ {
		for j := 0; j < len(copyPopulation[i]); j++ {
			// use random to judge whether a gene needs mutation
			if random.RandomFloat64(0, 1) < g.MutationProbability {
				var newGene int = g.randomSelect(j)
				// make sure that the mutated gene is different with the original one
				for newGene != len(clouds) && newGene == copyPopulation[i][j] && len(g.SelectableCloudsForApps[j]) > 1 {
					newGene = g.randomSelect(j)
				}
				//log.Printf("gene [%d][%d] mutates from %d to %d\n", i, j, copyPopulation[i][j], newGene)
				copyPopulation[i][j] = newGene
			}
		}

		fixDependence(clouds, apps, copyPopulation[i])

		// After mutation, if the chromosome becomes unacceptable, we discard it, and randomly generate a new acceptable one
		// This is to control the population mutate to good direction
		if !Acceptable(clouds, apps, copyPopulation[i]) {
			copyPopulation[i] = RandomFitSchedule(clouds, apps)
		}
	}

	return copyPopulation
}

// for every app, if any of its dependent apps is rejected, we also reject it.
// fix dependence for a chromosome, avoiding regenerating due to invalid.
// regenerating means restart to evolve
func fixDependence(clouds []model.Cloud, apps []model.Application, chromosome Chromosome) {
	// sometimes when we fix a dependence, we change an app A to be rejected, then we should also change the apps depending on A to be rejected in the following cycles
	var oldAccepted []int
	var newAccepted []int
	for i := 0; i < len(apps); i++ {
		if chromosome[i] == len(clouds) {
			continue
		}
		for j := 0; j < len(apps[i].Depend); j++ {
			if chromosome[apps[i].Depend[j].AppIdx] == len(clouds) {
				chromosome[i] = len(clouds)
				break
			}
		}
		if chromosome[i] != len(clouds) {
			oldAccepted = append(oldAccepted, i)
			newAccepted = append(newAccepted, i)
		}
	}

	for {
		//if len(oldAccepted) != len(newAccepted) {
		//	fmt.Println("old", oldAccepted)
		//	fmt.Println("new", newAccepted)
		//}
		oldAccepted = newAccepted
		newAccepted = make([]int, 0)
		for i := 0; i < len(oldAccepted); i++ {
			for j := 0; j < len(apps[oldAccepted[i]].Depend); j++ {
				if chromosome[apps[oldAccepted[i]].Depend[j].AppIdx] == len(clouds) {
					chromosome[oldAccepted[i]] = len(clouds)
					break
				}
			}
			if chromosome[oldAccepted[i]] != len(clouds) {
				newAccepted = append(newAccepted, oldAccepted[i])
			}
		}
		if len(oldAccepted) == len(newAccepted) {
			break
		}
	}
}

func (g *Genetic) initRejectTime(clouds []model.Cloud, apps []model.Application, initPopulation Population) {
	var nonZeroChroNum int
	var complTimes []float64
	for i := 0; i < len(initPopulation); i++ {
		tmpClouds := model.CloudsCopy(clouds)
		tmpApps := model.AppsCopy(apps)
		tmpSolution := model.SolutionCopy(model.Solution{SchedulingResult: initPopulation[i]})

		tmpClouds = SimulateDeploy(tmpClouds, tmpApps, tmpSolution)
		CalcStartComplTime(tmpClouds, tmpApps, tmpSolution.SchedulingResult)
		for j := 0; j < len(tmpClouds); j++ {
			if tmpClouds[j].TotalTaskComplTime > 0 {
				complTimes = append(complTimes, tmpClouds[j].TotalTaskComplTime)
				nonZeroChroNum++
			}

		}
	}

	sort.Float64s(complTimes)

	// after some tests these values of the parameters seem good
	// high acceptance
	var cutRate float64 = 0.1
	var longestPara float64 = 1.5
	var averagePara float64 = 4
	var timesPara float64 = 2 // to give difference between rejection and long execution time

	//// lower application consuming time, these are bad!!!!!!
	//var cutRate float64 = 0.1
	//var longestPara float64 = 1
	//var averagePara float64 = 2
	//var timesPara float64 = 2 // to give difference between rejection and long execution time

	var start, end int = int(float64(nonZeroChroNum) * cutRate), int(float64(nonZeroChroNum) * (1 - cutRate))
	var totalMidRange float64 = 0
	for i := start; i <= end; i++ {
		totalMidRange += complTimes[i]
	}
	var midAve float64 = totalMidRange / float64(end-start+1)
	g.RejectExecTime = timesPara * math.Min(complTimes[end]*longestPara, midAve*averagePara) // double insurance, in case that either of them is abnormal
}

func (g *Genetic) Schedule(clouds []model.Cloud, apps []model.Application) (model.Solution, error) {
	// make sure that all time attributes of each app are 0
	for i := 0; i < len(apps); i++ {
		if apps[i].StartTime != 0 {
			log.Panicf("apps[%d].StartTime is %g", i, apps[i].StartTime)
		}
		if apps[i].ImagePullDoneTime != 0 {
			log.Panicf("apps[%d].ImagePullDoneTime is %g", i, apps[i].ImagePullDoneTime)
		}
		if apps[i].DataInputDoneTime != 0 {
			log.Panicf("apps[%d].DataInputDoneTime is %g", i, apps[i].DataInputDoneTime)
		}
		if apps[i].StableTime != 0 {
			log.Panicf("apps[%d].StableTime is %g", i, apps[i].StableTime)
		}
		if apps[i].TaskCompletionTime != 0 {
			log.Panicf("apps[%d].TaskCompletionTime is %g", i, apps[i].TaskCompletionTime)
		}
		//apps[i].StartTime = 0
		//apps[i].ImagePullDoneTime = 0
		//apps[i].TaskCompletionTime = 0
	}

	// initialize a population
	var initPopulation Population = g.initialize(clouds, apps)
	//for i, chromosome := range initPopulation {
	//	log.Println(i, chromosome, len(chromosome))
	//}
	g.initRejectTime(clouds, apps, initPopulation)
	log.Println("g.RejectExecTime:", g.RejectExecTime)

	//log.Println("---selection after initialization-------")
	// there are IterationCount+1 iterations in total, this is the No. 0 iteration
	currentPopulation := g.selectionOperator(clouds, apps, initPopulation) // Iteration No. 0
	//for i, chromosome := range currentPopulation {
	//	log.Println(i, chromosome)
	//}

	// No. 1 iteration to No. g.IterationCount iteration
	for iteration := 1; iteration <= g.IterationCount; iteration++ {
		//log.Printf("---crossover in iteration %d-------\n", iteration)
		currentPopulation = g.crossoverOperator(clouds, apps, currentPopulation)
		//for i, chromosome := range currentPopulation {
		//	log.Println(i, chromosome)
		//}

		for i := 0; i < len(currentPopulation); i++ {
			//for j := 0; j < len(currentPopulation[i]); j++ {
			//	if currentPopulation[i][j] == len(clouds) {
			//		continue
			//	}
			//	for k := 0; k < len(apps[j].Depend); k++ {
			//		if currentPopulation[i][apps[j].Depend[k].AppIdx] == len(clouds) {
			//			log.Println("before fix")
			//			log.Println(j, k, currentPopulation[i][j], currentPopulation[i][apps[j].Depend[k].AppIdx])
			//		}
			//	}
			//}
			fixDependence(clouds, apps, currentPopulation[i])
			for j := 0; j < len(currentPopulation[i]); j++ {
				if currentPopulation[i][j] == len(clouds) {
					continue
				}
				for k := 0; k < len(apps[j].Depend); k++ {
					if currentPopulation[i][apps[j].Depend[k].AppIdx] == len(clouds) {
						log.Panicln(j, k, currentPopulation[i][j], currentPopulation[i][apps[j].Depend[k].AppIdx])
					}
				}
			}
		}

		//log.Printf("--------mutation in iteration %d-------\n", iteration)
		currentPopulation = g.mutationOperator(clouds, apps, currentPopulation)
		//for i, chromosome := range currentPopulation {
		//	log.Println(i, chromosome)
		//}
		//log.Printf("--------selection in iteration %d-------\n", iteration)
		currentPopulation = g.selectionOperator(clouds, apps, currentPopulation)
		//for i, chromosome := range currentPopulation {
		//	log.Println(i, chromosome)
		//}

		clouds1 := model.CloudsCopy(clouds)
		apps1 := model.AppsCopy(apps)
		solution1 := model.SolutionCopy(model.Solution{
			SchedulingResult: g.BestAcceptableUntilNow,
		})
		//log.Println(Acceptable(clouds1, apps1, solution1.SchedulingResult))
		if !Acceptable(clouds1, apps1, solution1.SchedulingResult) {
			log.Panicln()
		}

		// if at least one acceptable solution has been found, and if the best fitness until now has not been updated for a certain number of iterations, we think that the solution is already stable enough, and stop the algorithm
		if len(g.BestAcceptableUntilNowUpdateIterations) > 1 && float64(iteration)-g.BestUntilNowUpdateIterations[len(g.BestUntilNowUpdateIterations)-1] > float64(g.StopNoUpdateIteration) {
			break
		}
	}

	if len(g.BestAcceptableUntilNowUpdateIterations) == 1 {
		return model.Solution{}, fmt.Errorf("no acceptable solution is found in %d iterations", g.IterationCount)
	}

	return model.Solution{SchedulingResult: g.BestAcceptableUntilNow}, nil
}

// DrawChart draw g.FitnessRecordIterationBest and g.FitnessRecordBestUntilNow on a line chart
func (g *Genetic) DrawChart() {
	var drawChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
		var xValuesIterationBest []float64
		for i, _ := range g.FitnessRecordIterationBest {
			xValuesIterationBest = append(xValuesIterationBest, float64(i))
		}

		graph := chart.Chart{
			Title: "Evolution",
			//TitleStyle: chart.Style{
			//	Show: true,
			//},
			//Width: 600,
			//Height: 1800,
			//DPI:    300,
			XAxis: chart.XAxis{
				Name:      "Iteration Number",
				NameStyle: chart.StyleShow(),
				Style:     chart.StyleShow(),
				ValueFormatter: func(v interface{}) string {
					return strconv.FormatInt(int64(v.(float64)), 10)
				},
			},
			YAxis: chart.YAxis{
				AxisType:  chart.YAxisSecondary,
				Name:      "Fitness",
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
					Name:    "Best Fitness in each iteration",
					XValues: xValuesIterationBest,
					YValues: g.FitnessRecordIterationBest,
				},
				chart.ContinuousSeries{
					Name: "Best Fitness in all iterations",
					// the first value (iteration -1) is much different with others, which will cause that we cannot observer the trend of the evolution
					XValues: g.BestUntilNowUpdateIterations[1:],
					YValues: g.FitnessRecordBestUntilNow[1:],
					Style: chart.Style{
						Show:            true,
						StrokeDashArray: []float64{5.0, 3.0, 2.0, 3.0},
						StrokeWidth:     1,
					},
				},
				chart.ContinuousSeries{
					Name:    "Best Acceptable Fitness in each iterations",
					XValues: xValuesIterationBest,
					YValues: g.FitnessRecordIterationBestAcceptable,
					Style: chart.Style{
						Show:            true,
						StrokeDashArray: []float64{2.0, 3.0},
						StrokeWidth:     1,
					},
				},
				chart.ContinuousSeries{
					Name:    "Best Acceptable Fitness in all iterations",
					XValues: g.BestAcceptableUntilNowUpdateIterations[1:],
					YValues: g.FitnessRecordBestAcceptableUntilNow[1:],
					Style: chart.Style{
						Show:            true,
						StrokeDashArray: []float64{5.0, 3.0},
						StrokeWidth:     1,
					},
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

	http.HandleFunc("/", drawChartFunc)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("Error: http.ListenAndServe(\":8080\", nil)", err)
	}
}
