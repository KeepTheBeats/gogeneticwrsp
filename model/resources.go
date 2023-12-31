package model

// Resources : resources that clouds have and applications require
type Resources struct {
	CPU     CPUResource `json:"cpu"`
	Memory  float64     `json:"memory"`  // unit Byte (B)
	Storage float64     `json:"storage"` // unit Byte (B)

	// network resources
	NetCondClouds     []NetworkCondition `json:"netCondClouds"`     // network condition between this cloud and every other cloud
	NetCondImage      NetworkCondition   `json:"netCondImage"`      // network condition between this cloud and image repository
	NetCondController NetworkCondition   `json:"netCondController"` // network condition between this cloud and Architecture Controller
	UpBwImage         float64            `json:"upBwImage"`         // upstream bandwidth from this cloud to image repository
	UpBwController    float64            `json:"upBwController"`    // upstream bandwidth from this cloud to Architecture Controller
}

// ResCopy deep copy a resource
func ResCopy(src Resources) Resources {
	var dst Resources = src
	dst.NetCondClouds = make([]NetworkCondition, len(src.NetCondClouds))
	copy(dst.NetCondClouds, src.NetCondClouds)
	return dst
}

type CPUResource struct {
	LogicalCores float64 `json:"logicalCores"` // number of logical cores
	BaseClock    float64 `json:"baseClock"`    // unit GHz
}

type NetworkCondition struct {
	RTT    float64 `json:"rtt"`    // Round-Trip Time, unit millisecond (ms)
	DownBw float64 `json:"doneBw"` // downstream bandwidth, unit Mb/s
}

type ServiceResources struct {
	CPUClock float64 `json:"cpuClock"` // unit GHz
	Memory   float64 `json:"memory"`   // unit Byte (B)
	Storage  float64 `json:"storage"`  // unit Byte (B)
}

type TaskResources struct {
	CPUCycle float64 `json:"cpuCycle"` // number of CPU cycles needed to execute the task
	Memory   float64 `json:"memory"`   // unit Byte (B)
	Storage  float64 `json:"storage"`  // unit Byte (B)
}

type Dependence struct {
	AppIdx int `json:"appIdx"` // the index of the dependent application

	// following items are effective only when the dependent application is a service
	DownBw float64 `json:"doneBw"` // required downstream bandwidth to the dependent service, unit Mb/s
	UpBw   float64 `json:"upBw"`   // required upstream bandwidth to the dependent service, unit Mb/s
	RTT    float64 `json:"rtt"`    // required Round-Trip Time to the dependent service, unit millisecond (ms)
}
