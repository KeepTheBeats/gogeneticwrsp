package main

import (
	"fmt"
	"log"
	"time"

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
	var numInGroup []int = []int{6, 9, 19, 13, 7, 10, 6} // total 70
	var appArrivalTimeIntervals []time.Duration = []time.Duration{0 * time.Second, 20 * time.Second, 30 * time.Second, 30 * time.Second, 15 * time.Second, 15 * time.Second, 15 * time.Second}
	//var numCloud int = 10
	//var numInGroup []int = []int{3, 4, 5, 6, 3, 5, 4} // total 30

	// generate clouds and apps, and write to files
	experimenttools.GenerateClouds(numCloud)
	for i := 0; i < len(numInGroup); i++ {
		experimenttools.GenerateApps(numInGroup[i], fmt.Sprintf("%d", i), 0.5)
	}

	// read clouds and apps from files
	var clouds []model.Cloud
	var appGroups [][]model.Application
	clouds = experimenttools.ReadClouds(numCloud)
	for i := 0; i < len(numInGroup); i++ {
		appGroups = append(appGroups, experimenttools.ReadApps(numInGroup[i], fmt.Sprintf("%d", i)))
	}

	experimenttools.ContinuousExperiment(clouds, appGroups, appArrivalTimeIntervals)

}
