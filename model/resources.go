package model

// Resources : resources that clouds have and applications require
type Resources struct {
	CPU        CPUResource `json:"cpu"`
	Memory     float64     `json:"memory"`     // unit Byte (B)
	Storage    float64     `json:"storage"`    // unit Byte (B)
	NetLatency float64     `json:"netLatency"` // unit millisecond (ms)
}

type CPUResource struct {
	LogicalCores float64 `json:"logicalCores"` // number of logical cores
	BaseClock    float64 `json:"baseClock"`    // unit GHz
}

type ServiceResources struct {
	CPUClock   float64 `json:"cpuClock"`   // unit GHz
	Memory     float64 `json:"memory"`     // unit Byte (B)
	Storage    float64 `json:"storage"`    // unit Byte (B)
	NetLatency float64 `json:"netLatency"` // unit millisecond (ms)
}

type TaskResources struct {
	CPUCycle   float64 `json:"cpuCycle"`   // unit number of CPU cycles needed to execute the task
	Memory     float64 `json:"memory"`     // unit Byte (B)
	Storage    float64 `json:"storage"`    // unit Byte (B)
	NetLatency float64 `json:"netLatency"` // unit millisecond (ms)
}
