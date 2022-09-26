package model

// Resources : resources that clouds have and applications require
type Resources struct {
	CPU        float64 // unit logical cores
	Memory     float64 // unit Byte (B)
	Storage    float64 // unit Byte (B)
	NetLatency float64 // unit millisecond (ms)
}
