package model

// Cloud : clouds that applications can be scheduled to
type Cloud struct {
	Capacity    Resources `json:"capacity"`
	Allocatable Resources `json:"allocatable"`
}
