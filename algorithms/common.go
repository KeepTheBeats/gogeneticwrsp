package algorithms

import (
	"gogeneticwrsp/model"
	"sort"
)

// SchedulingAlgorithm is the interface that all algorithms should implement
type SchedulingAlgorithm interface {
	Schedule(clouds []model.Cloud, apps []model.Application) (model.Solution, error)
}

// SimulateDeploy record the applications in the target cloud's RunningApps
func SimulateDeploy(clouds []model.Cloud, apps []model.Application, solution model.Solution) []model.Cloud {
	var appsCopy []model.Application = model.AppsCopy(apps)
	var deployedClouds []model.Cloud = model.CloudsCopy(clouds)
	for appIndex := 0; appIndex < len(solution.SchedulingResult); appIndex++ {
		// this application is rejected, and will not be deployed
		if solution.SchedulingResult[appIndex] == len(clouds) {
			continue
		}

		// Add this app to the record of running apps on this cloud
		thisApp := appsCopy[appIndex]
		deployedClouds[solution.SchedulingResult[appIndex]].RunningApps = append(deployedClouds[solution.SchedulingResult[appIndex]].RunningApps, thisApp)

		// service does not need to subtract resources here, instead, we check it latter

		// task does not need to subtract resources, because:
		// 1. at one time, only one task will be executed on the cloud;
		// 2. we will compare the remaining resources with the tasks with the highest requirements for acceptance check
	}
	return deployedClouds
}

// TrulyDeploy
// 1. record the applications in the target cloud's RunningApps
// 2. delete the cloud resources needed by applications
func TrulyDeploy(clouds []model.Cloud, apps []model.Application, solution model.Solution) []model.Cloud {
	var appsCopy []model.Application = model.AppsCopy(apps)
	var deployedClouds []model.Cloud = model.CloudsCopy(clouds)
	for appIndex := 0; appIndex < len(solution.SchedulingResult); appIndex++ {
		// this application is rejected, and will not be deployed
		if solution.SchedulingResult[appIndex] == len(clouds) {
			continue
		}

		thisApp := appsCopy[appIndex]
		cloudIndex := solution.SchedulingResult[appIndex]

		// Add this app to the record of running apps on this cloud
		deployedClouds[cloudIndex].RunningApps = append(deployedClouds[cloudIndex].RunningApps, thisApp)

		// subtract cloud allocatable resources
		if !thisApp.IsTask {
			deployedClouds[cloudIndex].Allocatable.CPU.LogicalCores -= thisApp.SvcReq.CPUClock / deployedClouds[cloudIndex].Allocatable.CPU.BaseClock
			deployedClouds[cloudIndex].Allocatable.Memory -= thisApp.SvcReq.Memory
			deployedClouds[cloudIndex].Allocatable.Storage -= thisApp.SvcReq.Storage

			for _, dependence := range thisApp.Depend {
				dependCloudIdx := solution.SchedulingResult[dependence.AppIdx]
				deployedClouds[cloudIndex].Allocatable.NetCondClouds[dependCloudIdx].DownBw -= dependence.DownBw
				deployedClouds[dependCloudIdx].Allocatable.NetCondClouds[cloudIndex].DownBw -= dependence.UpBw
			}
		}

		// task does not need to subtract resources, because:
		// 1. at one time, only one task will be executed on the cloud;
		// 2. we will compare the remaining resources with the tasks with the highest requirements for acceptance check
	}
	return deployedClouds
}

// Acceptable check whether a chromosome is acceptable
func Acceptable(clouds []model.Cloud, apps []model.Application, schedulingResult []int) bool {

	var deployedClouds []model.Cloud = SimulateDeploy(clouds, apps, model.Solution{SchedulingResult: schedulingResult})

	// check every cloud
	for cloudIndex := 0; cloudIndex < len(deployedClouds); cloudIndex++ {

		deployedClouds[cloudIndex].TmpAlloc = model.ResCopy(deployedClouds[cloudIndex].Allocatable)

		curCPULC := deployedClouds[cloudIndex].TmpAlloc.CPU.LogicalCores
		curMem := deployedClouds[cloudIndex].TmpAlloc.Memory
		curStorage := deployedClouds[cloudIndex].TmpAlloc.Storage

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
				curCPULC -= deployedApp.SvcReq.CPUClock / deployedClouds[cloudIndex].TmpAlloc.CPU.BaseClock
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
				if deployedClouds[cloudIndex].TmpAlloc.NetCondClouds[dependentCloudIdx].RTT > dependence.RTT {
					return false
				}

				// check downstream and upstream bandwidth requirements
				if deployedApp.IsTask {
					if deployedClouds[cloudIndex].TmpAlloc.NetCondClouds[dependentCloudIdx].DownBw < dependence.DownBw ||
						deployedClouds[dependentCloudIdx].TmpAlloc.NetCondClouds[cloudIndex].DownBw < dependence.UpBw {
						return false
					}
				} else {
					deployedClouds[cloudIndex].TmpAlloc.NetCondClouds[dependentCloudIdx].DownBw -= dependence.DownBw
					deployedClouds[dependentCloudIdx].TmpAlloc.NetCondClouds[cloudIndex].DownBw -= dependence.UpBw
					if deployedClouds[cloudIndex].TmpAlloc.NetCondClouds[dependentCloudIdx].DownBw < 0 || deployedClouds[dependentCloudIdx].TmpAlloc.NetCondClouds[cloudIndex].DownBw < 0 {
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
