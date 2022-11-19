package model

import (
	"fmt"
	"testing"
)

func TestResCopy(t *testing.T) {
	var cloud Cloud
	cloud.Capacity = Resources{
		CPU: CPUResource{
			1,
			1,
		},
		Memory:  1,
		Storage: 1,
		NetCondClouds: []NetworkCondition{
			NetworkCondition{
				RTT:    1,
				DownBw: 1,
			}, NetworkCondition{
				RTT:    1,
				DownBw: 1,
			},
		},
		NetCondImage: NetworkCondition{
			RTT:    1,
			DownBw: 1,
		},
		NetCondController: NetworkCondition{
			RTT:    1,
			DownBw: 1,
		},
		UpBwImage:      1,
		UpBwController: 1,
	}
	cloud.TmpAlloc = ResCopy(cloud.Capacity)
	//cloud.TmpAlloc = cloud.Capacity
	fmt.Println(cloud.Capacity)
	fmt.Println(cloud.TmpAlloc)
	cloud.Capacity.CPU.LogicalCores += 1
	cloud.Capacity.CPU.BaseClock += 1
	cloud.Capacity.NetCondClouds[0].RTT += 1
	fmt.Println(cloud.Capacity)
	fmt.Println(cloud.TmpAlloc)
}
