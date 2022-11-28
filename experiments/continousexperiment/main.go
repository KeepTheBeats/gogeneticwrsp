package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
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

	var numCloud int = 9
	var groupNum int = 9
	experimenttools.GenerateNumTimeGroup(groupNum)

	var numTime experimenttools.NumTimeGroup = experimenttools.ReadNumTimeGroup(groupNum)
	var numInGroup []int = numTime.NumInGroup
	var appArrivalTimeIntervals []time.Duration = numTime.TimeIntervals

	//var numInGroup []int = []int{6, 9, 19, 13, 7, 10, 6} // total 70
	//var appArrivalTimeIntervals []time.Duration = []time.Duration{0 * time.Second, 20 * time.Second, 30 * time.Second, 30 * time.Second, 15 * time.Second, 15 * time.Second, 15 * time.Second}

	//// generate clouds and apps, and write to files
	experimenttools.GenerateClouds(numCloud)
	for i := 0; i < len(numInGroup); i++ {
		experimenttools.GenerateApps(numInGroup[i], fmt.Sprintf("%d", i), taskProportion)
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
