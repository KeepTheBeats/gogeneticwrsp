package experiments

import (
	"fmt"
	"gogeneticwrsp/model"
)

func Cloud4App40() {
	var numCloud int = 4
	var numApp int = 40

	fmt.Printf("Execute experiment with %d clouds and %d applications!\n", numCloud, numApp)

	var clouds []model.Cloud = make([]model.Cloud, numCloud)
	var apps []model.Application = make([]model.Application, numApp)

	fmt.Println(clouds)
	fmt.Println(len(clouds))
	fmt.Println(apps)
	fmt.Println(len(apps))

	// generate clouds
	for i := 0; i < numCloud; i++ {

	}

	// generate applications
	for i := 0; i < numApp; i++ {

	}
}
