package algorithms

import (
	"gogeneticwrsp/model"
	"log"
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
	//clouds1 := model.CloudsCopy(clouds)
	//apps1 := model.AppsCopy(apps)
	//solution1 := model.SolutionCopy(solution)
	//log.Println(Acceptable(clouds1, apps1, solution1.SchedulingResult))
	//if !Acceptable(clouds1, apps1, solution1.SchedulingResult) {
	//	log.Panicln()
	//}

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
				//log.Println(cloudIndex, dependCloudIdx)
				dependentApp := appsCopy[dependence.AppIdx]

				// apps do not need to communicate with dependent tasks, so in this condition, the dependence does not need network resources
				if dependentApp.IsTask {
					continue
				}

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
				dependentApp := apps[dependence.AppIdx]
				if dependentCloudIdx == len(deployedClouds) {
					return false // the dependent app is rejected
				}

				// cloudIndex: current cloud
				// dependentCloudIdx: dependent cloud

				// apps do not need to communicate with dependent tasks, so in this condition, the dependence does not need network resources
				if dependentApp.IsTask {
					continue
				}

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

// CalcRemainingApps
// resClouds have the information: 1. resource usage; 2. update time of clouds in last round 3. deployed apps
// timeClouds have the information: 1. execution time; 2. deployed apps
// I need to use them to calculate: at this time what applications are still running on each cloud, and how many cycles of them still need to be executed
// timeSinceLastDeploy unit is second
func CalcRemainingApps(resClouds, timeClouds []model.Cloud, timeSinceLastDeploy float64) []model.Application {
	if len(timeClouds) != len(resClouds) {
		log.Panicf("len(timeClouds): %d, len(resClouds): %d\n", len(timeClouds), len(resClouds))
	}

	timeCloudsCopy := model.CloudsCopy(timeClouds)
	//resCloudsCopy := model.CloudsCopy(resClouds)

	var remainingApps []model.Application
	for j := 0; j < len(timeCloudsCopy); j++ {
		// initialize TmpAlloc.CPU.LogicalCores
		timeCloudsCopy[j].TmpAlloc.CPU.LogicalCores = timeCloudsCopy[j].Allocatable.CPU.LogicalCores
		sort.Sort(model.AppSlice(timeCloudsCopy[j].RunningApps))
		for k := 0; k < len(timeCloudsCopy[j].RunningApps); k++ {
			thisApp := timeCloudsCopy[j].RunningApps[k]
			if thisApp.IsTask { // only retain tasks not done
				if timeSinceLastDeploy < thisApp.TaskCompletionTime { // tasks not done
					remainingApp := model.AppCopy(thisApp)
					remainingApp.StartTime = 0
					remainingApp.ImagePullDoneTime = 0
					remainingApp.DataInputDoneTime = 0
					remainingApp.StableTime = 0
					remainingApp.TaskCompletionTime = 0
					remainingApp.IsNew = false
					remainingApp.CloudRemainingOn = j
					remainingApp.ImagePullDone = false
					remainingApp.AlreadyStable = false
					remainingApp.CanMigrate = true

					if timeSinceLastDeploy > thisApp.ImagePullDoneTime { // after image pulling
						remainingApp.ImagePullDone = true
					}

					// tasks being executed
					if timeSinceLastDeploy > thisApp.StableTime {
						// calculate how many CPU cycles are not done
						executedTime := timeSinceLastDeploy - thisApp.StableTime
						executedCycles := executedTime * (timeCloudsCopy[j].TmpAlloc.CPU.LogicalCores * timeCloudsCopy[j].TmpAlloc.CPU.BaseClock)
						remainingCycles := thisApp.TaskReq.CPUCycle - executedCycles

						remainingApp.AlreadyStable = true // after starting executing, it was already stable
						remainingApp.CanMigrate = false   // cannot be migrated after starting executing
						remainingApp.TaskReq.CPUCycle = remainingCycles

					}

					remainingApps = append(remainingApps, remainingApp)
				}
			} else { // retain all services
				remainingApp := model.AppCopy(thisApp)
				remainingApp.StartTime = 0
				remainingApp.ImagePullDoneTime = 0
				remainingApp.DataInputDoneTime = 0
				remainingApp.StableTime = 0
				remainingApp.TaskCompletionTime = 0
				remainingApp.IsNew = false
				remainingApp.CloudRemainingOn = j
				remainingApp.ImagePullDone = false
				remainingApp.AlreadyStable = false
				remainingApp.CanMigrate = true

				if timeSinceLastDeploy > thisApp.ImagePullDoneTime { // after image pulling
					remainingApp.ImagePullDone = true
				}

				if timeSinceLastDeploy > thisApp.StableTime { // after startup, start executing, occupy the resource
					remainingApp.AlreadyStable = true
					timeCloudsCopy[j].TmpAlloc.CPU.LogicalCores -= thisApp.SvcReq.CPUClock / timeCloudsCopy[j].TmpAlloc.CPU.BaseClock
				}
				remainingApps = append(remainingApps, remainingApp)
			}

		}
	}

	// Then we fix the app indexes and dependence indexes
	var idxMap map[int]int = make(map[int]int)
	for i := 0; i < len(remainingApps); i++ {
		idxMap[remainingApps[i].AppIdx] = i
	}
	for i := 0; i < len(remainingApps); i++ {
		remainingApps[i].AppIdx = idxMap[remainingApps[i].AppIdx]
		for j := 0; j < len(remainingApps[i].Depend); {
			if newIdx, exist := idxMap[remainingApps[i].Depend[j].AppIdx]; exist { // the dependent app is remaining, we need to update the dependent index
				remainingApps[i].Depend[j].AppIdx = newIdx
				j++
			} else { // the dependent app is finished, we need to delete the dependence
				remainingApps[i].Depend = append(remainingApps[i].Depend[:j], remainingApps[i].Depend[j+1:]...)
			}
		}
	}

	// the dependent apps of executing tasks also cannot be migrated
	var forbidMigration func(appIdx int)
	forbidMigration = func(appIdx int) {
		remainingApps[appIdx].CanMigrate = false
		for i := 0; i < len(remainingApps[appIdx].Depend); i++ {
			forbidMigration(remainingApps[appIdx].Depend[i].AppIdx)
		}
	}

	for i := 0; i < len(remainingApps); i++ {
		if !remainingApps[i].CanMigrate {
			forbidMigration(i)
		}
	}

	return remainingApps
}

// CloudMeetApp check whether a cloud can meet an application
func CloudMeetApp(cloud model.Cloud, app model.Application) bool {

	return true
}
