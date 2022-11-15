package model

import "time"

// Cloud : clouds that applications can be scheduled to
type Cloud struct {
	Capacity           Resources     `json:"capacity"`
	Allocatable        Resources     `json:"allocatable"`
	RunningApps        []Application `json:"runningApps"`
	TotalTaskComplTime float64       `json:"totalTaskComplTime"`
	UpdateTime         time.Time     `json:"updateTime"`
}
