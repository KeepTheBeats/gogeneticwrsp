package experimenttools

import (
	"log"
	"math"

	"github.com/KeepTheBeats/routing-algorithms/random"

	"gogeneticwrsp/model"
)

// randomly choose a CPUClock from some CPUs
func chooseResCPU() model.CPUResource {
	cpu := cpuToChoose[random.RandomInt(0, len(cpuToChoose)-1)]
	cpuRes := model.CPUResource{
		LogicalCores: cpu.logicalCores,
		BaseClock:    cpu.baseClock,
	}
	return cpuRes
}

// generate CPU cycles needed by a task
func generateTaskCPU() float64 {
	var CPUCycle float64
	lowerBound, upperBound, miu, sigma := 129024000.00, 578604236800.00, 51419176466.20, 125585987435.47
	// from parameters in several related works
	CPUCycle = random.NormalRandomBM(lowerBound, upperBound, miu, sigma)
	return CPUCycle
}

// generate CPU clock that should be reserved for a Service
func generateSvcCPU() float64 {
	var CPUClock float64
	lowerBound, upperBound, miu, sigma := 1.0, 14.80, 3.91, 3.46 // from parameters in several related works
	CPUClock = random.NormalRandomBM(lowerBound, upperBound, miu, sigma)
	return CPUClock
}

// CPUClock logical cores,
func generateResourceCPU(lowerBound, upperBound, miu, sigma float64, forCapacity bool) float64 {
	raw_cores := random.NormalRandomBM(lowerBound, upperBound, miu, sigma)
	if !forCapacity {
		return raw_cores
	}
	// Normally, the number of CPUClock physical cores is a multiple of 2,
	// and normally, every physical core has 2 logical cores,
	// so normally, the number of CPUClock logical cores is a multiple of 4.
	multiple_of_4_cores := int(raw_cores/4) * 4
	return float64(multiple_of_4_cores)
}

// randomly choose a CPUClock from some CPUs
func generateResourceCPUCapacity(lowerBound, upperBound float64) float64 {
	return float64(random.RandomInt(int(lowerBound), int(upperBound)))
}

// memory and storage Byte, forCapacity, a power of 2
//func generateResourceMemoryStorageCapacity(powerLowerBound, powerUpperBound, miu, sigma float64) float64 {
//	power := int(random.NormalRandomBM(powerLowerBound, powerUpperBound, miu, sigma))
//	return math.Pow(2, float64(power))
//}

// randomly choose a capacity of Memory, unit B
func chooseResMem() float64 {
	return resMemToChoose[random.RandomInt(0, len(resMemToChoose)-1)] * 1024 * 1024 * 1024
}

// randomly choose a capacity of Storage, unit B
func chooseResStor() float64 {
	return resStorToChoose[random.RandomInt(0, len(resStorToChoose)-1)] * 1024 * 1024 * 1024
}

func generateResourceMemoryStorageCapacity(lowerBound, upperBound float64) float64 {
	return random.RandomFloat64(lowerBound, upperBound)
}

// randomly choose a request of Memory, unit B
func chooseReqMem() float64 {
	return reqMemToChoose[random.RandomInt(0, len(reqMemToChoose)-1)] * 1024 * 1024 * 1024
}

// randomly choose a capacity of Storage, unit B
func chooseReqStor() float64 {
	return ReqStorToChoose[random.RandomInt(0, len(ReqStorToChoose)-1)] * 1024 * 1024 * 1024
}

// memory and storage Byte, forRequest
func generateResourceMemoryStorageRequest(lowerBound, upperBound, miu, sigma float64) float64 {
	return random.NormalRandomBM(lowerBound, upperBound, miu, sigma)
}

// network latency milliseconds
func generateResourceNetLatency(power10LowerBound, power10UpperBound, miu, sigma float64) float64 {
	power10 := int(random.NormalRandomBM(power10LowerBound, power10UpperBound, miu, sigma))
	head := random.RandomInt(1, 9)
	return float64(head) * math.Pow10(power10)
}

