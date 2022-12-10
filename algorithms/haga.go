package algorithms

import (
	"fmt"
	"github.com/KeepTheBeats/routing-algorithms/random"
	"gogeneticwrsp/model"
	"log"
	"math"
	"sort"
)

type HAGA struct {
	GroupNum         int
	GroupSize        int     // task group size
	VMGamma          float64 // VM Selection constant
	SelectedCloudNum int

	MutationStartingPos int
	MutationEndingPos   int

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

	RejectExecTime float64

	SchedulingResult []int        // The apps are divided into groups, so we should record the current scheduling result
	CloudPheBind     []idxPheBind // the bind from the cloud index to the pheromone on clouds
}

func NewHAGA(groupNum int, vmGamma float64, chromosomesCount int, iterationCount int, crossoverProbability float64, mutationProbability float64, stopNoUpdateIteration int, clouds []model.Cloud, apps []model.Application) *HAGA {
	if err := model.DependencyValid(apps); err != nil {
		log.Panicf("model.DependencyValid(apps), err: %s", err.Error())
	}

	return &HAGA{
		GroupNum:         groupNum,
		GroupSize:        int(math.Ceil(float64(len(apps)) / float64(groupNum))),
		VMGamma:          vmGamma,
		SelectedCloudNum: int(math.Ceil(float64(len(clouds)) * vmGamma)),

		ChromosomesCount:      chromosomesCount,
		IterationCount:        iterationCount,
		CrossoverProbability:  crossoverProbability,
		MutationProbability:   mutationProbability,
		StopNoUpdateIteration: stopNoUpdateIteration,
	}
}

func (h *HAGA) Schedule(clouds []model.Cloud, apps []model.Application) (model.Solution, error) {
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
	}

	// initialize the scheduling result
	var schedulingResult []int = make([]int, len(apps))
	for i := 0; i < len(apps); i++ {
		schedulingResult[i] = len(clouds)
	}
	h.SchedulingResult = schedulingResult

	// divide applications into groups
	log.Println("GroupSize:", h.GroupSize, "GroupNum:", h.GroupNum)
	var groups [][]int
	var thisGroup []int
	var chosenApps map[int]struct{} = make(map[int]struct{})

	// return the input idx and all apps have dependence relationship with idx
	var addDepApps func(int) []int = func(idx int) []int {
		var depApps []int = []int{idx}
		chosenApps[idx] = struct{}{}
		for {
			beforeLen := len(depApps)
			for i := 0; i < len(depApps); i++ { // for all added apps
				for j := 0; j < len(apps); j++ { // check not added apps
					if _, exist := chosenApps[j]; exist {
						continue
					}
					// add apps that depend on it or are depended on by it
					if model.CheckDepend(apps, idx, j) || model.CheckDepend(apps, j, idx) {
						depApps = append(depApps, j)
						chosenApps[j] = struct{}{}
					}
				}
			}
			afterLen := len(depApps)
			if beforeLen == afterLen { // no find new
				break
			}
		}
		return depApps
	}

	for i := 0; i < len(apps); i++ {
		if _, exist := chosenApps[i]; exist {
			continue
		}

		addGroup := addDepApps(i)
		thisGroup = append(thisGroup, addGroup...)

		if len(chosenApps) == len(apps) || len(thisGroup) > h.GroupSize {
			groups = append(groups, thisGroup)
			thisGroup = []int{}
		}
	}

	// initialize this bind to store the pheromone of clouds
	var cloudPheBind []idxPheBind = make([]idxPheBind, len(clouds))
	for i := 0; i < len(clouds); i++ {
		cloudPheBind[i] = idxPheBind{
			idx:       i,
			pheromone: 0,
		}
	}
	h.CloudPheBind = cloudPheBind

	var appGroup, cloudGroup []int
	var cloudGroupSize int
	for i := 0; i < len(groups); i++ {
		appGroup = groups[i]
		cloudGroup = []int{}
		if i == 0 {
			cloudGroupSize = len(clouds)
		} else {
			// evaporate pheromone from all clouds
			h.pheromoneEvaporation(clouds, apps, h.SchedulingResult)
			sort.Sort(pheSort(h.CloudPheBind)) // sort by order of pheromone
			log.Println("pheromone:", h.CloudPheBind)
			cloudGroupSize = h.SelectedCloudNum
		}
		// select clouds for this app group
		for j := 0; j < cloudGroupSize; j++ {
			cloudGroup = append(cloudGroup, h.CloudPheBind[j].idx)
		}
		log.Println("To schedule group:", appGroup)
		log.Println("on the clouds:", cloudGroup)

		thisResult, err := h.scheduleGroup(appGroup, cloudGroup, clouds, apps)
		if err != nil {
			log.Panicf("scheduleGroup error: %s", err.Error())
		}

		// combine thisResult into schedulingResult
		log.Println(thisResult)
		h.SchedulingResult = h.mergeResults(h.SchedulingResult, thisResult, appGroup)

	}

	return model.Solution{SchedulingResult: h.SchedulingResult}, nil
}

