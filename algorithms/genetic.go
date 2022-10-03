package algorithms

import (
	"log"
	"math"
	"sort"

	"github.com/KeepTheBeats/routing-algorithms/random"

	"gogeneticwrsp/model"
)

type Chromosome []int
type Population []Chromosome

type Genetic struct {
	ChromosomesCount             int
	IterationCount               int
	BestUntilNow                 Chromosome
	CrossoverProbability         float64
	MutationProbability          float64
	FitnessRecordIterationBest   []float64
	FitnessRecordBestUntilNow    []float64
	BestUntilNowUpdateIterations []float64
}

func NewGenetic(chromosomesCount int, iterationCount int, crossoverProbability float64, mutationProbability float64, clouds []model.Cloud, apps []model.Application) Genetic {
	initBestUntilNow := make(Chromosome, len(apps))
	return Genetic{
		ChromosomesCount:             chromosomesCount,
		IterationCount:               iterationCount,
		BestUntilNow:                 initBestUntilNow,
		CrossoverProbability:         crossoverProbability,
		MutationProbability:          mutationProbability,
		FitnessRecordBestUntilNow:    []float64{Fitness(clouds, apps, initBestUntilNow, false)},
		BestUntilNowUpdateIterations: []float64{-1}, // We define that the first BestUntilNow is set in the No. -1 iteration
	}
}

// Fitness calculate the fitness value of this scheduling result
// There are 2 stages in our algorithm.
// 1. not strict, to find the big area of the solution, not strict to avoid the local optimal solutions;
// 2. strict, to find the optimal solution in the area given by stage 1, strict to choose better solutions.
func Fitness(clouds []model.Cloud, apps []model.Application, chromosome Chromosome, strict bool) float64 {
	var deployedClouds []model.Cloud = SimulateDeploy(clouds, apps, model.Solution{SchedulingResult: chromosome})
	var fitnessValue float64

	// the fitnessValue is based on each application
	for appIndex := 0; appIndex < len(chromosome); appIndex++ {
		fitnessValue += fitnessOneApp(deployedClouds, apps[appIndex], chromosome[appIndex], strict)
	}

	return fitnessValue
}

// fitness of a single application
func fitnessOneApp(clouds []model.Cloud, app model.Application, chosenCloudIndex int, strict bool) float64 {
	var cpuFitness float64 // fitness about CPU
	if clouds[chosenCloudIndex].Allocatable.CPU >= 0 {
		cpuFitness = 1
	} else { // CPU is compressible resource in Kubernetes
		cpuFitness = clouds[chosenCloudIndex].Capacity.CPU / ((0 - clouds[chosenCloudIndex].Allocatable.CPU) + clouds[chosenCloudIndex].Capacity.CPU)
	}

	var memoryFitness float64 // fitness about memory
	if clouds[chosenCloudIndex].Allocatable.Memory >= 0 {
		memoryFitness = 1
	} else { // Memory is incompressible resource in Kubernetes
		memoryFitness = 0
	}

	var storageFitness float64 // fitness about storage
	if clouds[chosenCloudIndex].Allocatable.Storage >= 0 {
		storageFitness = 1
	} else { // Storage is incompressible resource in Kubernetes
		storageFitness = 0
	}

	var netLatencyFitness float64 // fitness about network latency
	if clouds[chosenCloudIndex].Allocatable.NetLatency <= app.Requests.NetLatency {
		netLatencyFitness = 1
	} else {
		netLatencyFitness = app.Requests.NetLatency / clouds[chosenCloudIndex].Allocatable.NetLatency
	}

	// if incompressible resources cannot be met, in strict strategy, we set the fitness as 0.
	if strict && (memoryFitness == 0 || storageFitness == 0) {
		return 0
	}

	// the higher priority, the more important the application, the more its fitness should be scaled up
	return (cpuFitness + memoryFitness + storageFitness + netLatencyFitness) * float64(app.Priority)
}

func (g *Genetic) initialize(clouds []model.Cloud, apps []model.Application) Population {
	var initPopulation Population
	// in a population, there are g.ChromosomesCount chromosomes (individuals)
	for i := 0; i < g.ChromosomesCount; i++ {
		var chromosome Chromosome
		// in a chromosome, there are len(apps) genes. the value range of each gene is [0, len(clouds)-1]
		for j := 0; j < len(apps); j++ {
			gene := random.RandomInt(0, len(clouds)-1)
			chromosome = append(chromosome, gene)
		}
		initPopulation = append(initPopulation, chromosome)
	}
	return initPopulation
}

