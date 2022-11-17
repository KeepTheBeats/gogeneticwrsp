package model

import (
	"fmt"
)

// Application needs to be scheduled. Two types: service or task
// Services run forever.
// A task needs to do some workload and will release resources after the completion of the work.
type Application struct {
	IsTask             bool             `json:"isTask"`
	SvcReq             ServiceResources `json:"svcReq"`
	TaskReq            TaskResources    `json:"taskReq"`
	Priority           uint16           `json:"priority"`           // range [100, 65535], if the range is [1, 65535], the 2 will be much more prior to 1, because 2 is 2 times 1
	AppIdx             int              `json:"appIdx"`             // index in all apps to be scheduled
	TaskCompletionTime float64          `json:"taskCompletionTime"` // only task has this, time duration from "the moment that all apps start to be deployed" to "the moment of this task's completion"
	Depend             []Dependence     `json:"depend"`             // the dependence information of this application
}

// DependencyValid Check whether Dependency is Valid
func DependencyValid(apps []Application) error {
	// If app A depends on app B, priority(A) must be smaller than priority(B)
	for i := 0; i < len(apps); i++ {
		for j := 0; j < len(apps[i].Depend); j++ {
			if apps[i].Priority >= apps[apps[i].Depend[j].AppIdx].Priority {
				return fmt.Errorf("app index %d with priority %d depend on app index %d with priority %d", i, apps[i].Priority, apps[i].Depend[j].AppIdx, apps[apps[i].Depend[j].AppIdx].Priority)
			}
		}
	}
	return nil
}

type AppSlice []Application

func (as AppSlice) Len() int {
	return len(as)
}

func (as AppSlice) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}

func (as AppSlice) Less(i, j int) bool {
	return as[i].Priority > as[j].Priority
}
