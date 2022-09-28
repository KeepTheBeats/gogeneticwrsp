package model

// Resources : resources that clouds have and applications require
type Resources struct {
	CPU        float64 `json:"cpu"`        // unit logical cores
	Memory     float64 `json:"memory"`     // unit Byte (B)
	Storage    float64 `json:"storage"`    // unit Byte (B)
	NetLatency float64 `json:"netlatency"` // unit millisecond (ms)
}
