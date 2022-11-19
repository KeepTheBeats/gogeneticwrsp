package main

import (
	"fmt"
	"log"

	"gogeneticwrsp/algorithms"
	"gogeneticwrsp/experimenttools"
	"gogeneticwrsp/model"
)

type Experimenter struct {
	Name                string
	ExperimentAlgorithm algorithms.SchedulingAlgorithm
	ExperimentSolution  model.Solution
}

func main() {
	// set the log to show line number and file name
	log.SetFlags(0 | log.Lshortfile)

	var numCloud int = 10
	var numInGroup []int = []int{3, 4, 5, 6, 3, 5, 4} // total 30

	// generate clouds and apps, and write to files
	experimenttools.GenerateClouds(numCloud)
	for i := 0; i < len(numInGroup); i++ {
		experimenttools.GenerateApps(numInGroup[i], fmt.Sprintf("%d", i))
	}

	// read clouds and apps from files
	var clouds []model.Cloud
	var appGroups [][]model.Application
	clouds = experimenttools.ReadClouds(numCloud)
	for i := 0; i < len(numInGroup); i++ {
		appGroups = append(appGroups, experimenttools.ReadApps(numInGroup[i], fmt.Sprintf("%d", i)))
	}

	experimenttools.ContinuousExperiment(clouds, appGroups)

}
