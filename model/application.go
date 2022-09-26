package model

// Application : applications that need to be scheduled
type Application struct {
	Requests Resources
	Priority uint16 // range [1, 65535]
}
