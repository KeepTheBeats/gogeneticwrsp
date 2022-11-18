package algorithms

import (
	"gogeneticwrsp/model"
	"sort"
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

		// Add this app to the record of running apps on this cloud
		thisApp := apps[appIndex]
		deployedClouds[solution.SchedulingResult[appIndex]].RunningApps = append(deployedClouds[solution.SchedulingResult[appIndex]].RunningApps, thisApp)

		// service does not need to subtract resources here, instead, we check it latter

		// task does not need to subtract resources, because:
		// 1. at one time, only one task will be executed on the cloud;
		// 2. we will compare the remaining resources with the tasks with highest requirements for acceptance check
	}
	return deployedClouds
}

// Acceptable check whether a chromosome is acceptable
func Acceptable(clouds []model.Cloud, apps []model.Application, schedulingResult []int) bool {

	var deployedClouds []model.Cloud = SimulateDeploy(clouds, apps, model.Solution{SchedulingResult: schedulingResult})

	// check every cloud
	for cloudIndex := 0; cloudIndex < len(deployedClouds); cloudIndex++ {

		curCPULC := deployedClouds[cloudIndex].Allocatable.CPU.LogicalCores
		curMem := deployedClouds[cloudIndex].Allocatable.Memory
		curStorage := deployedClouds[cloudIndex].Allocatable.Storage

		// if a task together with the services with priorities higher than this task uses up any type of resources, this solution cannot be accpeted
		sort.Sort(model.AppSlice(deployedClouds[cloudIndex].RunningApps))
		for _, deployedApp := range deployedClouds[cloudIndex].RunningApps {
			//fmt.Println(deployedApp)
			//fmt.Println(curCPULC, curMem, curStorage)
			if deployedApp.IsTask { // task releases resources after completion, so all applications with the priorities lower than it can wait for its completion, so it will not block other applications with lower priorities.
				if curMem < deployedApp.TaskReq.Memory || curStorage < deployedApp.TaskReq.Storage {
					return false
				}
			} else { // service should take up the resources, so it will block other applications.
				curCPULC -= deployedApp.SvcReq.CPUClock / deployedClouds[cloudIndex].Allocatable.CPU.BaseClock
				curMem -= deployedApp.SvcReq.Memory
				curStorage -= deployedApp.SvcReq.Storage
				if curCPULC < 0 || curMem < 0 || curStorage < 0 {
					return false
				}
			}

			// network bandwidths and RTT
			for _, dependence := range deployedApp.Depend {
				dependentCloudIdx := schedulingResult[dependence.AppIdx]
				if dependentCloudIdx == len(deployedClouds) {
					return false // the dependent app is rejected
				}

				// cloudIndex: current cloud
				// dependentCloudIdx: dependent cloud

				// check RTT requirements
				if deployedClouds[cloudIndex].Allocatable.NetCondClouds[dependentCloudIdx].RTT > dependence.RTT {
					return false
				}

				// check downstream bandwidth requirements
				if deployedApp.IsTask {
					if deployedClouds[cloudIndex].Allocatable.NetCondClouds[dependentCloudIdx].DownBw < dependence.DownBw {
						return false
					}
				} else {
					deployedClouds[cloudIndex].Allocatable.NetCondClouds[dependentCloudIdx].DownBw -= dependence.DownBw
					if deployedClouds[cloudIndex].Allocatable.NetCondClouds[dependentCloudIdx].DownBw < 0 {
						return false
					}
				}
			}
		}
	}

	return true
}

// CloudMeetApp check whether a cloud can meet an application
func CloudMeetApp(cloud model.Cloud, app model.Application) bool {

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
