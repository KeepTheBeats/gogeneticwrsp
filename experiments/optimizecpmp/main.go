package main

import (
	"fmt"
	"gogeneticwrsp/algorithms"
	"gogeneticwrsp/experimenttools"
	"gogeneticwrsp/model"
	"log"
	"sort"
)

func main() {
	// set the log to show line number and file name
	log.SetFlags(0 | log.Lshortfile)

	//log.Println("Hello World!")

	var testRecords []cpmpRecord

	for cp := 0.3; cp <= 1.0; cp += 0.1 {
		for mp := 0.003; mp <= 0.02; mp += 0.001 {
			var fitnessSum float64 = 0
			for i := 0; i < 10; i++ {
				fmt.Printf("testing cp: %g, mp: %g, round %d\n", cp, mp, i)
				fitnessSum += testParameters(cp, mp)
			}
			testRecords = append(testRecords, cpmpRecord{
				cp:      cp,
				mp:      mp,
				fitness: fitnessSum / 10,
			})
		}
	}

	sort.Sort(cpmpRecordSloce(testRecords))

	for i := 0; i < len(testRecords); i++ {
		fmt.Printf("cp: %g, mp: %g, fitness: %g\n", testRecords[i].cp, testRecords[i].mp, testRecords[i].fitness)
	}

}

type cpmpRecord struct {
	cp      float64
	mp      float64
	fitness float64
}

type cpmpRecordSloce []cpmpRecord

func (cmrs cpmpRecordSloce) Len() int {
	return len(cmrs)
}

func (cmrs cpmpRecordSloce) Swap(i, j int) {
	cmrs[i], cmrs[j] = cmrs[j], cmrs[i]
}

func (cmrs cpmpRecordSloce) Less(i, j int) bool {
	return cmrs[i].fitness < cmrs[j].fitness
}

func testParameters(crossoverProbability float64, mutationProbability float64) float64 {
	var numCloud, numApp int = 10, 30
	var appSuffix string = "0"

	// generate clouds and apps, and write to files
	//experimenttools.GenerateClouds(numCloud)
	//experimenttools.GenerateApps(numApp, appSuffix)

	// read clouds and apps from files
	var clouds []model.Cloud
	var apps []model.Application

	clouds = experimenttools.ReadClouds(numCloud)
	apps = experimenttools.ReadApps(numApp, appSuffix)

	geneticAlgorithm := algorithms.NewGenetic(200, 5000, crossoverProbability, mutationProbability, 200, algorithms.RandomFitSchedule, clouds, apps)

	_, err := geneticAlgorithm.Schedule(clouds, apps)
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
	//log.Println("solution:", solution)
	return geneticAlgorithm.FitnessRecordBestAcceptableUntilNow[i]

}
