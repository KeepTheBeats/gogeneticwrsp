package main

import (
	"gogeneticwrsp/algorithms"
	"gogeneticwrsp/experimenttools"
	"gogeneticwrsp/model"
	"log"
	"os"
	"strconv"
)

type Experimenter struct {
	Name                string
	ExperimentAlgorithm algorithms.SchedulingAlgorithm
	ExperimentSolution  model.Solution
}

func main() {
	// set the log to show line number and file name
	log.SetFlags(0 | log.Lshortfile)
	var taskProportion float64
	// read task proportion form the input parameter
	if len(os.Args) > 1 {
		var errParse error
		taskProportion, errParse = strconv.ParseFloat(os.Args[1], 64)
		if errParse != nil {
			log.Println("errParse,", errParse)
			taskProportion = 0.5
		}
	} else {
		taskProportion = 0.5
	}
	log.Println("taskProportion", taskProportion)

	var numCloud, numApp int = 7, 100
	var appSuffix string = "0"

	// generate clouds and apps, and write to files
	//experimenttools.GenerateCloudsApps(numCloud, numApp, appSuffix)
	experimenttools.GenerateClouds(numCloud)
	experimenttools.GenerateApps(numApp, appSuffix, taskProportion)

	// read clouds and apps from files
	var clouds []model.Cloud
	var apps []model.Application
	//clouds, apps = experimenttools.ReadCloudsApps(numCloud, numApp, appSuffix)
	clouds = experimenttools.ReadClouds(numCloud)
	apps = experimenttools.ReadApps(numApp, appSuffix)

	experimenttools.OneTimeExperiment(clouds, apps)

}
