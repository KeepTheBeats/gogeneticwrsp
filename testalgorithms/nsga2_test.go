package testalgorithms

import (
	"gogeneticwrsp/algorithms"
	"gogeneticwrsp/experimenttools"
	"gogeneticwrsp/model"
	"log"
	"testing"
)

func TestSchedule(t *testing.T) {
	log.SetFlags(0 | log.Lshortfile)

	var numCloud, numApp int = 10, 50
	var appSuffix string = "0"

	// generate clouds and apps, and write to files
	experimenttools.GenerateClouds(numCloud)
	experimenttools.GenerateApps(numApp, appSuffix, 0.5)

	// read clouds and apps from files
	var clouds []model.Cloud
	var apps []model.Application

	clouds = experimenttools.ReadClouds(numCloud)
	apps = experimenttools.ReadApps(numApp, appSuffix)

	nsgaiiAlgorithm := algorithms.NewNSGAII(200, 5000, 1, 0.25, 250, clouds, apps)
	solution, err := nsgaiiAlgorithm.Schedule(clouds, apps)
	if err != nil {
		log.Panicf("geneticAlgorithm.Schedule(clouds, apps), error: %s", err.Error())
	}

	//for i := 0; i < len(nsgaiiAlgorithm.FitnessRecordIterationBest); i++ {
	//	log.Printf("Iteration %d: FitnessRecordIterationBest: %f\n", i, nsgaiiAlgorithm.FitnessRecordIterationBest[i])
	//	log.Printf("Iteration %d: FitnessRecordIterationBestAcceptable: %f\n", i, nsgaiiAlgorithm.FitnessRecordIterationBestAcceptable[i])
	//}

	for i := 0; i < len(nsgaiiAlgorithm.FitnessRecordBestAcceptableUntilNow); i++ {
		log.Printf("Iteration %d: FitnessRecordBestAcceptableUntilNow: %f\n", int(nsgaiiAlgorithm.BestAcceptableUntilNowUpdateIterations[i]), nsgaiiAlgorithm.FitnessRecordBestAcceptableUntilNow[i])
	}

	log.Println()

	log.Println("solution:", solution)
	nsgaiiAlgorithm.DrawChart()
}
