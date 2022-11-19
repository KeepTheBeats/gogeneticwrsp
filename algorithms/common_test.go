package algorithms

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"gogeneticwrsp/model"
)

func forTestClouds1(t *testing.T) []model.Cloud {
	var clouds []model.Cloud
	cloudsJson := []byte(`[{"capacity":{"cpu":64,"memory":68719476736,"storage":17592186044416,"netlatency":500},"allocatable":{"cpu":64,"memory":68719476736,"storage":17592186044416,"netlatency":500}},{"capacity":{"cpu":100,"memory":68719476736,"storage":549755813888,"netlatency":900},"allocatable":{"cpu":100,"memory":68719476736,"storage":549755813888,"netlatency":900}},{"capacity":{"cpu":32,"memory":137438953472,"storage":4398046511104,"netlatency":3},"allocatable":{"cpu":32,"memory":137438953472,"storage":4398046511104,"netlatency":3}},{"capacity":{"cpu":116,"memory":137438953472,"storage":2199023255552,"netlatency":100},"allocatable":{"cpu":116,"memory":137438953472,"storage":2199023255552,"netlatency":100}}]`)
	err := json.Unmarshal(cloudsJson, &clouds)
	if err != nil {
		t.Fatalf("json.Unmarshal(cloudsJson, &clouds) error: %s", err.Error())
	}
	return clouds
}

func forTestClouds2(t *testing.T) []model.Cloud {
	var clouds []model.Cloud
	cloudsJson := []byte(fmt.Sprintf(`[{"capacity":{"cpu":64,"memory":68719476736,"storage":17592186044416,"netlatency":500},"allocatable":{"cpu":%g,"memory":%g,"storage":%g,"netlatency":500}},{"capacity":{"cpu":100,"memory":68719476736,"storage":549755813888,"netlatency":900},"allocatable":{"cpu":%g,"memory":%g,"storage":%g,"netlatency":900}},{"capacity":{"cpu":32,"memory":137438953472,"storage":4398046511104,"netlatency":3},"allocatable":{"cpu":%g,"memory":%g,"storage":%g,"netlatency":3}},{"capacity":{"cpu":116,"memory":137438953472,"storage":2199023255552,"netlatency":100},"allocatable":{"cpu":%g,"memory":%g,"storage":%g,"netlatency":100}}]`, 64-1.7891381895587495, 68719476736-1683896794.0429783, 17592186044416-1161094647.719078, 100-2.6318879938329616, 68719476736-848307113.353564, 549755813888-30701192.75704515, 32-3.256552011508981, 137438953472-1689436976.1793702, 4398046511104-2063899709.206542, 116-3.489322191917918, 137438953472-1808688469.6637993, 2199023255552-2261225728.8706594))
	err := json.Unmarshal(cloudsJson, &clouds)
	if err != nil {
		t.Fatalf("json.Unmarshal(cloudsJson, &clouds) error: %s", err.Error())
	}
	return clouds
}

func forTestClouds3(t *testing.T) []model.Cloud {
	var clouds []model.Cloud
	cloudsJson := []byte(fmt.Sprintf(`[{"capacity":{"cpu":64,"memory":68719476736,"storage":17592186044416,"netlatency":500},"allocatable":{"cpu":%g,"memory":%g,"storage":%g,"netlatency":500}},{"capacity":{"cpu":100,"memory":68719476736,"storage":549755813888,"netlatency":900},"allocatable":{"cpu":%v,"memory":%v,"storage":%v,"netlatency":900}},{"capacity":{"cpu":32,"memory":137438953472,"storage":4398046511104,"netlatency":3},"allocatable":{"cpu":%v,"memory":%v,"storage":%v,"netlatency":3}},{"capacity":{"cpu":116,"memory":137438953472,"storage":2199023255552,"netlatency":100},"allocatable":{"cpu":%v,"memory":%v,"storage":%v,"netlatency":100}}]`, 64-1.7891381895587495-2.6318879938329616-3.256552011508981-3.489322191917918, 68719476736-1683896794.0429783-848307113.353564-1689436976.1793702-1808688469.6637993, 17592186044416-1161094647.719078-30701192.75704515-2063899709.206542-2261225728.8706594, 100, 68719476736, 549755813888, 32, 137438953472, 4398046511104, 116, 137438953472, 2199023255552))
	err := json.Unmarshal(cloudsJson, &clouds)
	if err != nil {
		t.Fatalf("json.Unmarshal(cloudsJson, &clouds) error: %s", err.Error())
	}
	return clouds
}

