package main

import (
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

	var numCloud, numApp int = 10, 70

	// generate clouds and apps, and write to files
	//experimenttools.GenerateCloudsApps(numCloud, numApp)

	// read clouds and apps from files
	var clouds []model.Cloud
	var apps []model.Application
	clouds, apps = experimenttools.ReadCloudsApps(numCloud, numApp)

	experimenttools.ContinuousExperiment(clouds, apps)

}
