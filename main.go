package main

import (
	"fmt"
	"gogeneticwrsp/algorithms"
	"log"
	"time"

	"gogeneticwrsp/experimenttools"
	"gogeneticwrsp/model"
)

func main() {
	// set the log to show line number and file name
	log.SetFlags(0 | log.Lshortfile)

	//log.Println("Hello World!")

	//var numCloud, numApp int = 10, 140
	var numCloud, numApp int = 5, 40
	var appSuffix string = "0"

	// generate clouds and apps, and write to files
	//experimenttools.GenerateCloudsApps(numCloud, numApp, appSuffix)
	experimenttools.GenerateClouds(numCloud)
	experimenttools.GenerateApps(numApp, appSuffix, 0.5)

	// read clouds and apps from files
	var clouds []model.Cloud
	var apps []model.Application
	//clouds, apps = experimenttools.ReadCloudsApps(numCloud, numApp, appSuffix)
	clouds = experimenttools.ReadClouds(numCloud)
	apps = experimenttools.ReadApps(numApp, appSuffix)
	//for i := 0; i < numCloud; i++ {
	//	log.Println(clouds[i])
	//}
	//
	//for i := 0; i < numApp; i++ {
	//	log.Println(apps[i])
	//}

	var CPUCapa, CPUReq, MemCapa, MemReq, StoCapa, StoReq, BWCapa, BWReq float64
	for i := 0; i < numCloud; i++ {
		CPUCapa += clouds[i].Capacity.CPU.BaseClock * clouds[i].Capacity.CPU.LogicalCores
		MemCapa += clouds[i].Capacity.Memory
		StoCapa += clouds[i].Capacity.Storage
		for j := 0; j < numCloud; j++ {
			if j != i {
				BWCapa += clouds[i].Capacity.NetCondClouds[j].DownBw
			}
		}
	}

	for i := 0; i < numApp; i++ {
		if !apps[i].IsTask {
			CPUReq += apps[i].SvcReq.CPUClock
			MemReq += apps[i].SvcReq.Memory
			StoReq += apps[i].SvcReq.Storage
		} else {
			MemReq += apps[i].TaskReq.Memory
			StoReq += apps[i].TaskReq.Storage
		}
		for j := 0; j < len(apps[i].Depend); j++ {
			BWReq += (apps[i].Depend[j].DownBw + apps[i].Depend[j].UpBw)
		}
	}
	fmt.Println("CPU:", CPUReq, CPUCapa, CPUReq/CPUCapa)
	fmt.Println("Mem:", MemReq, MemCapa, MemReq/MemCapa)
	fmt.Println("Sto:", StoReq, StoCapa, StoReq/StoCapa)
	fmt.Println("BW:", BWReq, BWCapa, BWReq/BWCapa)

	//geneticAlgorithm := algorithms.NewGenetic(200, 5000, 0.7, 0.01, 200, algorithms.InitializeUndeployedChromosome, clouds, apps)
	//geneticAlgorithm := algorithms.NewGenetic(100, 5000, 0.7, 0.007, 200, algorithms.InitializeAcceptableChromosome, clouds, apps)
	geneticAlgorithm := algorithms.NewGenetic(200, 5000, 0.4, 0.003, 250, algorithms.RandomFitSchedule, algorithms.OnePointCrossOver, true, false, clouds, apps)
	haga := algorithms.NewHAGA(10, 0.6, 200, 5000, 0.6, 0.7, 250, clouds, apps)
	nsga := algorithms.NewNSGAII(200, 5000, 1, 0.25, 250, clouds, apps)
	ff := algorithms.NewFirstFit(clouds, apps)
	rf := algorithms.NewRandomFit(clouds, apps)
	geneticBefore := time.Now()
	geneticSolution, err := geneticAlgorithm.Schedule(clouds, apps)
	geneticAfter := time.Now()
	if err != nil {
		//log.Printf("geneticAlgorithm.Schedule(clouds, apps), error: %s", err.Error())
		log.Panicf("geneticAlgorithm.Schedule(clouds, apps), error: %s", err.Error())
	}

	hagaBefore := time.Now()
	hagaSolution, err := haga.Schedule(clouds, apps)
	hagaAfter := time.Now()
	if err != nil {
		//log.Printf("haga.Schedule(clouds, apps), error: %s", err.Error())
		log.Panicf("haga.Schedule(clouds, apps), error: %s", err.Error())
	}

	nsgaBefore := time.Now()
	nsgaSolution, err := nsga.Schedule(clouds, apps)
	nsgaAfter := time.Now()
	if err != nil {
		//log.Printf("nsga.Schedule(clouds, apps), error: %s", err.Error())
		log.Panicf("nsga.Schedule(clouds, apps), error: %s", err.Error())
	}

	ffBefore := time.Now()
	ffSolution, err := ff.Schedule(clouds, apps)
	ffAfter := time.Now()
	if err != nil {
		//log.Printf("ff.Schedule(clouds, apps), error: %s", err.Error())
		log.Panicf("ff.Schedule(clouds, apps), error: %s", err.Error())
	}

	rfBefore := time.Now()
	rfSolution, err := rf.Schedule(clouds, apps)
	rfAfter := time.Now()
	if err != nil {
		//log.Printf("rf.Schedule(clouds, apps), error: %s", err.Error())
		log.Panicf("rf.Schedule(clouds, apps), error: %s", err.Error())
	}

	log.Println("geneticAlgorithm calculation time:", geneticAfter.Sub(geneticBefore).Seconds())
	log.Println("haga calculation time:", hagaAfter.Sub(hagaBefore).Seconds())
	log.Println("nsga calculation time:", nsgaAfter.Sub(nsgaBefore).Seconds())
	log.Println("ff calculation time:", ffAfter.Sub(ffBefore).Seconds())
	log.Println("rf calculation time:", rfAfter.Sub(rfBefore).Seconds())

	tmpClouds := model.CloudsCopy(clouds)
	tmpApps := model.AppsCopy(apps)
	tmpSolution := model.SolutionCopy(geneticSolution)

	tmpClouds = algorithms.SimulateDeploy(tmpClouds, tmpApps, tmpSolution)
	algorithms.CalcStartComplTime(tmpClouds, tmpApps, tmpSolution.SchedulingResult)
	log.Println("final, geneticAlgorithm.RejectExecTime:", geneticAlgorithm.RejectExecTime)
	for j := 0; j < len(tmpClouds); j++ {
		log.Println(tmpClouds[j].TotalTaskComplTime, len(tmpClouds[j].RunningApps))
	}

	//for i := 0; i < len(geneticAlgorithm.FitnessRecordIterationBest); i++ {
	//	log.Printf("Iteration %d: FitnessRecordIterationBest: %f\n", i, geneticAlgorithm.FitnessRecordIterationBest[i])
	//	log.Printf("Iteration %d: FitnessRecordIterationBestAcceptable: %f\n", i, geneticAlgorithm.FitnessRecordIterationBestAcceptable[i])
	//}

	log.Println()
	if len(geneticAlgorithm.FitnessRecordBestUntilNow) != len(geneticAlgorithm.BestUntilNowUpdateIterations) {
		log.Panicf("len(geneticAlgorithm.FitnessRecordBestUntilNow): %d, len(geneticAlgorithm.BestUntilNowUpdateIterations): %d\n", len(geneticAlgorithm.FitnessRecordBestUntilNow), len(geneticAlgorithm.BestUntilNowUpdateIterations))
	}

	for i := 0; i < len(geneticAlgorithm.FitnessRecordBestAcceptableUntilNow); i++ {
		log.Printf("Iteration %d: FitnessRecordBestAcceptableUntilNow: %f\n", int(geneticAlgorithm.BestAcceptableUntilNowUpdateIterations[i]), geneticAlgorithm.FitnessRecordBestAcceptableUntilNow[i])
	}

	log.Println("geneticSolution:", geneticSolution)
	log.Println("hagaSolution:", hagaSolution)
	log.Println("nsgaSolution:", nsgaSolution)
	log.Println("ffSolution:", ffSolution)
	log.Println("rfSolution:", rfSolution)

	// draw geneticAlgorithm.FitnessRecordIterationBest and geneticAlgorithm.FitnessRecordBestUntilNow on a line chart
	geneticAlgorithm.DrawChart()

}
