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

type AppResources struct {
	CPU        float64 `json:"cpu"`        // unit number of logical cores
	Memory     float64 `json:"memory"`     // unit Byte (B)
	Storage    float64 `json:"storage"`    // unit Byte (B)
	NetLatency float64 `json:"netLatency"` // unit millisecond (ms)
}