func forTestClouds4(t *testing.T) []model.Cloud {
	var clouds []model.Cloud
	cloudsJson := []byte(fmt.Sprintf(`[{"capacity":{"cpu":64,"memory":68719476736,"storage":17592186044416,"netlatency":500},"allocatable":{"cpu":%g,"memory":%g,"storage":%g,"netlatency":500}},{"capacity":{"cpu":100,"memory":68719476736,"storage":549755813888,"netlatency":900},"allocatable":{"cpu":%v,"memory":%v,"storage":%v,"netlatency":900}},{"capacity":{"cpu":32,"memory":137438953472,"storage":4398046511104,"netlatency":3},"allocatable":{"cpu":%v,"memory":%v,"storage":%v,"netlatency":3}},{"capacity":{"cpu":116,"memory":137438953472,"storage":2199023255552,"netlatency":100},"allocatable":{"cpu":%v,"memory":%v,"storage":%v,"netlatency":100}}]`, 64-1.7891381895587495-2.6318879938329616-3.256552011508981, 68719476736-1683896794.0429783-848307113.353564-1689436976.1793702, 17592186044416-1161094647.719078-30701192.75704515-2063899709.206542, 100, 68719476736, 549755813888, 32, 137438953472, 4398046511104, 116, 137438953472, 2199023255552))
	err := json.Unmarshal(cloudsJson, &clouds)
	if err != nil {
		t.Fatalf("json.Unmarshal(cloudsJson, &clouds) error: %s", err.Error())
	}
	return clouds
}

func forTestClouds5(t *testing.T) []model.Cloud {
	var clouds []model.Cloud
	cloudsJson := []byte(fmt.Sprintf(`[{"capacity":{"cpu":64,"memory":68719476736,"storage":17592186044416,"netlatency":500},"allocatable":{"cpu":%g,"memory":%g,"storage":%g,"netlatency":500}},{"capacity":{"cpu":100,"memory":68719476736,"storage":549755813888,"netlatency":900},"allocatable":{"cpu":%v,"memory":%v,"storage":%v,"netlatency":900}},{"capacity":{"cpu":32,"memory":137438953472,"storage":4398046511104,"netlatency":3},"allocatable":{"cpu":%v,"memory":%v,"storage":%v,"netlatency":3}},{"capacity":{"cpu":116,"memory":137438953472,"storage":2199023255552,"netlatency":100},"allocatable":{"cpu":%v,"memory":%v,"storage":%v,"netlatency":100}}]`, 64-1.7891381895587495-2.6318879938329616, 68719476736-1683896794.0429783-848307113.353564, 17592186044416-1161094647.719078-30701192.75704515, 100, 68719476736, 549755813888, 32, 137438953472, 4398046511104, 116, 137438953472, 2199023255552))
	err := json.Unmarshal(cloudsJson, &clouds)
	if err != nil {
		t.Fatalf("json.Unmarshal(cloudsJson, &clouds) error: %s", err.Error())
	}
	return clouds
}

func forTestApps1(t *testing.T) []model.Application {
	var apps []model.Application
	appsJson := []byte(`[{"requests":{"cpu":1.7891381895587495,"memory":1683896794.0429783,"storage":1161094647.719078,"netlatency":900},"priority":171},{"requests":{"cpu":2.6318879938329616,"memory":848307113.353564,"storage":30701192.75704515,"netlatency":100},"priority":741},{"requests":{"cpu":3.256552011508981,"memory":1689436976.1793702,"storage":2063899709.206542,"netlatency":400},"priority":298},{"requests":{"cpu":3.489322191917918,"memory":1808688469.6637993,"storage":2261225728.8706594,"netlatency":9000},"priority":524}]`)
	err := json.Unmarshal(appsJson, &apps)
	if err != nil {
		t.Fatalf("json.Unmarshal(appsJson, &apps) error: %s", err.Error())
	}
	return apps
}

