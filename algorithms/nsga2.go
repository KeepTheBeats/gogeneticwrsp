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

type NSGAII struct {
	ChromosomesCount       int
	IterationCount         int
	BestUntilNow           Chromosome
	BestAcceptableUntilNow Chromosome
	CrossoverProbability   float64 // according to the paper, Crossover Probability is 1.0
	MutationProbability    float64
	StopNoUpdateIteration  int

	FitnessRecordIterationBest   []float64
	FitnessRecordBestUntilNow    []float64
	BestUntilNowUpdateIterations []float64

	FitnessRecordIterationBestAcceptable   []float64
	FitnessRecordBestAcceptableUntilNow    []float64
	BestAcceptableUntilNowUpdateIterations []float64

	SelectableCloudsForApps [][]int

	RejectRepairTime      float64 // We set this as the RepairTime of rejected applications
	RejectLatencyOverhead float64 // We set this as the LatencyOverhead of rejected applications
}

func NewNSGAII(chromosomesCount int, iterationCount int, crossoverProbability float64, mutationProbability float64, stopNoUpdateIteration int, clouds []model.Cloud, apps []model.Application) *NSGAII {
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

	return &NSGAII{
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
	}
}

func (n *NSGAII) Schedule(clouds []model.Cloud, apps []model.Application) (model.Solution, error) {
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

	// initialize a population
	var initPopulation Population = n.initialize(clouds, apps)
	//for i, chromosome := range initPopulation {
	//	log.Println(i, chromosome, len(chromosome))
	//}

	n.initRejectFitness(clouds, apps, initPopulation)
	log.Println("n.RejectRepairTime:", n.RejectRepairTime, "n.RejectLatencyOverhead:", n.RejectLatencyOverhead)

	currentPopulation := n.selectionOperator(clouds, apps, initPopulation) // Iteration No. 0

	// No. 1 iteration to No. g.IterationCount iteration
	for iteration := 1; iteration <= n.IterationCount; iteration++ {
		//log.Printf("---crossover in iteration %d-------\n", iteration)
		currentPopulation = n.crossoverOperator(clouds, apps, currentPopulation)
		//for i, chromosome := range currentPopulation {
		//	log.Println(i, chromosome)
		//}

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

		//log.Printf("--------mutation in iteration %d-------\n", iteration)
		currentPopulation = n.mutationOperator(clouds, apps, currentPopulation)
		//for i, chromosome := range currentPopulation {
		//	log.Println(i, chromosome)
		//}
		//log.Printf("--------selection in iteration %d-------\n", iteration)
		currentPopulation = n.selectionOperator(clouds, apps, currentPopulation)
		//for i, chromosome := range currentPopulation {
		//	log.Println(i, chromosome)
		//}

		clouds1 := model.CloudsCopy(clouds)
		apps1 := model.AppsCopy(apps)
		solution1 := model.SolutionCopy(model.Solution{
			SchedulingResult: n.BestAcceptableUntilNow,
		})
		//log.Println(Acceptable(clouds1, apps1, solution1.SchedulingResult))
		if !Acceptable(clouds1, apps1, solution1.SchedulingResult) {
			log.Panicln()
		}

		// if at least one acceptable solution has been found, and if the best fitness until now has not been updated for a certain number of iterations, we think that the solution is already stable enough, and stop the algorithm
		if len(n.BestAcceptableUntilNowUpdateIterations) > 1 && float64(iteration)-n.BestUntilNowUpdateIterations[len(n.BestUntilNowUpdateIterations)-1] > float64(n.StopNoUpdateIteration) {
			break
		}
	}

	if len(n.BestAcceptableUntilNowUpdateIterations) == 1 {
		return model.Solution{}, fmt.Errorf("no acceptable solution is found in %d iterations", n.IterationCount)
	}

	return model.Solution{SchedulingResult: n.BestAcceptableUntilNow}, nil
}