// generate rtt between two clouds or image repository or Architecture Controller
// I have 78 data from related works and tests, 4 groups:
// group 1: index [1, 12], lowerBound upperBound miu sigma: 20 40 30 10.44465936
// group 2: index [13, 24], lowerBound upperBound miu sigma: 110 18000 4488.5 6654.286486
// group 3: index [25, 36], lowerBound upperBound miu sigma: 0.508 5.571 2.550916667 1.51254905
// group 3: index [37, 78], lowerBound upperBound miu sigma: 45.186 324.426 181.289619 90.76114479
func generateResourceRTT() float64 {
	var rtt float64
	var lowerBound, upperBound, miu, sigma float64
	var idx int = random.RandomInt(1, 78)
	switch {
	case idx >= 1 && idx <= 12:
		lowerBound, upperBound, miu, sigma = 20, 40, 30, 10.44465936
	case idx >= 13 && idx <= 24:
		lowerBound, upperBound, miu, sigma = 110, 18000, 4488.5, 6654.286486
	case idx >= 25 && idx <= 36:
		lowerBound, upperBound, miu, sigma = 0.508, 5.571, 2.550916667, 1.51254905
	case idx >= 37 && idx <= 78:
		lowerBound, upperBound, miu, sigma = 45.186, 324.426, 181.289619, 90.76114479
	default:
		log.Panicln("generateResourceRTT, unexpected idx:", idx)
	}
	rtt = random.NormalRandomBM(lowerBound, upperBound, miu, sigma)
	return rtt
}

// generate bandwidth between two clouds or image repository or Architecture Controller
func generateResourceBW() float64 {
	var bandwidth float64
	lowerBound, upperBound, miu, sigma := 0.873, 935.0, 145.2336143, 215.0395931 // from parameters in several related works and tests
	bandwidth = random.NormalRandomBM(lowerBound, upperBound, miu, sigma)
	return bandwidth
}

// choose the number of depended apps of an app
func chooseDepNum() int {
	return depNumsToChoose[random.RandomInt(0, len(depNumsToChoose)-1)]
}

// choose the required bandwidth of apps
func chooseReqBW() float64 {
	return reqBwToChoose[random.RandomInt(0, len(reqBwToChoose)-1)]
}

// choose the required Round trip time of apps
func chooseReqRTT() float64 {
	return reqRttToChoose[random.RandomInt(0, len(reqRttToChoose)-1)]
}

// generate Input Data Size for an application, unit B
func generateInputSize() float64 {
	return random.RandomFloat64(430080, 65011712) // uniform distribution
}

// generate container image size for an application, unit B
func generateImageSize() float64 {
	var imageSize float64
	// from images in the first page of https://hub.docker.com/search?image_filter=official&q=&type=image
	lowerBound, upperBound, miu, sigma := 13619.2, 1043333120.0, 343125649.8, 322164492.5
	imageSize = random.NormalRandomBM(lowerBound, upperBound, miu, sigma)
	return imageSize
}

// Priority of application range [1, 65535]
func generatePriority(lowerBound, upperBound, miu, sigma float64) uint16 {
	return uint16(random.NormalRandomBM(lowerBound, upperBound, miu, sigma))
}

// Priority of application range [1, 65535]
func generateUniformPriority(lowerBound, upperBound float64) uint16 {
	return uint16(random.RandomInt(int(lowerBound), int(upperBound)))
}

// Priority of application range [1, 65535]
func generatePowerPriority(lowerBoundPower, upperBoundPower float64) uint16 {
	power := random.RandomInt(int(lowerBoundPower), int(upperBoundPower))
	priority := math.Pow(2, float64(power))
	if priority > 65535 {
		priority = 65535
	}
	return uint16(priority)
}

// generate the number of applications in an app group
func genAppNumGroup() int {
	return random.RandomInt(4, 14) // 4 and 14 are from related works
}

// generate the time interval between 2 app groups
func genTimeIntervalGroups() float64 {
	// lowerBound cannot be 0, because if two groups are at the same time, the line charts cannot be generated according to the time
	return random.ExponentialRandom(1, math.MaxFloat64, 1.0/15.0)
}
