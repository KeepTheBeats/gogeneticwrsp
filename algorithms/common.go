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
		deployedClouds[solution.SchedulingResult[appIndex]].Allocatable.CPU -= apps[appIndex].Requests.CPU
		deployedClouds[solution.SchedulingResult[appIndex]].Allocatable.Memory -= apps[appIndex].Requests.Memory
		deployedClouds[solution.SchedulingResult[appIndex]].Allocatable.Storage -= apps[appIndex].Requests.Storage
		// NetLatency does not need to be subtracted, but if we use NetBandwidth, we need to subtract it.
	}
	return deployedClouds
}