// bind a cloud index to its pheromone level
type idxPheBind struct {
	idx       int
	pheromone float64
}

type pheSort []idxPheBind

func (p pheSort) Len() int {
	return len(p)
}

func (p pheSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p pheSort) Less(i, j int) bool {
	return p[i].pheromone < p[j].pheromone
}

func (h *HAGA) pheromoneEvaporation(clouds []model.Cloud, apps []model.Application, schedulingResult []int) {
	cloudsCopy := model.CloudsCopy(clouds)
	appsCopy := model.AppsCopy(apps)
	solutionCopy := model.SolutionCopy(model.Solution{
		SchedulingResult: schedulingResult,
	})

	var deployedClouds []model.Cloud = SimulateDeploy(cloudsCopy, appsCopy, solutionCopy)

	var appGroup, cloudGroup []int
	var appGroupMap, cloudGroupMap map[int]struct{} = make(map[int]struct{}), make(map[int]struct{})
	for i := 0; i < len(cloudsCopy); i++ {
		cloudGroup = append(cloudGroup, i)
		cloudGroupMap[i] = struct{}{}
	}
	for i := 0; i < len(appsCopy); i++ {
		appGroup = append(appGroup, i)
		appGroupMap[i] = struct{}{}
	}

	appsCopy = h.calculateTaskTime(appGroupMap, cloudGroupMap, appGroup, cloudGroup, deployedClouds, appsCopy, solutionCopy.SchedulingResult)

	var acceptedTaskNum int
	var totalExecTime float64

	for i := 0; i < len(appsCopy); i++ {
		if !appsCopy[i].IsTask || solutionCopy.SchedulingResult[i] == len(deployedClouds) {
			continue
		}
		acceptedTaskNum++
		totalExecTime += appsCopy[i].TaskCompletionTime
	}

	pheromoneToEvaporate := totalExecTime / float64(acceptedTaskNum)

	for i := 0; i < len(h.CloudPheBind); i++ {
		h.CloudPheBind[i].pheromone -= pheromoneToEvaporate
		if h.CloudPheBind[i].pheromone < 0 {
			h.CloudPheBind[i].pheromone = 0
		}
	}

}

// after scheduling one group, we need to add pheromone
func (h *HAGA) addPheromone(appGroupMap, cloudGroupMap map[int]struct{}, appGroup, cloudGroup []int, clouds []model.Cloud, apps []model.Application) {
	cloudsCopy := model.CloudsCopy(clouds)
	appsCopy := model.AppsCopy(apps)
	solutionCopy := model.SolutionCopy(model.Solution{
		SchedulingResult: h.BestAcceptableUntilNow,
	})

	var deployedClouds []model.Cloud = SimulateDeploy(cloudsCopy, appsCopy, solutionCopy)

	appsCopy = h.calculateTaskTime(appGroupMap, cloudGroupMap, appGroup, cloudGroup, deployedClouds, appsCopy, solutionCopy.SchedulingResult)

	var totalExecTime float64
	for i := 0; i < len(appGroup); i++ {
		if !appsCopy[appGroup[i]].IsTask {
			continue
		}
		totalExecTime += appsCopy[appGroup[i]].TaskCompletionTime
	}

	pheromoneToAdd := totalExecTime

	for i := 0; i < len(deployedClouds); i++ {
		if _, cloudChosen := cloudGroupMap[i]; !cloudChosen {
			continue
		}
		if len(deployedClouds[i].RunningApps) == 0 {
			continue
		}
		// only add pheromone to the chosen clouds where apps are running
		for j := 0; j < len(h.CloudPheBind); j++ {
			if h.CloudPheBind[j].idx == i {
				h.CloudPheBind[j].pheromone += pheromoneToAdd
			}
		}
	}
}