func TestSimulateDeploy(t *testing.T) {
	testCases := []struct {
		name           string
		clouds         []model.Cloud
		apps           []model.Application
		solution       model.Solution
		expectedResult []model.Cloud
	}{
		{
			name:   "case1",
			clouds: forTestClouds1(t),
			apps:   forTestApps1(t),
			solution: model.Solution{
				SchedulingResult: []int{0, 1, 2, 3},
			},
			expectedResult: forTestClouds2(t),
		},
		{
			name:   "case2",
			clouds: forTestClouds1(t),
			apps:   forTestApps1(t),
			solution: model.Solution{
				SchedulingResult: []int{0, 0, 0, 0},
			},
			expectedResult: forTestClouds3(t),
		},
		{
			name:   "case3",
			clouds: forTestClouds1(t),
			apps:   forTestApps1(t),
			solution: model.Solution{
				SchedulingResult: []int{0, 0, 0, 4},
			},
			expectedResult: forTestClouds4(t),
		},
		{
			name:   "case4",
			clouds: forTestClouds1(t),
			apps:   forTestApps1(t),
			solution: model.Solution{
				SchedulingResult: []int{0, 0, 4, 4},
			},
			expectedResult: forTestClouds5(t),
		},
	}
	for _, testCase := range testCases {
		t.Logf("test: %s", testCase.name)
		actualResult := SimulateDeploy(testCase.clouds, testCase.apps, testCase.solution)
		assert.Equal(t, testCase.expectedResult, actualResult, fmt.Sprintf("%s: result is not expected", testCase.name))
	}
}

func TestAcceptable(t *testing.T) {
	var apps []model.Application = []model.Application{
		model.Application{
			IsTask: false,
			SvcReq: model.ServiceResources{
				CPUClock: 50,
				Memory:   50,
				Storage:  50,
			},
			Priority: 500,
			AppIdx:   0,
		},
		model.Application{
			IsTask: false,
			SvcReq: model.ServiceResources{
				CPUClock: 15,
				Memory:   15,
				Storage:  15,
			},
			Priority: 1000,
			AppIdx:   1,
		},
		model.Application{
			IsTask: false,
			SvcReq: model.ServiceResources{
				CPUClock: 135,
				Memory:   135,
				Storage:  135,
			},
			Priority: 750,
			AppIdx:   2,
		},
		model.Application{
			IsTask: false,
			SvcReq: model.ServiceResources{
				CPUClock: 1352,
				Memory:   1235,
				Storage:  1325,
			},
			Priority: 751,
			AppIdx:   3,
		},
		model.Application{
			IsTask: false,
			SvcReq: model.ServiceResources{
				CPUClock: 352,
				Memory:   235,
				Storage:  325,
			},
			Priority: 8751,
			AppIdx:   4,
		},
		model.Application{
			IsTask: true,
			TaskReq: model.TaskResources{
				CPUCycle: 32,
				Memory:   25,
				Storage:  35,
			},
			Priority: 851,
			AppIdx:   5,
		},
		model.Application{
			IsTask: true,
			TaskReq: model.TaskResources{
				CPUCycle: 832,
				Memory:   825,
				Storage:  835,
			},
			Priority: 810,
			AppIdx:   6,
		},
		model.Application{
			IsTask: true,
			TaskReq: model.TaskResources{
				CPUCycle: 83,
				Memory:   82,
				Storage:  83,
			},
			Priority: 111,
			AppIdx:   7,
		},
	}

	var clouds []model.Cloud = []model.Cloud{
		model.Cloud{
			Capacity: model.Resources{
				CPU: model.CPUResource{
					LogicalCores: 20,
					BaseClock:    10000,
				},
				Memory:  10000,
				Storage: 10000,
			},
			Allocatable: model.Resources{
				CPU: model.CPUResource{
					LogicalCores: 20,
					BaseClock:    10000,
				},
				Memory:  10000,
				Storage: 10000,
			},
		},
		model.Cloud{
			Capacity: model.Resources{
				CPU: model.CPUResource{
					LogicalCores: 45,
					BaseClock:    10000,
				},
				Memory:  10000,
				Storage: 10000,
			},
			Allocatable: model.Resources{
				CPU: model.CPUResource{
					LogicalCores: 20,
					BaseClock:    10000,
				},
				Memory:  10000,
				Storage: 10000,
			},
		},
	}

	var schedulingResult []int = []int{1, 1, 1, 1, 1, 1, 1, 1}

	fmt.Println(Acceptable(clouds, apps, schedulingResult))
}
