package experiments

import (
	"math"

	"github.com/KeepTheBeats/routing-algorithms/random"
)

// CPU logical cores,
func generateResourceCPU(lowerBound, upperBound, miu, sigma float64, forCapacity bool) float64 {
	raw_cores := random.NormalRandomBM(lowerBound, upperBound, miu, sigma)
	if !forCapacity {
		return raw_cores
	}
	// Normally, the number of CPU physical cores is a multiple of 2,
	// and normally, every physical core has 2 logical cores,
	// so normally, the number of CPU logical cores is a multiple of 4.
	multiple_of_4_cores := int(raw_cores/4) * 4
	return float64(multiple_of_4_cores)
}

// memory and storage Byte, forCapacity, a power of 2
func generateResourceMemoryStorageCapacity(powerLowerBound, powerUpperBound, miu, sigma float64) float64 {
	power := int(random.NormalRandomBM(powerLowerBound, powerUpperBound, miu, sigma))
	return math.Pow(2, float64(power))
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

// Priority of application range [100, 65535]
func generatePriority(lowerBound, upperBound, miu, sigma float64) uint16 {
	return uint16(random.NormalRandomBM(lowerBound, upperBound, miu, sigma))
}
