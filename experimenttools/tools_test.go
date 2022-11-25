package experimenttools

import (
	"fmt"
	"math"
	"testing"
)

func TestPrintChooseResCPU(t *testing.T) {
	for i := 0; i < len(cpuToChoose); i++ {
		fmt.Println(cpuToChoose[i].baseClock * cpuToChoose[i].logicalCores)
	}
}

func TestInnerChooseResCPU(t *testing.T) {
	fmt.Println("Generate 40 CPUClock:")
	for i := 0; i < 40; i++ {
		fmt.Println(chooseResCPU())
	}
}

func TestInnerChooseDepNum(t *testing.T) {
	fmt.Println("Generate 40 Dependence Number:")
	for i := 0; i < 40; i++ {
		fmt.Println(chooseDepNum())
	}
}

func TestInnerGenerateTaskCPU(t *testing.T) {
	fmt.Println("Generate 40 task CPU:")
	for i := 0; i < 40; i++ {
		fmt.Printf("%gG cycles\n", generateTaskCPU()/1024/1024/1024)
	}
}

func TestInnerGenerateSvcCPU(t *testing.T) {
	fmt.Println("Generate 40 service CPU:")
	for i := 0; i < 40; i++ {
		fmt.Println(generateSvcCPU())
	}
}

func TestInnerGenerateResourceRTT(t *testing.T) {
	fmt.Println("Generate 40 rtt:")
	for i := 0; i < 40; i++ {
		fmt.Println(generateResourceRTT())
	}
}

func TestInnerGenerateResourceBW(t *testing.T) {
	fmt.Println("Generate 40 bandwidth:")
	for i := 0; i < 40; i++ {
		fmt.Println(generateResourceBW())
	}
}

//func TestInnerGenerateResourceCPU(t *testing.T) {
//	fmt.Println("Generate 40 CPU resources for requests:")
//	for i := 0; i < 40; i++ {
//		fmt.Println(generateResourceCPU(0.5, 5, 1.5, 2.5, false))
//	}
//	fmt.Println("Generate 40 CPU resources for capacities:")
//	for i := 0; i < 40; i++ {
//		fmt.Println(generateResourceCPU(16, 131.9, 32, 32, true))
//	}
//}

//func TestInnerGenerateResourceMemoryStorageCapacity(t *testing.T) {
//	fmt.Println("Generate 40 Memory resources for capacities:")
//	for i := 0; i < 40; i++ {
//		fmt.Println(generateResourceMemoryStorageCapacity(math.Pow(2, 34), math.Pow(2, 39.9))/math.Pow(2, 30), "GiB")
//	}
//	fmt.Println("Generate 40 Storage resources for capacities:")
//	for i := 0; i < 40; i++ {
//		fmt.Println(generateResourceMemoryStorageCapacity(math.Pow(2, 39), math.Pow(2, 44.9))/math.Pow(2, 30), "GiB")
//	}
//}

//func TestInnerGenerateResourceMemoryStorageRequest(t *testing.T) {
//	fmt.Println("Generate 40 Memory resources for requests:")
//	for i := 0; i < 40; i++ {
//		fmt.Println(generateResourceMemoryStorageRequest(math.Pow(2, 25), math.Pow(2, 31), math.Pow(2, 29), 2000*math.Pow(2, 20))/math.Pow(2, 20), "MiB")
//	}
//	fmt.Println("Generate 40 Storage resources for requests:")
//	for i := 0; i < 40; i++ {
//		fmt.Println(generateResourceMemoryStorageRequest(0, math.Pow(2, 32), math.Pow(2, 30), math.Pow(2, 32))/math.Pow(2, 30), "GiB")
//	}
//}

//func TestInnerGenerateResourceNetLatency(t *testing.T) {
//	fmt.Println("Generate 40 NetLatency resources for capacities:")
//	for i := 0; i < 40; i++ {
//		fmt.Println(generateResourceNetLatency(0, 4, 2, 2), "ms")
//	}
//	fmt.Println("Generate 40 NetLatency resources for requests:")
//	for i := 0; i < 40; i++ {
//		fmt.Println(generateResourceNetLatency(1, 5, 3, 3), "ms")
//	}
//}

func TestInnerGeneratePriority(t *testing.T) {
	fmt.Println("Generate 40 uniform Priority values of application:")
	var priorities []float64
	for i := 0; i < 10000; i++ {
		p := generateUniformPriority(10, 65535)
		priorities = append(priorities, float64(p))
		//fmt.Println(generateUniformPriority(10, 65535))
	}
	fmt.Println("mean:", mean(priorities))
	fmt.Println("std:", std(priorities))
	fmt.Println("Generate 40 Priority values of application:")
	for i := 0; i < 40; i++ {
		fmt.Println(generatePriority(10, 65535, 32963.604, 18811.737890349978))
	}
	fmt.Println("Generate 40 power Priority values of application:")
	for i := 0; i < 40; i++ {
		fmt.Println(generatePowerPriority(0, 16))
	}
}

func mean(a []float64) float64 {
	var sum float64 = 0
	for i := 0; i < len(a); i++ {
		sum += a[i]
	}
	return sum / float64(len(a))
}

func variance(v []float64) float64 {
	var res float64 = 0
	var m = mean(v)
	var n int = len(v)
	for i := 0; i < n; i++ {
		res += (v[i] - m) * (v[i] - m)
	}
	return res / float64(n-1)
}

func std(v []float64) float64 {
	return math.Sqrt(variance(v))
}

func TestInnerGenAppNumGroup(t *testing.T) {
	fmt.Println("Generate 40 numbers of applications in an app group:")
	for i := 0; i < 40; i++ {
		fmt.Println(genAppNumGroup())
	}
}

func TestInnerGenTimeIntervalGroups(t *testing.T) {
	fmt.Println("Generate 40 time intervals between 2 app groups:")
	for i := 0; i < 40; i++ {
		fmt.Println(genTimeIntervalGroups())
	}
}
