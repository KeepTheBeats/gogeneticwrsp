package model

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestSortApps(t *testing.T) {
	var apps []Application = []Application{
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock: 50,
				Memory:   50,
				Storage:  50,
			},
			Priority: 500,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock: 15,
				Memory:   15,
				Storage:  15,
			},
			Priority: 1000,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock: 135,
				Memory:   135,
				Storage:  135,
			},
			Priority: 750,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock: 1352,
				Memory:   1235,
				Storage:  1325,
			},
			Priority: 751,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock: 352,
				Memory:   235,
				Storage:  325,
			},
			Priority: 8751,
		},
		Application{
			IsTask: true,
			TaskReq: TaskResources{
				CPUCycle: 32,
				Memory:   25,
				Storage:  35,
			},
			Priority: 851,
		},
		Application{
			IsTask: true,
			TaskReq: TaskResources{
				CPUCycle: 832,
				Memory:   825,
				Storage:  835,
			},
			Priority: 810,
		},
		Application{
			IsTask: true,
			TaskReq: TaskResources{
				CPUCycle: 83,
				Memory:   82,
				Storage:  83,
			},
			Priority: 111,
		},
	}

	var apps2 []Application = AppsCopy(apps)

	for i := 0; i < len(apps); i++ {
		fmt.Println(apps[i])
	}

	fmt.Println("-------------sort---------------")
	sort.Sort(AppSlice(apps))

	for i := 0; i < len(apps); i++ {
		fmt.Println(apps[i])
	}
	fmt.Println("-------------apps2:---------------")
	for i := 0; i < len(apps2); i++ {
		fmt.Println(apps2[i])
	}

	var sortResult []Application = []Application{
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock: 352,
				Memory:   235,
				Storage:  325,
			},
			Priority: 8751,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock: 15,
				Memory:   15,
				Storage:  15,
			},
			Priority: 1000,
		},
		Application{
			IsTask: true,
			TaskReq: TaskResources{
				CPUCycle: 32,
				Memory:   25,
				Storage:  35,
			},
			Priority: 851,
		},
		Application{
			IsTask: true,
			TaskReq: TaskResources{
				CPUCycle: 832,
				Memory:   825,
				Storage:  835,
			},
			Priority: 810,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock: 1352,
				Memory:   1235,
				Storage:  1325,
			},
			Priority: 751,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock: 135,
				Memory:   135,
				Storage:  135,
			},
			Priority: 750,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock: 50,
				Memory:   50,
				Storage:  50,
			},
			Priority: 500,
		},
		Application{
			IsTask: true,
			TaskReq: TaskResources{
				CPUCycle: 83,
				Memory:   82,
				Storage:  83,
			},
			Priority: 111,
		},
	}

	assert.Equal(t, apps, sortResult, fmt.Sprintf("result is not expected"))
}

func TestDependencyValid(t *testing.T) {
	var apps []Application = []Application{
		Application{
			Depend: []Dependence{
				Dependence{
					AppIdx: 1,
				},
				Dependence{
					AppIdx: 2,
				},
				Dependence{
					AppIdx: 3,
				},
				Dependence{
					AppIdx: 4,
				},
			},
			Priority: 200,
		},
		Application{
			Depend: []Dependence{
				Dependence{
					AppIdx: 2,
				},
				Dependence{
					AppIdx: 3,
				},
			},
			Priority: 300,
		},
		Application{
			Depend: []Dependence{
				Dependence{
					AppIdx: 3,
				},
			},
			Priority: 400,
		},
		Application{
			Depend: []Dependence{
				Dependence{
					AppIdx: 4,
				},
			},
			Priority: 500,
		},
		Application{
			Depend:   []Dependence{},
			Priority: 600,
		},
	}
	if err := DependencyValid(apps); err != nil {
		t.Errorf("DependencyValid error: %s", err.Error())
	}
}

