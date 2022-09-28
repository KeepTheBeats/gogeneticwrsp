package algorithms

import (
	"fmt"
	"log"
	"sort"

	"github.com/KeepTheBeats/routing-algorithms/random"

	"gogeneticwrsp/model"
)

type Chromosome []int
type Population []Chromosome

type Genetic struct {
	ChromosomesCount int
	IterationCount   int
	BestUntilNow     Chromosome
}

func NewGenetic(chromosomesCount int, iterationCount int, clouds []model.Cloud, apps []model.Application) Genetic {
	return Genetic{
		ChromosomesCount: chromosomesCount,
		IterationCount:   iterationCount,
		BestUntilNow:     make(Chromosome, len(apps)),
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
		// in a chromosome, there are len(apps) genes
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
	for i := 0; i < len(population); i++ {
		fitness := Fitness(clouds, apps, population[i], strict)
		fitnesses[i] = fitness
	}

	// calculate the cumulative fitnesses of the chromosomes in this population
	cumulativeFitnesses[0] = fitnesses[0]
	for i := 1; i < len(population); i++ {
		cumulativeFitnesses[i] = cumulativeFitnesses[i-1] + fitnesses[i]
	}

	// select g.ChromosomesCount chromosomes to generate a new population
	var newPopulation Population
	for i := 0; i < g.ChromosomesCount; i++ {

		// roulette-wheel selection
		// selectionThreshold is a random float64 in [0, biggest cumulative fitness)
		selectionThreshold := random.RandomFloat64(0, cumulativeFitnesses[len(cumulativeFitnesses)-1])
		// selected the smallest cumulativeFitnesses that is begger than selectionThreshold
		selectedChromosomeIndex := sort.SearchFloat64s(cumulativeFitnesses, selectionThreshold)

		newPopulation = append(newPopulation, population[selectedChromosomeIndex])
		// record the best chromosome until now
		if Fitness(clouds, apps, population[selectedChromosomeIndex], strict) > Fitness(clouds, apps, g.BestUntilNow, strict) {
			g.BestUntilNow = population[selectedChromosomeIndex]
		}
	}

	return newPopulation
}

func (g *Genetic) crossoverOperator(clouds []model.Cloud, apps []model.Application, population Population) Population {
	return Population{}
}

func (g *Genetic) mutationOperator(clouds []model.Cloud, apps []model.Application, population Population) Population {
	return Population{}
}

func (g *Genetic) Schedule(clouds []model.Cloud, apps []model.Application) (model.Solution, error) {
	// initialize a population
	var initPopulation Population = g.initialize(clouds, apps)
	for _, chromosome := range initPopulation {
		log.Println(chromosome)
	}

	currentPopulation := g.selectionOperator(clouds, apps, initPopulation, false)
	fmt.Println(currentPopulation)

	return model.Solution{}, nil
}
