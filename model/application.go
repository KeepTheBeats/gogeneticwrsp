package model

// Application needs to be scheduled. Two types: service or task
// Services run forever.
// A task needs to do some workload and will release resources after the completion of the work.
type Application struct {
	Type     bool             `json:"type"`
	SvcReq   ServiceResources `json:"svcReq"`
	TaskReq  TaskResources    `json:"taskReq"`
	Priority uint16           `json:"priority"` // range [100, 65535], if the range is [1, 65535], the 2 will be much more prior to 1, because 2 is 2 times 1
}
