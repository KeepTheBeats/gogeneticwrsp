package main

import (
	"gogeneticwrsp/algorithms"
	"log"

	"gogeneticwrsp/experimenttools"
	"gogeneticwrsp/model"
)

func main() {
	// set the log to show line number and file name
	log.SetFlags(0 | log.Lshortfile)

	//log.Println("Hello World!")

	var numCloud, numApp int = 10, 40

	// generate clouds and apps, and write to files
	//experimenttools.GenerateCloudsApps(numCloud, numApp)

	// read clouds and apps from files
	var clouds []model.Cloud
	var apps []model.Application
	clouds, apps = experimenttools.ReadCloudsApps(numCloud, numApp)
	//for i := 0; i < numCloud; i++ {
	//	log.Println(clouds[i])
	//}
	//
	//for i := 0; i < numApp; i++ {
	//	log.Println(apps[i])
	//}

	//geneticAlgorithm := algorithms.NewGenetic(200, 5000, 0.7, 0.01, 200, algorithms.InitializeUndeployedChromosome, clouds, apps)
	//geneticAlgorithm := algorithms.NewGenetic(100, 5000, 0.7, 0.007, 200, algorithms.InitializeAcceptableChromosome, clouds, apps)
	geneticAlgorithm := algorithms.NewGenetic(200, 5000, 0.7, 0.01, 200, algorithms.RandomFitSchedule, clouds, apps)

	solution, err := geneticAlgorithm.Schedule(clouds, apps)
	if err != nil {
		log.Printf("geneticAlgorithm.Schedule(clouds, apps), error: %s", err.Error())
	}

	for i := 0; i < len(geneticAlgorithm.FitnessRecordIterationBest); i++ {
		log.Printf("Iteration %d: FitnessRecordIterationBest: %f\n", i, geneticAlgorithm.FitnessRecordIterationBest[i])
		log.Printf("Iteration %d: FitnessRecordIterationBestAcceptable: %f\n", i, geneticAlgorithm.FitnessRecordIterationBestAcceptable[i])
	}

	log.Println()
	if len(geneticAlgorithm.FitnessRecordBestUntilNow) != len(geneticAlgorithm.BestUntilNowUpdateIterations) {
		log.Panicf("len(geneticAlgorithm.FitnessRecordBestUntilNow): %d, len(geneticAlgorithm.BestUntilNowUpdateIterations): %d\n", len(geneticAlgorithm.FitnessRecordBestUntilNow), len(geneticAlgorithm.BestUntilNowUpdateIterations))
	}

	for i := 0; i < len(geneticAlgorithm.FitnessRecordBestAcceptableUntilNow); i++ {
		log.Printf("Iteration %d: FitnessRecordBestAcceptableUntilNow: %f\n", int(geneticAlgorithm.BestAcceptableUntilNowUpdateIterations[i]), geneticAlgorithm.FitnessRecordBestAcceptableUntilNow[i])
	}

	if err != nil {
		log.Printf("geneticAlgorithm.Schedule(clouds, apps), error: %s", err.Error())
	}
	log.Println("solution:", solution)

	// draw geneticAlgorithm.FitnessRecordIterationBest and geneticAlgorithm.FitnessRecordBestUntilNow on a line chart
	geneticAlgorithm.DrawChart()

}
