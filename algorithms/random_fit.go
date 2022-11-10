package algorithms

import (
	"github.com/KeepTheBeats/routing-algorithms/random"
	"gogeneticwrsp/model"
)

type RandomFit struct {
}

func NewRandomFit(clouds []model.Cloud, apps []model.Application) *RandomFit {
	return &RandomFit{}
}

func (rf *RandomFit) Schedule(clouds []model.Cloud, apps []model.Application) (model.Solution, error) {
	schedulingResult := RandomFitSchedule(clouds, apps)
	return model.Solution{SchedulingResult: schedulingResult}, nil
}

func RandomFitSchedule(clouds []model.Cloud, apps []model.Application) []int {
	var schedulingResult []int = make([]int, len(apps))

	// set all applications undeployed
	for i := 0; i < len(apps); i++ {
		schedulingResult[i] = len(clouds)
	}

	// traverse apps in random order
	var undeployed []int = make([]int, len(apps))
	for i := 0; i < len(apps); i++ {
		undeployed[i] = i // record the original index of undeployed apps
	}

	for len(undeployed) > 0 {
		appIndex := random.RandomInt(0, len(undeployed)-1) // appIndex in undeployed

		// traverse clouds in random order
		var untried []int = make([]int, len(clouds))
		for i := 0; i < len(untried); i++ {
			untried[i] = i // record the original index of untried clouds
		}

		for len(untried) > 0 {
			cloudIndex := random.RandomInt(0, len(untried)-1) // cloudIndex in untried
			if clouds[untried[cloudIndex]].Allocatable.NetLatency > apps[undeployed[appIndex]].Requests.NetLatency {
				untried = append(untried[:cloudIndex], untried[cloudIndex+1:]...)
				continue
			}
			schedulingResult[undeployed[appIndex]] = untried[cloudIndex]

			if Acceptable(clouds, apps, schedulingResult) {
				untried = append(untried[:cloudIndex], untried[cloudIndex+1:]...)
				break
			}
			schedulingResult[undeployed[appIndex]] = len(clouds)
			untried = append(untried[:cloudIndex], untried[cloudIndex+1:]...)
		}

		undeployed = append(undeployed[:appIndex], undeployed[appIndex+1:]...)
	}

	return schedulingResult
}
