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
}

func NewGenetic(chromosomesCount int, iterationCount int, crossoverProbability float64, mutationProbability float64, stopNoUpdateIteration int, initFunc func([]model.Cloud, []model.Application) []int, clouds []model.Cloud, apps []model.Application) *Genetic {

	selectableCloudsForApps := make([][]int, len(apps))
	for i := 0; i < len(apps); i++ {
		selectableCloudsForApps[i] = append(selectableCloudsForApps[i], len(clouds))
		for j := 0; j < len(clouds); j++ {
			if clouds[j].Allocatable.NetLatency <= apps[i].Requests.NetLatency {
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

// Fitness calculate the fitness value of this scheduling result
// There are 2 stages in our algorithm.
// 1. not strict, to find the big area of the solution, not strict to avoid the local optimal solutions;
// 2. strict, to find the optimal solution in the area given by stage 1, strict to choose better solutions.
func Fitness(clouds []model.Cloud, apps []model.Application, chromosome Chromosome) float64 {
	var deployedClouds []model.Cloud = SimulateDeploy(clouds, apps, model.Solution{SchedulingResult: chromosome})
	var fitnessValue float64

	// the fitnessValue is based on each application
	for appIndex := 0; appIndex < len(chromosome); appIndex++ {
		fitnessValue += fitnessOneApp(deployedClouds, apps[appIndex], chromosome[appIndex])
	}

	return fitnessValue
}

// fitness of a single application
func fitnessOneApp(clouds []model.Cloud, app model.Application, chosenCloudIndex int) float64 {
	// The application is rejected, only set a small fitness, because in some aspects being rejected is better than being deployed incorrectly
	if chosenCloudIndex == len(clouds) {
		return 0
	}

	var subFitness []float64
	var overflow bool

	var cpuFitness float64 // fitness about CPU
	if clouds[chosenCloudIndex].Allocatable.CPU.LogicalCores >= 0 {
		cpuFitness = 1
	} else {
		// CPU is compressible resource in Kubernetes
		//overflowRate := clouds[chosenCloudIndex].Capacity.CPU / ((0 - clouds[chosenCloudIndex].Allocatable.CPU) + clouds[chosenCloudIndex].Capacity.CPU)
		cpuFitness = -1
		overflow = true
		// even CPU is compressible, I think we still should not allow it to overflow
	}
	subFitness = append(subFitness, cpuFitness)

	var memoryFitness float64 // fitness about memory
	if clouds[chosenCloudIndex].Allocatable.Memory >= 0 {
		memoryFitness = 1
	} else {
		// Memory is incompressible resource in Kubernetes, but the fitness is used for the selection in evolution
		// After evolution, we will only output the solution with no overflow resources
		//memoryFitness = 0
		//overflowRate := clouds[chosenCloudIndex].Capacity.Memory / ((0 - clouds[chosenCloudIndex].Allocatable.Memory) + clouds[chosenCloudIndex].Capacity.Memory)
		memoryFitness = -1
		overflow = true
	}
	subFitness = append(subFitness, memoryFitness)

	var storageFitness float64 // fitness about storage
	if clouds[chosenCloudIndex].Allocatable.Storage >= 0 {
		storageFitness = 1
	} else {
		// Storage is incompressible resource in Kubernetes, but the fitness is used for the selection in evolution
		// After evolution, we will only output the solution with no overflow resources
		//storageFitness = 0
		//overflowRate := clouds[chosenCloudIndex].Capacity.Storage / ((0 - clouds[chosenCloudIndex].Allocatable.Storage) + clouds[chosenCloudIndex].Capacity.Storage)
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
			if clouds[i].Allocatable.NetLatency > apps[undeployed[appIndex]].Requests.NetLatency {
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

func (g *Genetic) selectionOperator(clouds []model.Cloud, apps []model.Application, population Population, strict bool) Population {
	fitnesses := make([]float64, len(population))
	cumulativeFitnesses := make([]float64, len(population))

	// calculate the fitness of each chromosome in this population
	var maxFitness, minFitness float64 = 0, math.MaxFloat64 // record the max and min for standardization
	for i := 0; i < len(population); i++ {
		fitness := Fitness(clouds, apps, population[i])

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
		chosenFitness := Fitness(clouds, apps, population[selectedChromosomeIndex])
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

func (g *Genetic) crossoverOperator(clouds []model.Cloud, apps []model.Application, population Population, strict bool) Population {
	if len(apps) <= 1 { // only with at least 2 genes in a chromosome, can we do crossover
		return population
	}
	// avoid changing the original population, maybe not needed but for security
	var copyPopulation Population = make(Population, len(population))
	copy(copyPopulation, population)

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

func (g *Genetic) mutationOperator(clouds []model.Cloud, apps []model.Application, population Population, strict bool) Population {
	// avoid changing the original population, maybe not needed but for security
	var copyPopulation Population = make(Population, len(population))
	copy(copyPopulation, population)

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
		// After mutation, if the chromosome becomes unacceptable, we discard it, and randomly generate a new acceptable one
		// This is to control the population mutate to good direction
		if !Acceptable(clouds, apps, copyPopulation[i]) {
			copyPopulation[i] = RandomFitSchedule(clouds, apps)
		}
	}

	return copyPopulation
}

func (g *Genetic) Schedule(clouds []model.Cloud, apps []model.Application) (model.Solution, error) {
	// initialize a population
	var initPopulation Population = g.initialize(clouds, apps)
	for i, chromosome := range initPopulation {
		log.Println(i, chromosome, len(chromosome))
	}

	//log.Println("---selection after initialization-------")
	// there are IterationCount+1 iterations in total, this is the No. 0 iteration
	currentPopulation := g.selectionOperator(clouds, apps, initPopulation, false) // Iteration No. 0
	//for i, chromosome := range currentPopulation {
	//	log.Println(i, chromosome)
	//}

	// No. 1 iteration to No. g.IterationCount iteration
	for iteration := 1; iteration <= g.IterationCount; iteration++ {
		//log.Printf("---crossover in iteration %d-------\n", iteration)
		currentPopulation = g.crossoverOperator(clouds, apps, currentPopulation, false)
		//for i, chromosome := range currentPopulation {
		//	log.Println(i, chromosome)
		//}
		//log.Printf("--------mutation in iteration %d-------\n", iteration)
		currentPopulation = g.mutationOperator(clouds, apps, currentPopulation, false)
		//for i, chromosome := range currentPopulation {
		//	log.Println(i, chromosome)
		//}
		//log.Printf("--------selection in iteration %d-------\n", iteration)
		currentPopulation = g.selectionOperator(clouds, apps, currentPopulation, false)
		//for i, chromosome := range currentPopulation {
		//	log.Println(i, chromosome)
		//}

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
