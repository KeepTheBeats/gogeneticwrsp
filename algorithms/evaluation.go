package algorithms

import "gogeneticwrsp/model"

// CPUIdleRate calculates the CPUClock idle rate according to given clouds, apps, schedulingResult
func CPUIdleRate(clouds []model.Cloud, apps []model.Application, schedulingResult []int) float64 {
	var deployedClouds []model.Cloud = TrulyDeploy(clouds, apps, model.Solution{SchedulingResult: schedulingResult})
	var idleCPU, totalCPU float64
	for i := 0; i < len(clouds); i++ {
		idleCPU += deployedClouds[i].Allocatable.CPU.LogicalCores
		totalCPU += deployedClouds[i].Capacity.CPU.LogicalCores
	}
	return idleCPU / totalCPU
}

// MemoryIdleRate calculates the Memory idle rate according to given clouds, apps, schedulingResult
func MemoryIdleRate(clouds []model.Cloud, apps []model.Application, schedulingResult []int) float64 {
	var deployedClouds []model.Cloud = TrulyDeploy(clouds, apps, model.Solution{SchedulingResult: schedulingResult})
	var idleMemory, totalMemory float64
	for i := 0; i < len(clouds); i++ {
		idleMemory += deployedClouds[i].Allocatable.Memory
		totalMemory += deployedClouds[i].Capacity.Memory
	}
	return idleMemory / totalMemory
}

// StorageIdleRate calculates the Memory idle rate according to given clouds, apps, schedulingResult
func StorageIdleRate(clouds []model.Cloud, apps []model.Application, schedulingResult []int) float64 {
	var deployedClouds []model.Cloud = TrulyDeploy(clouds, apps, model.Solution{SchedulingResult: schedulingResult})
	var idleStorage, totalStorage float64
	for i := 0; i < len(clouds); i++ {
		idleStorage += deployedClouds[i].Allocatable.Storage
		totalStorage += deployedClouds[i].Capacity.Storage
	}
	return idleStorage / totalStorage
}

// BwIdleRate calculates the Bandwidth idle rate according to given clouds, apps, schedulingResult
func BwIdleRate(clouds []model.Cloud, apps []model.Application, schedulingResult []int) float64 {
	var deployedClouds []model.Cloud = TrulyDeploy(clouds, apps, model.Solution{SchedulingResult: schedulingResult})
	var idleBw, totalBw float64
	for i := 0; i < len(deployedClouds); i++ {
		for j := 0; j < len(deployedClouds); j++ {
			//if i != j && deployedClouds[i].Capacity.NetCondClouds[j].DownBw != deployedClouds[i].Allocatable.NetCondClouds[j].DownBw {
			if i != j {
				totalBw += deployedClouds[i].Capacity.NetCondClouds[j].DownBw
				idleBw += deployedClouds[i].Allocatable.NetCondClouds[j].DownBw
			}
		}
	}
	//if totalBw == 0 {
	//	idleBw = 1
	//	totalBw = 1
	//}
	return idleBw / totalBw
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

// AcceptedSvcPriRate calculates the priority acceptance rate of services according to given clouds, apps, schedulingResult
func AcceptedSvcPriRate(clouds []model.Cloud, apps []model.Application, schedulingResult []int) float64 {
	var acceptedPriority, totalPriority uint64
	for i := 0; i < len(apps); i++ {
		if !apps[i].IsTask {
			if schedulingResult[i] != len(clouds) {
				acceptedPriority += uint64(apps[i].Priority)
			}
			totalPriority += uint64(apps[i].Priority)
		}
	}
	return float64(acceptedPriority) / float64(totalPriority)
}

// AcceptedTaskPriRate calculates the priority acceptance rate of tasks according to given clouds, apps, schedulingResult
func AcceptedTaskPriRate(clouds []model.Cloud, apps []model.Application, schedulingResult []int) float64 {
	var acceptedPriority, totalPriority uint64
	for i := 0; i < len(apps); i++ {
		if apps[i].IsTask {
			if schedulingResult[i] != len(clouds) {
				acceptedPriority += uint64(apps[i].Priority)
			}
			totalPriority += uint64(apps[i].Priority)
		}
	}
	return float64(acceptedPriority) / float64(totalPriority)
}