func TestCombApps(t *testing.T) {
	var aJson []string = []string{
		`[{"isTask":true,"svcReq":{"cpuClock":0,"memory":0,"storage":0},"taskReq":{"cpuCycle":60733631877.55724,"memory":1073741824,"storage":2147483648},"priority":235,"appIdx":0,"startTime":0,"taskCompletionTime":0,"depend":null,"inputDataSize":58309992.29971895,"imageSize":453858325.4848866},{"isTask":false,"svcReq":{"cpuClock":5.694191059356553,"memory":524288000,"storage":3221225472},"taskReq":{"cpuCycle":0,"memory":0,"storage":0},"priority":390,"appIdx":1,"startTime":0,"taskCompletionTime":0,"depend":null,"inputDataSize":19788809.42030173,"imageSize":334477193.4813267},{"isTask":false,"svcReq":{"cpuClock":5.064412080950095,"memory":1073741824,"storage":2147483648},"taskReq":{"cpuCycle":0,"memory":0,"storage":0},"priority":117,"appIdx":2,"startTime":0,"taskCompletionTime":0,"depend":[{"appIdx":1,"doneBw":204.8,"upBw":0.098,"rtt":2000},{"appIdx":0,"doneBw":2048,"upBw":100,"rtt":10}],"inputDataSize":17459215.258456357,"imageSize":384723363.9728706}]`, `[{"isTask":false,"svcReq":{"cpuClock":1.0612741707704028,"memory":2147483648,"storage":2147483648},"taskReq":{"cpuCycle":0,"memory":0,"storage":0},"priority":390,"appIdx":0,"startTime":0,"taskCompletionTime":0,"depend":null,"inputDataSize":56507951.69683887,"imageSize":62182150.912367344},{"isTask":true,"svcReq":{"cpuClock":0,"memory":0,"storage":0},"taskReq":{"cpuCycle":248597124482.8448,"memory":1073741824,"storage":3221225472},"priority":440,"appIdx":1,"startTime":0,"taskCompletionTime":0,"depend":null,"inputDataSize":63376622.084139556,"imageSize":367404730.4727229},{"isTask":false,"svcReq":{"cpuClock":3.0061695806384914,"memory":1073741824,"storage":2147483648},"taskReq":{"cpuCycle":0,"memory":0,"storage":0},"priority":324,"appIdx":2,"startTime":0,"taskCompletionTime":0,"depend":[{"appIdx":1,"doneBw":1,"upBw":0.098,"rtt":20},{"appIdx":0,"doneBw":100,"upBw":204.8,"rtt":10}],"inputDataSize":30162792.739897724,"imageSize":647022698.1995752}]`, `[{"isTask":false,"svcReq":{"cpuClock":3.2017874745679276,"memory":1073741824,"storage":8589934592},"taskReq":{"cpuCycle":0,"memory":0,"storage":0},"priority":121,"appIdx":0,"startTime":0,"taskCompletionTime":0,"depend":[{"appIdx":2,"doneBw":20,"upBw":1,"rtt":20},{"appIdx":1,"doneBw":10,"upBw":10,"rtt":10}],"inputDataSize":34141721.80977099,"imageSize":654122103.3672068},{"isTask":true,"svcReq":{"cpuClock":0,"memory":0,"storage":0},"taskReq":{"cpuCycle":44623332651.08046,"memory":1073741824,"storage":4294967296},"priority":540,"appIdx":1,"startTime":0,"taskCompletionTime":0,"depend":null,"inputDataSize":42135988.06440907,"imageSize":617486854.3532362},{"isTask":false,"svcReq":{"cpuClock":5.228081361161438,"memory":1073741824,"storage":4294967296},"taskReq":{"cpuCycle":0,"memory":0,"storage":0},"priority":988,"appIdx":2,"startTime":0,"taskCompletionTime":0,"depend":null,"inputDataSize":12029692.781259924,"imageSize":346593738.56870246}]`, `[{"isTask":false,"svcReq":{"cpuClock":6.655871733597715,"memory":524288000,"storage":3221225472},"taskReq":{"cpuCycle":0,"memory":0,"storage":0},"priority":831,"appIdx":0,"startTime":0,"taskCompletionTime":0,"depend":null,"inputDataSize":45431514.32192842,"imageSize":201736525.24258387},{"isTask":true,"svcReq":{"cpuClock":0,"memory":0,"storage":0},"taskReq":{"cpuCycle":138276854495.81833,"memory":1073741824,"storage":8589934592},"priority":100,"appIdx":1,"startTime":0,"taskCompletionTime":0,"depend":[{"appIdx":2,"doneBw":204.8,"upBw":0.0098,"rtt":2},{"appIdx":0,"doneBw":10,"upBw":100,"rtt":20}],"inputDataSize":49024955.43281666,"imageSize":67524937.22977632},{"isTask":false,"svcReq":{"cpuClock":6.049885452932674,"memory":1073741824,"storage":2147483648},"taskReq":{"cpuCycle":0,"memory":0,"storage":0},"priority":314,"appIdx":2,"startTime":0,"taskCompletionTime":0,"depend":[{"appIdx":0,"doneBw":204.8,"upBw":1,"rtt":10}],"inputDataSize":59849443.87086652,"imageSize":689220133.2831464}]`, `[{"isTask":false,"svcReq":{"cpuClock":6.884937382628491,"memory":1073741824,"storage":2147483648},"taskReq":{"cpuCycle":0,"memory":0,"storage":0},"priority":525,"appIdx":0,"startTime":0,"taskCompletionTime":0,"depend":null,"inputDataSize":6321017.201510479,"imageSize":853816730.2866552},{"isTask":false,"svcReq":{"cpuClock":2.9244045552803954,"memory":1073741824,"storage":4294967296},"taskReq":{"cpuCycle":0,"memory":0,"storage":0},"priority":392,"appIdx":1,"startTime":0,"taskCompletionTime":0,"depend":[{"appIdx":2,"doneBw":20,"upBw":2048,"rtt":2000}],"inputDataSize":50745744.523149654,"imageSize":527159009.6706801},{"isTask":true,"svcReq":{"cpuClock":0,"memory":0,"storage":0},"taskReq":{"cpuCycle":166259063054.4873,"memory":2147483648,"storage":2147483648},"priority":540,"appIdx":2,"startTime":0,"taskCompletionTime":0,"depend":null,"inputDataSize":59372635.21929153,"imageSize":377821271.134044},{"isTask":true,"svcReq":{"cpuClock":0,"memory":0,"storage":0},"taskReq":{"cpuCycle":22932337122.78057,"memory":2147483648,"storage":8589934592},"priority":232,"appIdx":3,"startTime":0,"taskCompletionTime":0,"depend":[{"appIdx":1,"doneBw":204.8,"upBw":2048,"rtt":1}],"inputDataSize":51139376.99009056,"imageSize":556018572.305882},{"isTask":false,"svcReq":{"cpuClock":8.131271503561994,"memory":1073741824,"storage":8589934592},"taskReq":{"cpuCycle":0,"memory":0,"storage":0},"priority":315,"appIdx":4,"startTime":0,"taskCompletionTime":0,"depend":[{"appIdx":0,"doneBw":100,"upBw":100,"rtt":1}],"inputDataSize":19668387.026345227,"imageSize":521724012.0901065}]`,
	}

	var a [][]Application
	var combApp []Application
	var n int = 5
	for i := 0; i < n; i++ {
		var theseApps []Application
		if err := json.Unmarshal([]byte(aJson[i]), &theseApps); err != nil {
			t.Fatalf("%d, unmarshal error: %s", i, err.Error())
		}

		a = append(a, theseApps)
		combApp = CombApps(combApp, a[i])
	}

	for i := 0; i < len(a); i++ {
		for j := 0; j < len(a[i]); j++ {
			fmt.Println(j, a[i][j].Depend)
			//fmt.Println(a[i][j])
		}
	}
	fmt.Println("---------------------")
	for i := 0; i < len(combApp); i++ {
		fmt.Println(i, combApp[i].Depend)
		//fmt.Println(combApp[i])

	}
}
