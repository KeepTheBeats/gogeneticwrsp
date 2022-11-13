package algorithms

import (
	"gogeneticwrsp/model"
)

// SchedulingAlgorithm is the interface that all algorithms should implement
type SchedulingAlgorithm interface {
	Schedule(clouds []model.Cloud, apps []model.Application) (model.Solution, error)
}

// SimulateDeploy delete the cloud resources needed by applications
func SimulateDeploy(clouds []model.Cloud, apps []model.Application, solution model.Solution) []model.Cloud {
	var deployedClouds []model.Cloud = make([]model.Cloud, len(clouds))
	copy(deployedClouds, clouds)
	for appIndex := 0; appIndex < len(solution.SchedulingResult); appIndex++ {
		// this application is rejected, and will not be deployed
		if solution.SchedulingResult[appIndex] == len(clouds) {
			continue
		}

		if !apps[appIndex].IsTask { // service
			deployedClouds[solution.SchedulingResult[appIndex]].Allocatable.CPU.LogicalCores -= apps[appIndex].SvcReq.CPUClock / deployedClouds[solution.SchedulingResult[appIndex]].Allocatable.CPU.BaseClock
			deployedClouds[solution.SchedulingResult[appIndex]].Allocatable.Memory -= apps[appIndex].SvcReq.Memory
			deployedClouds[solution.SchedulingResult[appIndex]].Allocatable.Storage -= apps[appIndex].SvcReq.Storage
			// NetLatency does not need to be subtracted, but if we use NetBandwidth, we need to subtract it.
		} else { // task
			deployedClouds[solution.SchedulingResult[appIndex]].Allocatable.CPU.LogicalCores -= (apps[appIndex].TaskReq.CPUCycle / 1024 / 1024 / 1024) / deployedClouds[solution.SchedulingResult[appIndex]].Allocatable.CPU.BaseClock
			deployedClouds[solution.SchedulingResult[appIndex]].Allocatable.Memory -= apps[appIndex].TaskReq.Memory
			deployedClouds[solution.SchedulingResult[appIndex]].Allocatable.Storage -= apps[appIndex].TaskReq.Storage
			// NetLatency does not need to be subtracted, but if we use NetBandwidth, we need to subtract it.
		}
	}
	return deployedClouds
}

// Acceptable check whether a chromosome is acceptable
func Acceptable(clouds []model.Cloud, apps []model.Application, schedulingResult []int) bool {
	//log.Println(clouds)
	//log.Println(apps)
	//log.Println(schedulingResult)
	var deployedClouds []model.Cloud = SimulateDeploy(clouds, apps, model.Solution{SchedulingResult: schedulingResult})
	//log.Println(deployedClouds)

	// check every cloud
	for cloudIndex := 0; cloudIndex < len(deployedClouds); cloudIndex++ {
		if deployedClouds[cloudIndex].Allocatable.CPU.LogicalCores < 0 || deployedClouds[cloudIndex].Allocatable.Memory < 0 || deployedClouds[cloudIndex].Allocatable.Storage < 0 {
			//log.Println("deployedClouds[cloudIndex].Allocatable.CPUClock", deployedClouds[cloudIndex].Allocatable.CPUClock)
			//log.Println("deployedClouds[cloudIndex].Allocatable.Memory", deployedClouds[cloudIndex].Allocatable.Memory)
			//log.Println("deployedClouds[cloudIndex].Allocatable.Storage", deployedClouds[cloudIndex].Allocatable.Storage)
			return false
		}
	}

	//the check network latency for every application
	//for appIndex := 0; appIndex < len(schedulingResult); appIndex++ {
	//	if schedulingResult[appIndex] != len(clouds) {
	//		if deployedClouds[schedulingResult[appIndex]].Allocatable.NetLatency > apps[appIndex].SvcReq.NetLatency {
	//			fmt.Println("this reason")
	//			return false
	//		}
	//	}
	//}

	return true
}

// MeetNetLatency check whether a cloud can meet the network latency requirement of an application
func MeetNetLatency(cloud model.Cloud, app model.Application) bool {
	if app.IsTask { // task
		if cloud.Allocatable.NetLatency > app.TaskReq.NetLatency {
			return false
		}
	} else { // service
		if cloud.Allocatable.NetLatency > app.SvcReq.NetLatency {
			return false
		}
	}
	return true
}

// CPUIdleRate calculates the CPUClock idle rate according to given clouds, apps, schedulingResult
func CPUIdleRate(clouds []model.Cloud, apps []model.Application, schedulingResult []int) float64 {
	var deployedClouds []model.Cloud = SimulateDeploy(clouds, apps, model.Solution{SchedulingResult: schedulingResult})
	var idleCPU, totalCPU float64
	for i := 0; i < len(clouds); i++ {
		idleCPU += deployedClouds[i].Allocatable.CPU.LogicalCores
		totalCPU += deployedClouds[i].Capacity.CPU.LogicalCores
	}
	return idleCPU / totalCPU
}

// MemoryIdleRate calculates the Memory idle rate according to given clouds, apps, schedulingResult
func MemoryIdleRate(clouds []model.Cloud, apps []model.Application, schedulingResult []int) float64 {
	var deployedClouds []model.Cloud = SimulateDeploy(clouds, apps, model.Solution{SchedulingResult: schedulingResult})
	var idleMemory, totalMemory float64
	for i := 0; i < len(clouds); i++ {
		idleMemory += deployedClouds[i].Allocatable.Memory
		totalMemory += deployedClouds[i].Capacity.Memory
	}
	return idleMemory / totalMemory
}

// StorageIdleRate calculates the Memory idle rate according to given clouds, apps, schedulingResult
func StorageIdleRate(clouds []model.Cloud, apps []model.Application, schedulingResult []int) float64 {
	var deployedClouds []model.Cloud = SimulateDeploy(clouds, apps, model.Solution{SchedulingResult: schedulingResult})
	var idleStorage, totalStorage float64
	for i := 0; i < len(clouds); i++ {
		idleStorage += deployedClouds[i].Allocatable.Storage
		totalStorage += deployedClouds[i].Capacity.Storage
	}
	return idleStorage / totalStorage
}

// AcceptedPriority calculates the total priority of accept applications according to given clouds, apps, schedulingResult
func AcceptedPriority(clouds []model.Cloud, apps []model.Application, schedulingResult []int) uint64 {
	var totalPriority uint64
	for i := 0; i < len(apps); i++ {
		if schedulingResult[i] != len(clouds) {
			totalPriority += uint64(apps[i].Priority)
		}
	}
	return totalPriority
}

// TotalPriority calculates the total priority of all applications according to given clouds, apps, schedulingResult
func TotalPriority(clouds []model.Cloud, apps []model.Application, schedulingResult []int) uint64 {
	var totalPriority uint64
	for i := 0; i < len(apps); i++ {
		totalPriority += uint64(apps[i].Priority)
	}
	return totalPriority
}
