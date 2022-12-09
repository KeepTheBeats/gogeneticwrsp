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
	var twoPointCrossover, btSelection, cbMutation bool
	var boolValues []bool = []bool{false, true}
	var mpMinCb, mpMaxCb, mpStepCb float64 = 0.0, 1.0, 0.05
	var mpMinGb, mpMaxGb, mpStepGb float64 = 0.001, 0.02, 0.001
	var mpMinThis, mpMaxThis, mpStepThis float64
	for _, v1 := range boolValues {
		twoPointCrossover = v1
		for _, v2 := range boolValues {
			btSelection = v2
			for _, v3 := range boolValues {
				cbMutation = v3
				if cbMutation {
					mpMinThis, mpMaxThis, mpStepThis = mpMinCb, mpMaxCb, mpStepCb
				} else {
					mpMinThis, mpMaxThis, mpStepThis = mpMinGb, mpMaxGb, mpStepGb
				}
				for cp := 0.0; cp <= 1.0; cp += 0.1 {
					for mp := mpMinThis; mp <= mpMaxThis; mp += mpStepThis {
						var fitnessSum float64 = 0
						for i := 0; i < 10; i++ {
							fmt.Printf("testing cp: %g, mp: %g, twoPointCrossover: %t, btSelection: %t, cbMutation: %t, round %d\n", cp, mp, twoPointCrossover, btSelection, cbMutation, i)
							fitnessSum += testParameters(cp, mp, twoPointCrossover, btSelection, cbMutation)
						}
						testRecords = append(testRecords, cpmpRecord{
							cp:                cp,
							mp:                mp,
							twoPointCrossover: twoPointCrossover,
							btSelection:       btSelection,
							cbMutation:        cbMutation,
							fitness:           fitnessSum / 10,
						})
					}
				}
			}
		}
	}

	sort.Sort(cpmpRecordSloce(testRecords))

	for i := 0; i < len(testRecords); i++ {
		fmt.Printf("cp: %g, mp: %g, twoPointCrossover: %t, btSelection: %t, cbMutation: %t, fitness: %g\n", testRecords[i].cp, testRecords[i].mp, testRecords[i].twoPointCrossover, testRecords[i].btSelection, testRecords[i].cbMutation, testRecords[i].fitness)
	}

}

type cpmpRecord struct {
	cp                float64
	mp                float64
	twoPointCrossover bool
	btSelection       bool
	cbMutation        bool
	fitness           float64
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

func testParameters(crossoverProbability float64, mutationProbability float64, twoPointCrossover, btSelection, cbMutation bool) float64 {
	var numCloud, numApp int = 5, 40
	var appSuffix string = "0"

	// generate clouds and apps, and write to files
	//experimenttools.GenerateClouds(numCloud)
	//experimenttools.GenerateApps(numApp, appSuffix, 0.5)

	// read clouds and apps from files
	var clouds []model.Cloud
	var apps []model.Application

	clouds = experimenttools.ReadClouds(numCloud)
	apps = experimenttools.ReadApps(numApp, appSuffix)

	var crossoverOperator func(algorithms.Chromosome, algorithms.Chromosome) (algorithms.Chromosome, algorithms.Chromosome)
	if twoPointCrossover {
		crossoverOperator = algorithms.TwoPointCrossOver
	} else {
		crossoverOperator = algorithms.OnePointCrossOver
	}

	geneticAlgorithm := algorithms.NewGenetic(100, 5000, crossoverProbability, mutationProbability, 100, algorithms.RandomFitSchedule, crossoverOperator, btSelection, cbMutation, clouds, apps)

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
