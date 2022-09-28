package main

import (
	"log"

	"gogeneticwrsp/algorithms"
	"gogeneticwrsp/experiments"
	"gogeneticwrsp/model"
)

func main() {
	// set the log to show line number and file name
	log.SetFlags(log.LstdFlags | log.Llongfile)

	log.Println("Hello World!")

	var numCloud, numApp int = 4, 10

	// generate clouds and apps, and write to files
	//experiments.GenerateCloudsApps(numCloud, numApp)

	// read clouds and apps from files
	var clouds []model.Cloud
	var apps []model.Application
	clouds, apps = experiments.ReadCloudsApps(numCloud, numApp)

	for i := 0; i < numCloud; i++ {
		log.Println(clouds[i])
	}

	for i := 0; i < numApp; i++ {
		log.Println(apps[i])
	}

	geneticAlgorithm := algorithms.NewGenetic(10, 10, clouds, apps)

	solution, err := geneticAlgorithm.Schedule(clouds, apps)
	if err != nil {
		log.Printf("geneticAlgorithm.Schedule(clouds, apps), error: %s", err.Error())
	}
	log.Println("solution:", solution)

}
