package model

import (
	"fmt"
)

// Application needs to be scheduled. Two types: service or task
// Services run forever.
// A task needs to do some workload and will release resources after the completion of the work.
type Application struct {
	IsTask  bool             `json:"isTask"`
	SvcReq  ServiceResources `json:"svcReq"`
	TaskReq TaskResources    `json:"taskReq"`

	InputDataSize   float64 `json:"inputDataSize"`   // unit Byte (B)
	ImageSize       float64 `json:"imageSize"`       // container image size, unit Byte (B)
	StartUpCPUCycle float64 `json:"startUpCPUCycle"` // number of CPU cycles needed during the application startup

	Priority uint16 `json:"priority"` // range [1, 65535], 2 will be much more prior to 1, because 2 is twice of 1, users should consider this when setting the priorities. Users should know how big the difference between two applications.
	AppIdx   int    `json:"appIdx"`   // index in the apps to be scheduled
	OriIdx   int    `json:"oriIdx"`   // original index in all apps

	// used in one round of scheduling
	StartTime          float64 `json:"startTime"`          // service and task have this, time duration from "the moment that all apps start to be deployed" to "the moment of this application's start", unit second
	ImagePullDoneTime  float64 `json:"imagePullDoneTime"`  // service and task have this, time duration from "the moment that all apps start to be deployed" to "the moment of this application's image pulling being done", unit second
	DataInputDoneTime  float64 `json:"dataInputDoneTime"`  // service and task have this, time duration from "the moment that all apps start to be deployed" to "the moment of this application's data input being done", unit second
	StableTime         float64 `json:"stableTime"`         // service and task have this, time duration from "the moment that all apps start to be deployed" to "the moment when the application is stable, meaning application startup being done", unit second
	TaskCompletionTime float64 `json:"taskCompletionTime"` // only task has this, time duration from "the moment that all apps start to be deployed" to "the moment of this task's completion", unit second

	// used in all rounds of scheduling
	GeneratedTime      float64 `json:"generatedTime"`      // the time when an app scheduling request is generated
	TaskFinalComplTime float64 `json:"taskFinalComplTime"` // for task, the final time point at which a task is finished
	SvcSuspensionTime  float64 `json:"SvcSuspensionTime"`  // for service, the total length of time periods when a service is suspended

	Depend []Dependence `json:"depend"` // the dependence information of this application

	// for remaining apps
	IsNew bool `json:"isNew"` // whether this application is newly coming in this round. true: newly coming in this round; false: remaining from previous rounds
	// this group of parameters are effective only when IsNew == false
	CloudRemainingOn int  `json:"cloudRemainingOn"` // which cloud that this app was scheduled to in previous rounds
	ImagePullDone    bool `json:"imagePullDone"`    // whether the image was pulled down in previous rounds
	AlreadyStable    bool `json:"alreadyStable"`    // whether the app already finished startup in previous rounds
	CanMigrate       bool `json:"canMigrate"`       // whether the app can be migrated or can only be suspended
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

// CheckDepend check whether apps[i] depends on apps[j]
func CheckDepend(apps []Application, i, j int) bool {
	for k := 0; k < len(apps[i].Depend); k++ {
		if apps[i].Depend[k].AppIdx == j {
			return true
		}
	}
	return false
}

// CombApps combine two application slices
func CombApps(a, b []Application) []Application {
	aCopy := AppsCopy(a)
	bCopy := AppsCopy(b)
	// after combination, every index in bCopy will add aCopy diff, which is len(aCopy)
	diff := len(aCopy)
	for i := 0; i < len(bCopy); i++ {
		bCopy[i].AppIdx += diff
		for j := 0; j < len(bCopy[i].Depend); j++ {
			bCopy[i].Depend[j].AppIdx += diff
		}
	}
	aCopy = append(aCopy, bCopy...)
	return aCopy
}

// AppsCopy deep copy an Application Slice
func AppsCopy(src []Application) []Application {
	var dst []Application = make([]Application, len(src))
	for i := 0; i < len(dst); i++ {
		dst[i] = AppCopy(src[i])
	}
	return dst
}

// AppCopy deep copy an Application
func AppCopy(src Application) Application {
	var dst Application = src

	dst.Depend = make([]Dependence, len(src.Depend))
	copy(dst.Depend, src.Depend)

	return dst
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