// schedule a group of applications on some selected clouds
func (h *HAGA) scheduleGroup(appGroup, cloudGroup []int, clouds []model.Cloud, apps []model.Application) ([]int, error) {
	appGroupMap := make(map[int]struct{})
	cloudGroupMap := make(map[int]struct{})
	for i := 0; i < len(appGroup); i++ {
		appGroupMap[appGroup[i]] = struct{}{}
	}
	for i := 0; i < len(cloudGroup); i++ {
		cloudGroupMap[cloudGroup[i]] = struct{}{}
	}

	// this group of application can only be deployed on the chosen clouds
	selectableCloudsForApps := make([][]int, len(apps))
	for i := 0; i < len(apps); i++ {
		if _, appChosen := appGroupMap[i]; !appChosen {
			continue // only handle the chosen apps
		}
		selectableCloudsForApps[i] = append(selectableCloudsForApps[i], len(clouds))
		for j := 0; j < len(cloudGroup); j++ {
			selectableCloudsForApps[i] = append(selectableCloudsForApps[i], cloudGroup[j])
		}
		// increase the possibility of rejecting
		originalLen := len(selectableCloudsForApps[i]) - 1 // how many selectable clouds except for "rejecting"
		for j := 0; j < originalLen-1; j++ {
			selectableCloudsForApps[i] = append(selectableCloudsForApps[i], len(clouds))
		}
	}
	h.SelectableCloudsForApps = selectableCloudsForApps

	h.FitnessRecordIterationBest = []float64{}
	h.FitnessRecordBestUntilNow = []float64{-1}
	h.BestUntilNowUpdateIterations = []float64{-1}
	h.FitnessRecordIterationBestAcceptable = []float64{}
	h.FitnessRecordBestAcceptableUntilNow = []float64{-1}
	h.BestAcceptableUntilNowUpdateIterations = []float64{-1}

	// make sure the two variables acceptable
	var bestUntilNow, bestAcceptableUntilNow = make(Chromosome, len(apps)), make(Chromosome, len(apps))
	for i := 0; i < len(apps); i++ {
		bestUntilNow[i] = len(clouds)
		bestAcceptableUntilNow[i] = len(clouds)
	}
	h.BestUntilNow = bestUntilNow
	h.BestAcceptableUntilNow = bestAcceptableUntilNow

	h.MutationStartingPos = random.RandomInt(0, len(appGroup)-1)
	h.MutationEndingPos = random.RandomInt(0, len(appGroup)-1)

	// initialize a population
	var initPopulation Population = h.initialize(appGroupMap, cloudGroupMap, appGroup, cloudGroup, clouds, apps)
	//for i, chromosome := range initPopulation {
	//	log.Println(i, chromosome, len(chromosome))
	//}
	h.initRejectFitness(appGroupMap, cloudGroupMap, appGroup, cloudGroup, clouds, apps, initPopulation)
	log.Println("h.RejectExecTime:", h.RejectExecTime)

	currentPopulation := h.selectionOperator(appGroupMap, cloudGroupMap, appGroup, cloudGroup, clouds, apps, initPopulation) // Iteration No. 0

	// No. 1 iteration to No. g.IterationCount iteration
	for iteration := 1; iteration <= h.IterationCount; iteration++ {
		currentPopulation = h.crossoverOperator(appGroupMap, cloudGroupMap, appGroup, cloudGroup, clouds, apps, currentPopulation)

		for i := 0; i < len(currentPopulation); i++ {
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

		// the description in the paper is not clear, so I implement the mutation according to my understanding
		currentPopulation = h.mutationOperator(appGroupMap, cloudGroupMap, appGroup, cloudGroup, clouds, apps, currentPopulation)
		h.MutationStartingPos++
		if h.MutationStartingPos == len(appGroup) {
			h.MutationStartingPos = 0
		}
		h.MutationEndingPos--
		if h.MutationEndingPos == -1 {
			h.MutationEndingPos = len(appGroup) - 1
		}

		currentPopulation = h.selectionOperator(appGroupMap, cloudGroupMap, appGroup, cloudGroup, clouds, apps, currentPopulation)

		clouds1 := model.CloudsCopy(clouds)
		apps1 := model.AppsCopy(apps)
		solution1 := model.SolutionCopy(model.Solution{
			SchedulingResult: h.BestAcceptableUntilNow,
		})
		//log.Println(Acceptable(clouds1, apps1, solution1.SchedulingResult))
		if !Acceptable(clouds1, apps1, solution1.SchedulingResult) {
			log.Panicln()
		}

		// if at least one acceptable solution has been found, and if the best fitness until now has not been updated for a certain number of iterations, we think that the solution is already stable enough, and stop the algorithm
		if len(h.BestAcceptableUntilNowUpdateIterations) > 1 && float64(iteration)-h.BestUntilNowUpdateIterations[len(h.BestUntilNowUpdateIterations)-1] > float64(h.StopNoUpdateIteration) {
			break
		}

	}

	if len(h.BestAcceptableUntilNowUpdateIterations) == 1 {
		return []int{}, fmt.Errorf("no acceptable solution is found in %d iterations", h.IterationCount)
	}

	h.addPheromone(appGroupMap, cloudGroupMap, appGroup, cloudGroup, clouds, apps)

	return h.BestAcceptableUntilNow, nil
}

func (h *HAGA) initialize(appGroupMap map[int]struct{}, cloudGroupMap map[int]struct{}, appGroup, cloudGroup []int, clouds []model.Cloud, apps []model.Application) Population {
	var initPopulation Population
	// in a population, there are g.ChromosomesCount chromosomes (individuals)
	for i := 0; i < h.ChromosomesCount; i++ {
		var chromosome Chromosome = h.randomFitSchedule(appGroupMap, cloudGroupMap, appGroup, cloudGroup, clouds, apps)
		initPopulation = append(initPopulation, chromosome)
	}
	return initPopulation
}

// merge the additional scheduling result of this app group into the base scheduling result
func (h *HAGA) mergeResults(baseResult, additionalResult []int, appGroup []int) []int {
	var mergedResult []int = make([]int, len(baseResult))
	copy(mergedResult, baseResult)
	for i := 0; i < len(appGroup); i++ {
		mergedResult[appGroup[i]] = additionalResult[appGroup[i]]
	}
	return mergedResult
}

func (h *HAGA) randomFitSchedule(appGroupMap map[int]struct{}, cloudGroupMap map[int]struct{}, appGroup, cloudGroup []int, clouds []model.Cloud, apps []model.Application) []int {
	var schedulingResult []int = make([]int, len(apps))
	// set all new applications undeployed
	for i := 0; i < len(apps); i++ {
		schedulingResult[i] = len(clouds)
	}

	// traverse apps in random order
	var undeployed []int
	for i := 0; i < len(apps); i++ {
		if _, appChosen := appGroupMap[i]; appChosen {
			undeployed = append(undeployed, i) // record the original index of undeployed apps
		}
	}

	for len(undeployed) > 0 {
		appIndex := random.RandomInt(0, len(undeployed)-1) // appIndex in undeployed

		// traverse clouds in random order
		var untried []int
		for i := 0; i < len(clouds); i++ {
			if _, cloudChosen := cloudGroupMap[i]; cloudChosen {
				untried = append(untried, i) // record the original index of untried clouds
			}
		}

		for len(untried) > 0 {
			cloudIndex := random.RandomInt(0, len(untried)-1) // cloudIndex in untried
			if !CloudMeetApp(clouds[untried[cloudIndex]], apps[undeployed[appIndex]]) {
				untried = append(untried[:cloudIndex], untried[cloudIndex+1:]...)
				continue
			}
			schedulingResult[undeployed[appIndex]] = untried[cloudIndex]

			mergedResult := h.mergeResults(h.SchedulingResult, schedulingResult, appGroup)

			if Acceptable(clouds, apps, mergedResult) {
				untried = append(untried[:cloudIndex], untried[cloudIndex+1:]...)
				break
			}

			schedulingResult[undeployed[appIndex]] = len(clouds)
			untried = append(untried[:cloudIndex], untried[cloudIndex+1:]...)
		}

		undeployed = append(undeployed[:appIndex], undeployed[appIndex+1:]...)
	}

	return schedulingResult
}

func (h *HAGA) selectionOperator(appGroupMap, cloudGroupMap map[int]struct{}, appGroup, cloudGroup []int, clouds []model.Cloud, apps []model.Application, population Population) Population {
	fitnesses := make([]float64, len(population))
	cumulativeFitnesses := make([]float64, len(population))

	// calculate the fitness of each chromosome in this population
	var maxFitness, minFitness float64 = -math.MaxFloat64, math.MaxFloat64 // record the max and min for standardization
	for i, chromosome := range population {
		fitness := h.Fitness(appGroupMap, cloudGroupMap, appGroup, cloudGroup, clouds, apps, chromosome)

		fitnesses[i] = fitness
		if fitness > maxFitness {
			maxFitness = fitness
		}
		if fitness < minFitness {
			minFitness = fitness
		}
	}

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

	var newPopulation Population
	var bestFitnessInThisIteration float64 = -1
	var bestFitnessInThisIterationIndex int
	var bestAcceptableFitnessInThisIteration float64 = -1
	var bestAcceptableFitnessInThisIterationIndex int

	for i := 0; i < h.ChromosomesCount; i++ {

		var selectedChromosomeIndex int

		// roulette-wheel selection
		// selectionThreshold is a random float64 in [0, biggest cumulative fitness])
		selectionThreshold := random.RandomFloat64(0, cumulativeFitnesses[len(cumulativeFitnesses)-1])
		// selected the smallest cumulativeFitnesses that is begger than selectionThreshold
		selectedChromosomeIndex = sort.SearchFloat64s(cumulativeFitnesses, selectionThreshold)
		//log.Printf("selectionThreshold %f,selectedChromosomeIndex %d", selectionThreshold, selectedChromosomeIndex)

		// cannot directly append population[selectedChromosomeIndex], otherwise, the new population may be changed by strange reasons
		newChromosome := make(Chromosome, len(population[selectedChromosomeIndex]))
		copy(newChromosome, population[selectedChromosomeIndex])
		newPopulation = append(newPopulation, newChromosome)

		// record the best fitness in this iteration
		chosenFitness := h.Fitness(appGroupMap, cloudGroupMap, appGroup, cloudGroup, clouds, apps, population[selectedChromosomeIndex])
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

	//log.Println(bestFitnessInThisIteration, bestFitnessInThisIterationIndex, bestAcceptableFitnessInThisIteration, bestAcceptableFitnessInThisIterationIndex)

	// the best fitness in every iteration;
	h.FitnessRecordIterationBest = append(h.FitnessRecordIterationBest, bestFitnessInThisIteration)
	if bestFitnessInThisIteration > h.FitnessRecordBestUntilNow[len(h.FitnessRecordBestUntilNow)-1] {

		//log.Printf("In iteration %d, update g.FitnessRecordBestUntilNow from %f to %f, update g.BestUntilNow from %v, to %v\n", len(g.FitnessRecordIterationBest)-1, g.FitnessRecordBestUntilNow[len(g.FitnessRecordBestUntilNow)-1], bestFitnessInThisIteration, g.BestUntilNow, population[bestFitnessInThisIterationIndex])
		// the best chromosome until now;
		// if we directly use "=", g.BestUntilNow may be changed by strange reasons.
		copy(h.BestUntilNow, population[bestFitnessInThisIterationIndex])
		// the best fitness record in all iterations until now;
		h.FitnessRecordBestUntilNow = append(h.FitnessRecordBestUntilNow, bestFitnessInThisIteration)
		// the iteration during which the g.BestUntilNow is updated
		// len(g.FitnessRecordIterationBest)-1 is the current iteration number, there are IterationCount+1 iterations in total
		h.BestUntilNowUpdateIterations = append(h.BestUntilNowUpdateIterations, float64(len(h.FitnessRecordIterationBest)-1))
	}

	h.FitnessRecordIterationBestAcceptable = append(h.FitnessRecordIterationBestAcceptable, bestAcceptableFitnessInThisIteration)
	//fmt.Println("bestAcceptableFitnessInThisIteration", bestAcceptableFitnessInThisIteration)
	//fmt.Println("Fitness(clouds, apps, g.BestAcceptableUntilNow)", Fitness(clouds, apps, g.BestAcceptableUntilNow))
	//fmt.Println("g.BestAcceptableUntilNow)", g.BestAcceptableUntilNow)
	if bestAcceptableFitnessInThisIteration >= 0 && bestAcceptableFitnessInThisIteration > h.FitnessRecordBestAcceptableUntilNow[len(h.FitnessRecordBestAcceptableUntilNow)-1] {
		copy(h.BestAcceptableUntilNow, population[bestAcceptableFitnessInThisIterationIndex])
		h.FitnessRecordBestAcceptableUntilNow = append(h.FitnessRecordBestAcceptableUntilNow, bestAcceptableFitnessInThisIteration)
		h.BestAcceptableUntilNowUpdateIterations = append(h.BestAcceptableUntilNowUpdateIterations, float64(len(h.FitnessRecordIterationBestAcceptable)-1))
	}

	return newPopulation
}

// Fitness the makespan of this group is the fitness value
func (h *HAGA) Fitness(appGroupMap, cloudGroupMap map[int]struct{}, appGroup, cloudGroup []int, clouds []model.Cloud, apps []model.Application, chromosome Chromosome) float64 {
	var deployedClouds []model.Cloud = SimulateDeploy(clouds, apps, model.Solution{SchedulingResult: chromosome})
	var appsCopy []model.Application = model.AppsCopy(apps)

	appsCopy = h.calculateTaskTime(appGroupMap, cloudGroupMap, appGroup, cloudGroup, deployedClouds, appsCopy, chromosome)

	var fitnessValue float64
	for i := 0; i < len(appGroup); i++ {
		fitnessValue += h.fitnessOneApp(deployedClouds, appsCopy, appGroup[i], chromosome)
	}
	return fitnessValue
}

func (h *HAGA) fitnessOneApp(clouds []model.Cloud, apps []model.Application, appIdx int, chromosome Chromosome) float64 {
	if chromosome[appIdx] == len(clouds) || !apps[appIdx].IsTask {
		return 0
	}
	thisFitness := h.RejectExecTime - apps[appIdx].TaskCompletionTime
	if thisFitness < 0 {
		thisFitness = 0
	}
	return thisFitness
}

// HAGA only consider executing time, do not consider other times, and only consider tasks, do not consider services
func (h *HAGA) calculateTaskTime(appGroupMap, cloudGroupMap map[int]struct{}, appGroup, cloudGroup []int, clouds []model.Cloud, apps []model.Application, chromosome Chromosome) []model.Application {
	// initialization
	for i := 0; i < len(clouds); i++ {
		clouds[i].TotalTaskComplTime = 0
		clouds[i].TmpAlloc.CPU.LogicalCores = clouds[i].Allocatable.CPU.LogicalCores
	}

	// save the original order of apps
	unorderedApps := model.AppsCopy(apps)
	for cloudIndex := 0; cloudIndex < len(clouds); cloudIndex++ {
		if _, cloudChosen := cloudGroupMap[cloudIndex]; !cloudChosen {
			continue
		}
		for i := 0; i < len(clouds[cloudIndex].RunningApps); i++ {
			appIdx := clouds[cloudIndex].RunningApps[i].AppIdx
			if _, appChosen := appGroupMap[appIdx]; !appChosen || !apps[appIdx].IsTask {
				continue
			}
			latestStartTime := clouds[cloudIndex].TotalTaskComplTime

			clouds[cloudIndex].TotalTaskComplTime = latestStartTime
			clouds[cloudIndex].RunningApps[i].StartTime = latestStartTime
			unorderedApps[appIdx].StartTime = latestStartTime

			execTime := clouds[cloudIndex].RunningApps[i].TaskReq.CPUCycle / (clouds[cloudIndex].TmpAlloc.CPU.LogicalCores * clouds[cloudIndex].TmpAlloc.CPU.BaseClock * 1024 * 1024 * 1024) // unit: second

			clouds[cloudIndex].TotalTaskComplTime += execTime
			clouds[cloudIndex].RunningApps[i].TaskCompletionTime = clouds[cloudIndex].TotalTaskComplTime
			unorderedApps[appIdx].TaskCompletionTime = clouds[cloudIndex].TotalTaskComplTime

		}
	}
	// restore
	for i := 0; i < len(clouds); i++ {
		clouds[i].TmpAlloc.CPU.LogicalCores = clouds[i].Allocatable.CPU.LogicalCores
	}
	return unorderedApps
}

func (h *HAGA) initRejectFitness(appGroupMap, cloudGroupMap map[int]struct{}, appGroup, cloudGroup []int, clouds []model.Cloud, apps []model.Application, initPopulation Population) {
	var nonZeroChroNum int
	var complTimes []float64
	for i := 0; i < len(initPopulation); i++ {
		tmpClouds := model.CloudsCopy(clouds)
		tmpApps := model.AppsCopy(apps)
		tmpSolution := model.SolutionCopy(model.Solution{SchedulingResult: initPopulation[i]})

		tmpClouds = SimulateDeploy(tmpClouds, tmpApps, tmpSolution)
		tmpApps = h.calculateTaskTime(appGroupMap, cloudGroupMap, appGroup, cloudGroup, tmpClouds, tmpApps, tmpSolution.SchedulingResult)

		for j := 0; j < len(tmpClouds); j++ {
			if tmpClouds[j].TotalTaskComplTime > 0 {
				complTimes = append(complTimes, tmpClouds[j].TotalTaskComplTime)
				nonZeroChroNum++
			}

		}

	}

	if nonZeroChroNum == 0 {
		log.Println("nonZeroChroNum is 0")
		h.RejectExecTime = 0
		return
	}

	sort.Float64s(complTimes)

	// after some tests these values of the parameters seem good
	// high acceptance
	var cutRate float64 = 0.1
	var longestPara float64 = 1.5
	var averagePara float64 = 4
	var timesPara float64 = 2 // to give difference between rejection and long execution time

	var start, end int = int(float64(nonZeroChroNum) * cutRate), int(float64(nonZeroChroNum) * (1 - cutRate))
	var totalMidRange float64 = 0
	for i := start; i <= end; i++ {
		totalMidRange += complTimes[i]
	}
	var midAve float64 = totalMidRange / float64(end-start+1)
	h.RejectExecTime = timesPara * math.Min(complTimes[end]*longestPara, midAve*averagePara) // double insurance, in case that either of them is abnormal
}

func (h *HAGA) crossoverOperator(appGroupMap, cloudGroupMap map[int]struct{}, appGroup, cloudGroup []int, clouds []model.Cloud, apps []model.Application, population Population) Population {
	// avoid changing the original population, maybe not needed but for security
	var copyPopulation Population = PopulationCopy(population)

	// traverse all chromosomes in this population, use random to judge whether a chromosome needs crossover
	var indexesNeedCrossover []int
	for i := 0; i < len(copyPopulation); i++ {
		if random.RandomFloat64(0, 1) < h.CrossoverProbability {
			indexesNeedCrossover = append(indexesNeedCrossover, i)
		}
	}

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

		newFirstChromosome, newSecondChromosome := TwoPointCrossOver(firstChromosome, secondChromosome)

		// append the two new chromosomes in newPopulation
		newPopulation = append(newPopulation, newFirstChromosome, newSecondChromosome)
	}

	// directly put the chromosomes with no crossover to the new population
	for i := 0; i < len(copyPopulation); i++ {
		if !whetherCrossover[i] {
			newPopulation = append(newPopulation, copyPopulation[i])
		}
	}

	return newPopulation

}

func (h *HAGA) mutationOperator(appGroupMap, cloudGroupMap map[int]struct{}, appGroup, cloudGroup []int, clouds []model.Cloud, apps []model.Application, population Population) Population {
	var copyPopulation Population = PopulationCopy(population)
	for i := 0; i < len(copyPopulation); i++ {
		if random.RandomFloat64(0, 1) < h.MutationProbability {
			// the description in the paper is not clear, so I implement the mutation to my understanding
			pos1 := appGroup[h.MutationStartingPos]
			pos2 := appGroup[h.MutationEndingPos]
			copyPopulation[i][pos1], copyPopulation[i][pos2] = copyPopulation[i][pos2], copyPopulation[i][pos1]
		}

		fixDependence(clouds, apps, copyPopulation[i])
		// After mutation, if the chromosome becomes unacceptable, we discard it, and randomly generate a new acceptable one
		// This is to control the population mutate to good direction
		if !Acceptable(clouds, apps, copyPopulation[i]) {
			copyPopulation[i] = h.randomFitSchedule(appGroupMap, cloudGroupMap, appGroup, cloudGroup, clouds, apps)
		}
	}
	return copyPopulation
}
