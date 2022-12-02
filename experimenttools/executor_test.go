package experimenttools

import (
	"fmt"
	"github.com/KeepTheBeats/routing-algorithms/random"
	"gogeneticwrsp/algorithms"
	"gogeneticwrsp/model"
	"log"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCloudsApps(t *testing.T) {
	var n int = 5
	for i := 0; i < n; i++ {
		numApps := random.RandomInt(3, 6)
		GenerateApps(numApps, fmt.Sprintf("%d", i), 0.5)
	}
}

func TestNoChangeMCASGA(t *testing.T) {
	var numCloud, numApp int = 5, 15
	var appSuffix string = "0"

	// generate clouds and apps, and write to files
	GenerateClouds(numCloud)
	GenerateApps(numApp, appSuffix, 0.5)

	// read clouds and apps from files
	var clouds []model.Cloud
	var apps []model.Application

	clouds = ReadClouds(numCloud)
	apps = ReadApps(numApp, appSuffix)
	clouds = model.CloudsCopy(clouds)
	apps = model.AppsCopy(apps)

	oriClouds := model.CloudsCopy(clouds)
	oriApps := model.AppsCopy(apps)

	geneticAlgorithm := algorithms.NewGenetic(200, 5000, 0.7, 0.007, 200, algorithms.RandomFitSchedule, algorithms.OnePointCrossOver, false, false, clouds, apps)

	solution, err := geneticAlgorithm.Schedule(clouds, apps)
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
	log.Println("solution:", solution)
	log.Println(geneticAlgorithm.FitnessRecordBestAcceptableUntilNow[i])

	assert.Equal(t, oriClouds, clouds, fmt.Sprintf("oriClouds and clouds are not equal"))
	assert.Equal(t, oriApps, apps, fmt.Sprintf("oriApps and apps are not equal"))

	fmt.Println(reflect.DeepEqual(oriClouds, clouds))
	fmt.Println(reflect.DeepEqual(oriApps, apps))
}

func TestNoChangeFirstFit(t *testing.T) {
	var numCloud, numApp int = 5, 15
	var appSuffix string = "0"

	// generate clouds and apps, and write to files
	GenerateClouds(numCloud)
	GenerateApps(numApp, appSuffix, 0.5)

	// read clouds and apps from files
	var clouds []model.Cloud
	var apps []model.Application

	clouds = ReadClouds(numCloud)
	apps = ReadApps(numApp, appSuffix)
	clouds = model.CloudsCopy(clouds)
	apps = model.AppsCopy(apps)

	oriClouds := model.CloudsCopy(clouds)
	oriApps := model.AppsCopy(apps)

	ff := algorithms.NewFirstFit(clouds, apps)
	solution, err := ff.Schedule(clouds, apps)
	if err != nil {
		log.Panicf("solution, err := ff.Schedule(oriClouds, oriApps), error: %s", err.Error())
	}

	log.Println("solution:", solution)

	assert.Equal(t, oriClouds, clouds, fmt.Sprintf("oriClouds and clouds are not equal"))
	assert.Equal(t, oriApps, apps, fmt.Sprintf("oriApps and apps are not equal"))

	fmt.Println(reflect.DeepEqual(oriClouds, clouds))
	fmt.Println(reflect.DeepEqual(oriApps, apps))
}

func TestNoChangeRandomFit(t *testing.T) {
	var numCloud, numApp int = 5, 15
	var appSuffix string = "0"

	// generate clouds and apps, and write to files
	GenerateClouds(numCloud)
	GenerateApps(numApp, appSuffix, 0.5)

	// read clouds and apps from files
	var clouds []model.Cloud
	var apps []model.Application

	clouds = ReadClouds(numCloud)
	apps = ReadApps(numApp, appSuffix)
	clouds = model.CloudsCopy(clouds)
	apps = model.AppsCopy(apps)

	oriClouds := model.CloudsCopy(clouds)
	oriApps := model.AppsCopy(apps)

	rf := algorithms.NewRandomFit(clouds, apps)
	solution, err := rf.Schedule(clouds, apps)
	if err != nil {
		log.Panicf("solution, err := rf.Schedule(clouds, apps), error: %s", err.Error())
	}

	log.Println("solution:", solution)

	assert.Equal(t, oriClouds, clouds, fmt.Sprintf("oriClouds and clouds are not equal"))
	assert.Equal(t, oriApps, apps, fmt.Sprintf("oriApps and apps are not equal"))

	fmt.Println(reflect.DeepEqual(oriClouds, clouds))
	fmt.Println(reflect.DeepEqual(oriApps, apps))
}
