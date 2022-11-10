package model

// Application : applications that need to be scheduled
type Application struct {
	Requests AppResources `json:"appRequests"`
	Priority uint16       `json:"priority"` // range [100, 65535], if the range is [1, 65535], the 2 will be much more prior to 1, because 2 is 2 times 1
}