func (g *Genetic) selectionOperator(clouds []model.Cloud, apps []model.Application, population Population, strict bool) Population {
	fitnesses := make([]float64, len(population))
	cumulativeFitnesses := make([]float64, len(population))

	// calculate the fitness of each chromosome in this population
	var maxFitness, minFitness float64 = 0, math.MaxFloat64 // record the max and min for standardization
	for i := 0; i < len(population); i++ {
		fitness := Fitness(clouds, apps, population[i], strict)
		fitnesses[i] = fitness
		if fitness > maxFitness {
			maxFitness = fitness
		}
		if fitness < minFitness {
			minFitness = fitness
		}
	}

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

	for i := 0; i < len(fitnesses); i++ {
		fitnesses[i] += offset
	}
	//log.Println("Standardized fitness", fitnesses)

	// calculate the cumulative fitnesses of the chromosomes in this population
	cumulativeFitnesses[0] = fitnesses[0]
	for i := 1; i < len(population); i++ {
		cumulativeFitnesses[i] = cumulativeFitnesses[i-1] + fitnesses[i]
	}

	// select g.ChromosomesCount chromosomes to generate a new population
	var newPopulation Population
	var bestFitnessInThisIteration float64
	var bestFitnessInThisIterationIndex int
	for i := 0; i < g.ChromosomesCount; i++ {

		// roulette-wheel selection
		// selectionThreshold is a random float64 in [0, biggest cumulative fitness])
		selectionThreshold := random.RandomFloat64(0, cumulativeFitnesses[len(cumulativeFitnesses)-1])
		// selected the smallest cumulativeFitnesses that is begger than selectionThreshold
		selectedChromosomeIndex := sort.SearchFloat64s(cumulativeFitnesses, selectionThreshold)

		// cannot directly append population[selectedChromosomeIndex], otherwise, the new population may be changed by strange reasons
		newChromosome := make(Chromosome, len(population[selectedChromosomeIndex]))
		copy(newChromosome, population[selectedChromosomeIndex])
		newPopulation = append(newPopulation, newChromosome)

		// record the best fitness in this iteration
		chosenFitness := Fitness(clouds, apps, population[selectedChromosomeIndex], strict)
		if chosenFitness > bestFitnessInThisIteration {
			bestFitnessInThisIteration = chosenFitness
			bestFitnessInThisIterationIndex = selectedChromosomeIndex
		}
	}

	if Fitness(clouds, apps, g.BestUntilNow, strict) != g.FitnessRecordBestUntilNow[len(g.FitnessRecordBestUntilNow)-1] {
		log.Println("not equal:", g.FitnessRecordBestUntilNow[len(g.FitnessRecordBestUntilNow)-1], Fitness(clouds, apps, g.BestUntilNow, strict))
	}

	// record:
	// the best chromosome until now;
	// the best fitness record;
	// the best fitness in every iteration;
	// the iterations during which the g.BestUntilNow is updated;

	// the best fitness in every iteration;
	g.FitnessRecordIterationBest = append(g.FitnessRecordIterationBest, bestFitnessInThisIteration)
	if bestFitnessInThisIteration > Fitness(clouds, apps, g.BestUntilNow, strict) {

		log.Printf("In iteration %d, update g.FitnessRecordBestUntilNow from %f to %f, update g.BestUntilNow from %v, to %v\n", len(g.FitnessRecordIterationBest)-1, g.FitnessRecordBestUntilNow[len(g.FitnessRecordBestUntilNow)-1], bestFitnessInThisIteration, g.BestUntilNow, population[bestFitnessInThisIterationIndex])
		// the best chromosome until now;
		// if we directly use "=", g.BestUntilNow may be changed by strange reasons.
		copy(g.BestUntilNow, population[bestFitnessInThisIterationIndex])
		// the best fitness record;
		g.FitnessRecordBestUntilNow = append(g.FitnessRecordBestUntilNow, bestFitnessInThisIteration)
		// the iteration during which the g.BestUntilNow is updated
		// len(g.FitnessRecordIterationBest)-1 is the current iteration number, there are IterationCount+1 iterations in total
		g.BestUntilNowUpdateIterations = append(g.BestUntilNowUpdateIterations, float64(len(g.FitnessRecordIterationBest)-1))
	}

	return newPopulation
}

func (g *Genetic) crossoverOperator(clouds []model.Cloud, apps []model.Application, population Population, strict bool) Population {
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

	log.Println("indexesNeedCrossover:", indexesNeedCrossover)

	var newPopulation Population

	// randomly choose pairs of chromosomes to do crossover
	var whetherCrossover []bool = make([]bool, len(population))
	for len(indexesNeedCrossover) > 1 { // if len(indexesNeedCrossover) <= 1, stop crossover
		// choose two indexes for crossover;
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

	log.Println("whetherCrossover:", whetherCrossover)

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
				var newGene int = random.RandomInt(0, len(clouds)-1)
				// make sure that the mutated gene is different with the original one
				for newGene == copyPopulation[i][j] {
					newGene = random.RandomInt(0, len(clouds)-1)
				}
				//log.Printf("gene [%d][%d] mutates from %d to %d\n", i, j, copyPopulation[i][j], newGene)
				copyPopulation[i][j] = newGene
			}
		}
	}

	return copyPopulation
}

func (g *Genetic) Schedule(clouds []model.Cloud, apps []model.Application) (model.Solution, error) {
	// initialize a population
	var initPopulation Population = g.initialize(clouds, apps)
	for i, chromosome := range initPopulation {
		log.Println(i, chromosome)
	}

	log.Println("---selection after initialization-------")
	// there are IterationCount+1 iterations in total, this is the No. 0 iteration
	currentPopulation := g.selectionOperator(clouds, apps, initPopulation, false) // Iteration No. 0
	//for i, chromosome := range currentPopulation {
	//	log.Println(i, chromosome)
	//}

	// No. 1 iteration to No. g.IterationCount iteration
	for iteration := 1; iteration <= g.IterationCount; iteration++ {
		log.Printf("---crossover in iteration %d-------\n", iteration)
		currentPopulation = g.crossoverOperator(clouds, apps, currentPopulation, false)
		//for i, chromosome := range currentPopulation {
		//	log.Println(i, chromosome)
		//}
		log.Printf("--------mutation in iteration %d-------\n", iteration)
		currentPopulation = g.mutationOperator(clouds, apps, currentPopulation, false)
		//for i, chromosome := range currentPopulation {
		//	log.Println(i, chromosome)
		//}
		log.Printf("--------selection in iteration %d-------\n", iteration)
		currentPopulation = g.selectionOperator(clouds, apps, currentPopulation, false)
		//for i, chromosome := range currentPopulation {
		//	log.Println(i, chromosome)
		//}
	}

	return model.Solution{SchedulingResult: g.BestUntilNow}, nil
}
