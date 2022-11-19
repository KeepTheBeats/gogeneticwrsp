package model

import (
	"time"
)

// Cloud : clouds that applications can be scheduled to
type Cloud struct {
	Capacity           Resources     `json:"capacity"`    // total resources
	Allocatable        Resources     `json:"allocatable"` // allocatable resources
	TmpAlloc           Resources     `json:"tmpAlloc"`    // temporary allocatable resources, for temporary record during scheduling
	RunningApps        []Application `json:"runningApps"`
	TotalTaskComplTime float64       `json:"totalTaskComplTime"`
	UpdateTime         time.Time     `json:"updateTime"`
}

// CloudsCopy deep copy a Cloud Slice
func CloudsCopy(src []Cloud) []Cloud {
	var dst []Cloud = make([]Cloud, len(src))
	for i := 0; i < len(dst); i++ {
		dst[i] = CloudCopy(src[i])
	}
	return dst
}

// CloudCopy deep copy a Cloud
func CloudCopy(src Cloud) Cloud {
	var dst Cloud = src
	dst.Capacity = ResCopy(src.Capacity)
	dst.Allocatable = ResCopy(src.Allocatable)
	dst.TmpAlloc = ResCopy(src.TmpAlloc)
	dst.RunningApps = AppsCopy(src.RunningApps)
	return dst
}
