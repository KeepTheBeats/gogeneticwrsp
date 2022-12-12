package main

import (
	"encoding/csv"
	"fmt"
	"go/build"
	"gogeneticwrsp/algorithms"
	"gogeneticwrsp/experimenttools"
	"gogeneticwrsp/model"
	"log"
	"os"
	"runtime"
)

func main() {
	// set the log to show line number and file name
	log.SetFlags(0 | log.Lshortfile)

	//log.Println("Hello World!")

	var fitnesses []float64
	for i := 0; i < 100; i++ {
		fitness := testParameters(0.4, 0.003, false, true, false)
		fitnesses = append(fitnesses, fitness)
	}

	var csvContent [][]string
	csvContent = append(csvContent, []string{"round", "fitnesses"})
	round := 1

	for i := 0; i < len(fitnesses); i++ {
		csvContent = append(csvContent, []string{fmt.Sprintf("%d", round), fmt.Sprintf("%f", fitnesses[i])})
		round++
	}

	var csvPath string
	if runtime.GOOS == "windows" {
		csvPath = fmt.Sprintf("%s\\src\\gogeneticwrsp\\experiments\\validateaverage\\validateaverage.csv", build.Default.GOPATH)
	} else {
		csvPath = fmt.Sprintf("%s/src/gogeneticwrsp/experiments/validateaverage/validateaverage.csv", build.Default.GOPATH)
	}

	f, err := os.Create(csvPath)
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

func testParameters(crossoverProbability float64, mutationProbability float64, twoPointCrossover, btSelection, cbMutation bool) float64 {
	var numCloud, numApp int = 5, 40
	var appSuffix string = "0"

	// generate clouds and apps, and write to files
	//experimenttools.GenerateClouds(numCloud)
	//experimenttools.GenerateApps(numApp, appSuffix, 0.5)

	// read clouds and apps from files
	var clouds []model.Cloud
	var apps []model.Application

	clouds = experimenttools.ReadClouds(numCloud)
	apps = experimenttools.ReadApps(numApp, appSuffix)

	var crossoverOperator func(algorithms.Chromosome, algorithms.Chromosome) (algorithms.Chromosome, algorithms.Chromosome)
	if twoPointCrossover {
		crossoverOperator = algorithms.TwoPointCrossOver
	} else {
		crossoverOperator = algorithms.OnePointCrossOver
	}

	geneticAlgorithm := algorithms.NewGenetic(100, 5000, crossoverProbability, mutationProbability, 100, algorithms.RandomFitSchedule, crossoverOperator, btSelection, cbMutation, clouds, apps)

	_, err := geneticAlgorithm.Schedule(clouds, apps)
	if err != nil {
		log.Panicf("geneticAlgorithm.Schedule(clouds, apps), error: %s", err.Error())
	}

	//log.Println()
	if len(geneticAlgorithm.FitnessRecordBestUntilNow) != len(geneticAlgorithm.BestUntilNowUpdateIterations) {
		log.Panicf("len(geneticAlgorithm.FitnessRecordBestUntilNow): %d, len(geneticAlgorithm.BestUntilNowUpdateIterations): %d\n", len(geneticAlgorithm.FitnessRecordBestUntilNow), len(geneticAlgorithm.BestUntilNowUpdateIterations))
	}

	i := len(geneticAlgorithm.FitnessRecordBestAcceptableUntilNow) - 1
	//log.Printf("Iteration %d: FitnessRecordBestAcceptableUntilNow: %f\n", int(geneticAlgorithm.BestAcceptableUntilNowUpdateIterations[i]), geneticAlgorithm.FitnessRecordBestAcceptableUntilNow[i])
	//
	//log.Println("solution:", solution)
	return geneticAlgorithm.FitnessRecordBestAcceptableUntilNow[i]

}