func (n *NSGAII) initialize(clouds []model.Cloud, apps []model.Application) Population {
	var initPopulation Population
	// in a population, there are g.ChromosomesCount chromosomes (individuals)
	for i := 0; i < n.ChromosomesCount; i++ {
		var chromosome Chromosome = RandomFitSchedule(clouds, apps)
		initPopulation = append(initPopulation, chromosome)
	}
	return initPopulation
}

func (n *NSGAII) averageFitness(clouds []model.Cloud, apps []model.Application, population Population) float64 {
	var sumFitness float64
	for _, chromosome := range population {
		sumFitness += n.Fitness(clouds, apps, chromosome).PrintFitness()
	}
	return sumFitness / float64(len(population))
}

func (n *NSGAII) selectionOperator(clouds []model.Cloud, apps []model.Application, population Population) Population {
	fitnesses := make([]NSGAIIFitness, len(population))
	for i, chromosome := range population {
		fitnesses[i] = n.Fitness(clouds, apps, chromosome)
	}

	tmpForPick := make([]int, len(fitnesses))
	var newPopulation Population = make(Population, len(population))
	var bestFitnessInThisIteration float64 = -1
	var bestFitnessInThisIterationIndex int
	var bestAcceptableFitnessInThisIteration float64 = -1
	var bestAcceptableFitnessInThisIterationIndex int

	for i := 0; i < n.ChromosomesCount; i++ {
		picked := random.RandomPickN(tmpForPick, 2)
		var selectedChromosomeIndex int
		if fitnesses[picked[0]].NfLess(fitnesses[picked[1]]) {
			selectedChromosomeIndex = picked[0]
		} else {
			selectedChromosomeIndex = picked[1]
		}
		newChromosome := make(Chromosome, len(population[selectedChromosomeIndex]))
		copy(newChromosome, population[selectedChromosomeIndex])
		newPopulation[i] = newChromosome

		chosenFitness := n.Fitness(clouds, apps, newChromosome).PrintFitness()

		if bestFitnessInThisIteration < 0 || chosenFitness < bestFitnessInThisIteration {
			bestFitnessInThisIteration = chosenFitness
			bestFitnessInThisIterationIndex = selectedChromosomeIndex
			if Acceptable(clouds, apps, population[selectedChromosomeIndex]) {
				bestAcceptableFitnessInThisIteration = chosenFitness
				bestAcceptableFitnessInThisIterationIndex = selectedChromosomeIndex
			}
		}

	}

	// acceptable and non-acceptable
	n.FitnessRecordIterationBest = append(n.FitnessRecordIterationBest, bestFitnessInThisIteration)
	if n.FitnessRecordBestUntilNow[len(n.FitnessRecordBestUntilNow)-1] < 0 || bestFitnessInThisIteration < n.FitnessRecordBestUntilNow[len(n.FitnessRecordBestUntilNow)-1] {
		copy(n.BestUntilNow, population[bestFitnessInThisIterationIndex])
		n.FitnessRecordBestUntilNow = append(n.FitnessRecordBestUntilNow, bestFitnessInThisIteration)
		n.BestUntilNowUpdateIterations = append(n.BestUntilNowUpdateIterations, float64(len(n.FitnessRecordIterationBest)-1))
	}

	// only acceptable
	n.FitnessRecordIterationBestAcceptable = append(n.FitnessRecordIterationBestAcceptable, bestAcceptableFitnessInThisIteration)
	if bestAcceptableFitnessInThisIteration >= 0 && (n.FitnessRecordBestAcceptableUntilNow[len(n.FitnessRecordBestAcceptableUntilNow)-1] < 0 || bestAcceptableFitnessInThisIteration < n.FitnessRecordBestAcceptableUntilNow[len(n.FitnessRecordBestAcceptableUntilNow)-1]) {
		copy(n.BestAcceptableUntilNow, population[bestAcceptableFitnessInThisIterationIndex])
		n.FitnessRecordBestAcceptableUntilNow = append(n.FitnessRecordBestAcceptableUntilNow, bestAcceptableFitnessInThisIteration)
		n.BestAcceptableUntilNowUpdateIterations = append(n.BestAcceptableUntilNowUpdateIterations, float64(len(n.FitnessRecordIterationBestAcceptable)-1))
	}
	//log.Println("Average New Population:", n.averageFitness(clouds, apps, newPopulation))

	return newPopulation
}

