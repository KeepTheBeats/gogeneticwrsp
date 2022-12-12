package testalgorithms

import (
	"gogeneticwrsp/algorithms"
	"gogeneticwrsp/experimenttools"
	"gogeneticwrsp/model"
	"log"
	"testing"
)

func TestHagaSchedule(t *testing.T) {
	log.SetFlags(0 | log.Lshortfile)

	var numCloud, numApp int = 10, 140
	var appSuffix string = "0"

	// generate clouds and apps, and write to files
	experimenttools.GenerateClouds(numCloud)
	experimenttools.GenerateApps(numApp, appSuffix, 0.5)

	// read clouds and apps from files
	var clouds []model.Cloud
	var apps []model.Application

	clouds = experimenttools.ReadClouds(numCloud)
	apps = experimenttools.ReadApps(numApp, appSuffix)

	hagaAlgorithm := algorithms.NewHAGA(10, 0.6, 200, 5000, 0.6, 0.7, 250, clouds, apps)
	solution, err := hagaAlgorithm.Schedule(clouds, apps)
	if err != nil {
		log.Panicf("geneticAlgorithm.Schedule(clouds, apps), error: %s", err.Error())
	}

	log.Println()
	log.Println("solution:", solution)
}
