package experiments

import (
	"encoding/json"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"math"
	"runtime"

	"gogeneticwrsp/model"
)

func getFilePaths(numCloud, numApp int) (string, string) {
	var cloudPath, appPath string
	if runtime.GOOS == "windows" {
		cloudPath = fmt.Sprintf("%s\\src\\gogeneticwrsp\\experiments\\cloud_%d_.json", build.Default.GOPATH, numCloud)
		appPath = fmt.Sprintf("%s\\src\\gogeneticwrsp\\experiments\\app_%d_.json", build.Default.GOPATH, numApp)
	} else {
		cloudPath = fmt.Sprintf("%s/src/gogeneticwrsp/experiments/cloud_%d_.json", build.Default.GOPATH, numCloud)
		appPath = fmt.Sprintf("%s/src/gogeneticwrsp/experiments/app_%d_.json", build.Default.GOPATH, numApp)
	}
	return cloudPath, appPath
}

// GenerateCloudsApps generates clouds and apps and writes them into files.
func GenerateCloudsApps(numCloud, numApp int) {

	log.Printf("generate %d clouds and %d applications and write them into files\n", numCloud, numApp)

	var clouds []model.Cloud = make([]model.Cloud, numCloud)
	var apps []model.Application = make([]model.Application, numApp)

	// generate clouds
	for i := 0; i < numCloud; i++ {
		clouds[i].Capacity.CPU = generateResourceCPU(16, 131.9, 32, 32, true)
		clouds[i].Capacity.Memory = generateResourceMemoryStorageCapacity(34, 39.9, 36, 2)
		clouds[i].Capacity.Storage = generateResourceMemoryStorageCapacity(39, 44.9, 41, 2)
		clouds[i].Capacity.NetLatency = generateResourceNetLatency(0, 4, 2, 2)

		clouds[i].Allocatable = clouds[i].Capacity
	}

	// generate applications
	for i := 0; i < numApp; i++ {
		apps[i].Requests.CPU = generateResourceCPU(0.5, 5, 1.5, 2.5, false)
		apps[i].Requests.Memory = generateResourceMemoryStorageRequest(math.Pow(2, 25), math.Pow(2, 31), math.Pow(2, 29), 2000*math.Pow(2, 20))

		apps[i].Requests.Storage = generateResourceMemoryStorageRequest(0, math.Pow(2, 32), math.Pow(2, 30), math.Pow(2, 32))
		apps[i].Requests.NetLatency = generateResourceNetLatency(1, 5, 3, 3)
		apps[i].Priority = generatePriority(100, 65535.9, 150, 300)
	}

	cloudsJson, err := json.Marshal(clouds)
	if err != nil {
		log.Fatalln("json.Marshal(clouds) error:", err.Error())

	}
	appsJson, err := json.Marshal(apps)
	if err != nil {
		log.Fatalln("json.Marshal(apps) error:", err.Error())
	}

	var cloudPath, appPath string = getFilePaths(numCloud, numApp)

	err = ioutil.WriteFile(cloudPath, cloudsJson, 0777)
	if err != nil {
		log.Fatalln("ioutil.WriteFile(cloudPath, cloudsJson, 0777) error:", err.Error())
	}
	err = ioutil.WriteFile(appPath, appsJson, 0777)
	if err != nil {
		log.Fatalln("ioutil.WriteFile(appPath, appsJson, 0777) error:", err.Error())
	}
}

// ReadCloudsApps from files
func ReadCloudsApps(numCloud, numApp int) ([]model.Cloud, []model.Application) {
	var cloudPath, appPath string = getFilePaths(numCloud, numApp)
	var clouds []model.Cloud
	var apps []model.Application

	cloudsJson, err := ioutil.ReadFile(cloudPath)
	if err != nil {
		log.Fatalln("ioutil.ReadFile(cloudPath) error:", err.Error())
	}
	appsJson, err := ioutil.ReadFile(appPath)
	if err != nil {
		log.Fatalln("ioutil.ReadFile(appPath) error:", err.Error())
	}

	err = json.Unmarshal(cloudsJson, &clouds)
	if err != nil {
		log.Fatalln("json.Unmarshal(cloudsJson, &clouds) error:", err.Error())
	}
	err = json.Unmarshal(appsJson, &apps)
	if err != nil {
		log.Fatalln("json.Unmarshal(appsJson, &apps) error:", err.Error())
	}

	return clouds, apps
}