type NSGAIIFitness struct {
	RepairTime      float64
	LatencyOverhead float64
}

// uniform weighted sum of the paper
func (nf NSGAIIFitness) PrintFitness() float64 {
	return 0.5*0.5*nf.RepairTime + 0.5*0.5*nf.LatencyOverhead
}

func (nf NSGAIIFitness) NfLess(cmp NSGAIIFitness) bool {
	return nf.PrintFitness() < cmp.PrintFitness()
}

type NSGAIIFitnessSlice []NSGAIIFitness

func (nfs NSGAIIFitnessSlice) Len() int {
	return len(nfs)
}

func (nfs NSGAIIFitnessSlice) Swap(i, j int) {
	nfs[i], nfs[j] = nfs[j], nfs[i]
}

func (nfs NSGAIIFitnessSlice) Less(i, j int) bool {
	return nfs[i].NfLess(nfs[j])
}

func NSGAIICalcStartComplTime(clouds []model.Cloud, apps []model.Application, chromosome Chromosome) []model.Application {
	// initialization
	for i := 0; i < len(clouds); i++ {
		clouds[i].TotalTaskComplTime = 0
		clouds[i].TmpAlloc.CPU.LogicalCores = clouds[i].Allocatable.CPU.LogicalCores
	}
	// save the original order of apps
	unorderedApps := model.AppsCopy(apps)

	for cloudIndex := 0; cloudIndex < len(clouds); cloudIndex++ {
		for i := 0; i < len(clouds[cloudIndex].RunningApps); i++ {
			appIdx := clouds[cloudIndex].RunningApps[i].AppIdx

			latestStartTime := clouds[cloudIndex].TotalTaskComplTime

			clouds[cloudIndex].TotalTaskComplTime = latestStartTime
			clouds[cloudIndex].RunningApps[i].StartTime = latestStartTime
			unorderedApps[appIdx].StartTime = latestStartTime

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
			unorderedApps[appIdx].ImagePullDoneTime = unorderedApps[appIdx].StartTime + imagePullTime

			// set data input done time
			clouds[cloudIndex].RunningApps[i].DataInputDoneTime = clouds[cloudIndex].RunningApps[i].StartTime + imagePullTime + dataInputTime
			unorderedApps[appIdx].DataInputDoneTime = unorderedApps[appIdx].StartTime + imagePullTime + dataInputTime

			// set stable time
			clouds[cloudIndex].RunningApps[i].StableTime = clouds[cloudIndex].RunningApps[i].StartTime + imagePullTime + dataInputTime + startUpTime
			unorderedApps[appIdx].StableTime = unorderedApps[appIdx].StartTime + imagePullTime + dataInputTime + startUpTime

			if !clouds[cloudIndex].RunningApps[i].IsTask { // Services take up the resource
				// take up cpu
				clouds[cloudIndex].TmpAlloc.CPU.LogicalCores -= clouds[cloudIndex].RunningApps[i].SvcReq.CPUClock / clouds[cloudIndex].TmpAlloc.CPU.BaseClock

			}
			// NSGAII consume the 3 parts of time
			clouds[cloudIndex].TotalTaskComplTime += imagePullTime + dataInputTime + startUpTime

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

func (n *NSGAII) Fitness(clouds []model.Cloud, apps []model.Application, chromosome Chromosome) NSGAIIFitness {
	var deployedClouds []model.Cloud = SimulateDeploy(clouds, apps, model.Solution{SchedulingResult: chromosome})
	var appsCopy []model.Application = model.AppsCopy(apps)
	appsCopy = NSGAIICalcStartComplTime(deployedClouds, appsCopy, chromosome)
	//for i := 0; i < len(deployedClouds); i++ {
	//	sort.Sort(model.AppSlice(deployedClouds[i].RunningApps))
	//	fmt.Println("Cloud:", i)
	//	for j := 0; j < len(deployedClouds[i].RunningApps); j++ {
	//		fmt.Println(deployedClouds[i].RunningApps[j].IsTask, deployedClouds[i].RunningApps[j].StartTime, deployedClouds[i].RunningApps[j].TaskCompletionTime)
	//	}
	//}
	//time.Sleep(101 * time.Second)

	var repairTime, latencyOverhead float64

	// the fitnessValue is based on each application
	for appIndex := 0; appIndex < len(chromosome); appIndex++ {
		thirRepairTime, thisLatencyOverhead := n.fitnessOneApp(deployedClouds, appsCopy, appIndex, chromosome)
		repairTime += thirRepairTime
		latencyOverhead += thisLatencyOverhead
	}

	return NSGAIIFitness{repairTime, latencyOverhead}
}

func (n *NSGAII) fitnessOneApp(clouds []model.Cloud, apps []model.Application, appIdx int, chromosome Chromosome) (float64, float64) {
	if chromosome[appIdx] == len(clouds) || apps[appIdx].IsTask { // NSGA-II only considers services
		return n.RejectRepairTime, n.RejectLatencyOverhead
	}
	var repairTime float64 = apps[appIdx].StableTime - (apps[appIdx].DataInputDoneTime - apps[appIdx].ImagePullDoneTime) // NSGA-II does not consider data input time
	var latencyOverhead float64
	for i := 0; i < len(apps[appIdx].Depend); i++ {
		depIdx := apps[appIdx].Depend[i].AppIdx
		latencyOverhead += clouds[chromosome[appIdx]].Allocatable.NetCondClouds[chromosome[depIdx]].RTT
	}
	// reject ones should have the biggest fitness
	if repairTime > n.RejectRepairTime {
		repairTime = n.RejectRepairTime
	}
	if latencyOverhead > n.RejectLatencyOverhead {
		latencyOverhead = n.RejectLatencyOverhead
	}
	return repairTime, latencyOverhead
}

// used for initialization of RejectRepairTime and RejectLatencyOverhead
func (n *NSGAII) fitnessOneAppInit(clouds []model.Cloud, apps []model.Application, appIdx int, chromosome Chromosome) (float64, float64) {
	if chromosome[appIdx] == len(clouds) {
		return -1, -1
	}
	var repairTime float64 = apps[appIdx].StableTime - (apps[appIdx].DataInputDoneTime - apps[appIdx].ImagePullDoneTime) // NSGA-II does not consider data input time
	var latencyOverhead float64
	for i := 0; i < len(apps[appIdx].Depend); i++ {
		depIdx := apps[appIdx].Depend[i].AppIdx
		latencyOverhead += clouds[chromosome[appIdx]].Allocatable.NetCondClouds[chromosome[depIdx]].RTT
	}
	return repairTime, latencyOverhead
}

func (n *NSGAII) initRejectFitness(clouds []model.Cloud, apps []model.Application, initPopulation Population) {
	var repairTimes, latencyOverheads []float64
	for _, chromosome := range initPopulation {
		deployedClouds := SimulateDeploy(clouds, apps, model.Solution{SchedulingResult: chromosome})
		appsCopy := model.AppsCopy(apps)
		appsCopy = NSGAIICalcStartComplTime(deployedClouds, appsCopy, chromosome)
		for i := 0; i < len(apps); i++ {
			thisRepairTime, thisLatencyOverhead := n.fitnessOneAppInit(deployedClouds, appsCopy, i, chromosome)
			if thisRepairTime >= 0 {
				repairTimes = append(repairTimes, thisRepairTime)
				latencyOverheads = append(latencyOverheads, thisLatencyOverhead)
			}
		}
	}

	sort.Float64s(repairTimes)
	sort.Float64s(latencyOverheads)

	var cutRate float64 = 0.1
	var longestPara float64 = 1.5
	var averagePara float64 = 4
	var timesPara float64 = 2 // to give difference between rejection and high fitness

	func() {
		var start, end int = int(float64(len(repairTimes)) * cutRate), int(float64(len(repairTimes)) * (1 - cutRate))
		var totalMidRange float64 = 0
		for i := start; i <= end; i++ {
			totalMidRange += repairTimes[i]
		}
		var midAve float64 = totalMidRange / float64(end-start+1)
		n.RejectRepairTime = timesPara * math.Min(repairTimes[end]*longestPara, midAve*averagePara) // double insurance, in case that either of them is abnormal
	}()

	func() {
		var start, end int = int(float64(len(latencyOverheads)) * cutRate), int(float64(len(latencyOverheads)) * (1 - cutRate))
		var totalMidRange float64 = 0
		for i := start; i <= end; i++ {
			totalMidRange += latencyOverheads[i]
		}
		var midAve float64 = totalMidRange / float64(end-start+1)
		n.RejectLatencyOverhead = timesPara * math.Min(latencyOverheads[end]*longestPara, midAve*averagePara) // double insurance, in case that either of them is abnormal
	}()

}

//
func (n *NSGAII) crossoverOperator(clouds []model.Cloud, apps []model.Application, population Population) Population {
	if len(apps) <= 1 { // only with at least 2 genes in a chromosome, can we do crossover
		return population
	}
	// avoid changing the original population, maybe not needed but for security
	var copyPopulation Population = PopulationCopy(population)

	// traverse all chromosomes in this population, use random to judge whether a chromosome needs crossover
	var indexesNeedCrossover []int
	for i := 0; i < len(copyPopulation); i++ {
		if random.RandomFloat64(0, 1) < n.CrossoverProbability {
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

		firstNew, secondNew := TwoPointCrossOver(firstChromosome, secondChromosome)

		// append the two new chromosomes in newPopulation
		newPopulation = append(newPopulation, firstNew, secondNew)

	}

	// directly put the chromosomes with no crossover to the new population
	for i := 0; i < len(copyPopulation); i++ {
		if !whetherCrossover[i] {
			newPopulation = append(newPopulation, copyPopulation[i])
		}
	}

	return newPopulation
}

// according to the paper, randomly regenerate scheduling schemes
func (n *NSGAII) mutationOperator(clouds []model.Cloud, apps []model.Application, population Population) Population {
	// avoid changing the original population, maybe not needed but for security
	var copyPopulation Population = PopulationCopy(population)
	for i := 0; i < len(copyPopulation); i++ {
		if random.RandomFloat64(0, 1) < n.MutationProbability {
			copyPopulation[i] = RandomFitSchedule(clouds, apps)
		}
	}
	return copyPopulation
}

// DrawChart draw g.FitnessRecordIterationBest and g.FitnessRecordBestUntilNow on a line chart
func (n *NSGAII) DrawChart() {
	var drawChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
		var xValuesIterationBest []float64
		for i, _ := range n.FitnessRecordIterationBest {
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
					YValues: n.FitnessRecordIterationBest,
				},
				chart.ContinuousSeries{
					Name: "Best Fitness in all iterations",
					// the first value (iteration -1) is much different with others, which will cause that we cannot observer the trend of the evolution
					XValues: n.BestUntilNowUpdateIterations[1:],
					YValues: n.FitnessRecordBestUntilNow[1:],
					Style: chart.Style{
						Show:            true,
						StrokeDashArray: []float64{5.0, 3.0, 2.0, 3.0},
						StrokeWidth:     1,
					},
				},
				chart.ContinuousSeries{
					Name:    "Best Acceptable Fitness in each iterations",
					XValues: xValuesIterationBest,
					YValues: n.FitnessRecordIterationBestAcceptable,
					Style: chart.Style{
						Show:            true,
						StrokeDashArray: []float64{2.0, 3.0},
						StrokeWidth:     1,
					},
				},
				chart.ContinuousSeries{
					Name:    "Best Acceptable Fitness in all iterations",
					XValues: n.BestAcceptableUntilNowUpdateIterations[1:],
					YValues: n.FitnessRecordBestAcceptableUntilNow[1:],
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
