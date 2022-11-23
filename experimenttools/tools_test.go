package experimenttools

import (
	"fmt"
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
	fmt.Println("Generate 40 Priority values of application:")
	for i := 0; i < 40; i++ {
		fmt.Println(generatePriority(100, 65535.9, 150, 300))
	}
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
