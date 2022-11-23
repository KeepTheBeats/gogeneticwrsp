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

	// set all new applications undeployed, old application to their previous clouds
	var noMigrate map[int]struct{} = make(map[int]struct{})
	for i := 0; i < len(apps); i++ {
		if apps[i].IsNew { // new apps are allowed to be rejected
			schedulingResult[i] = len(clouds)
		} else { // remaining apps are not allowed to be rejected
			schedulingResult[i] = apps[i].CloudRemainingOn
			if !apps[i].CanMigrate { // executing tasks and their dependent apps cannot be migrated
				noMigrate[i] = struct{}{}
			}
		}
	}

	// For every app in order, choose the first cloud that can meet its requirements
	for i := 0; i < len(apps); i++ {
		if _, exist := noMigrate[i]; exist {
			continue
		}
		for j := 0; j < len(clouds); j++ {
			if !CloudMeetApp(clouds[j], apps[i]) {
				continue
			}
			schedulingResult[i] = j
			if Acceptable(clouds, apps, schedulingResult) {
				break
			}
			if apps[i].IsNew {
				schedulingResult[i] = len(clouds)
			} else { // old apps cannot be rejected
				schedulingResult[i] = apps[i].CloudRemainingOn
			}
		}
	}
	return schedulingResult
}
