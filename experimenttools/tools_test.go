package experimenttools

import (
	"fmt"
	"math"
	"testing"
)

func TestInnerGenerateResourceCPU(t *testing.T) {
	fmt.Println("Generate 40 CPU resources for requests:")
	for i := 0; i < 40; i++ {
		fmt.Println(generateResourceCPU(0.5, 5, 1.5, 2.5, false))
	}
	fmt.Println("Generate 40 CPU resources for capacities:")
	for i := 0; i < 40; i++ {
		fmt.Println(generateResourceCPU(16, 131.9, 32, 32, true))
	}
}

func TestInnerGenerateResourceMemoryStorageCapacity(t *testing.T) {
	fmt.Println("Generate 40 Memory resources for capacities:")
	for i := 0; i < 40; i++ {
		fmt.Println(generateResourceMemoryStorageCapacity(34, 39.9, 36, 2)/math.Pow(2, 30), "GiB")
	}
	fmt.Println("Generate 40 Storage resources for capacities:")
	for i := 0; i < 40; i++ {
		fmt.Println(generateResourceMemoryStorageCapacity(39, 44.9, 41, 2)/math.Pow(2, 30), "GiB")
	}
}

func TestInnerGenerateResourceMemoryStorageRequest(t *testing.T) {
	fmt.Println("Generate 40 Memory resources for requests:")
	for i := 0; i < 40; i++ {
		fmt.Println(generateResourceMemoryStorageRequest(math.Pow(2, 25), math.Pow(2, 31), math.Pow(2, 29), 2000*math.Pow(2, 20))/math.Pow(2, 20), "MiB")
	}
	fmt.Println("Generate 40 Storage resources for requests:")
	for i := 0; i < 40; i++ {
		fmt.Println(generateResourceMemoryStorageRequest(0, math.Pow(2, 32), math.Pow(2, 30), math.Pow(2, 32))/math.Pow(2, 30), "GiB")
	}
}

func TestInnerGenerateResourceNetLatency(t *testing.T) {
	fmt.Println("Generate 40 NetLatency resources for capacities:")
	for i := 0; i < 40; i++ {
		fmt.Println(generateResourceNetLatency(0, 4, 2, 2), "ms")
	}
	fmt.Println("Generate 40 NetLatency resources for requests:")
	for i := 0; i < 40; i++ {
		fmt.Println(generateResourceNetLatency(1, 5, 3, 3), "ms")
	}
}

func TestInnerGeneratePriority(t *testing.T) {
	fmt.Println("Generate 40 Priority values of application:")
	for i := 0; i < 40; i++ {
		fmt.Println(generatePriority(100, 65535.9, 150, 300))
	}
}
