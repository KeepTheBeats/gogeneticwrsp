package algorithms

import (
	"gogeneticwrsp/model"
)

type FirstFit struct {
}

func NewFirstFit(clouds []model.Cloud, apps []model.Application) *FirstFit {
	return &FirstFit{}
}

func (ff *FirstFit) Schedule(clouds []model.Cloud, apps []model.Application) (model.Solution, error) {
	schedulingResult := FirstFitSchedule(clouds, apps)
	return model.Solution{SchedulingResult: schedulingResult}, nil
}

func FirstFitSchedule(clouds []model.Cloud, apps []model.Application) []int {
	var schedulingResult []int = make([]int, len(apps))

	// set all applications undeployed
	for i := 0; i < len(apps); i++ {
		schedulingResult[i] = len(clouds)
	}

	// For every app in order, choose the first cloud that can meet its requirements
	for i := 0; i < len(apps); i++ {
		for j := 0; j < len(clouds); j++ {
			if clouds[j].Allocatable.NetLatency > apps[i].SvcReq.NetLatency {
				continue
			}
			schedulingResult[i] = j
			if Acceptable(clouds, apps, schedulingResult) {
				break
			}
			schedulingResult[i] = len(clouds)
		}
	}
	return schedulingResult
}
