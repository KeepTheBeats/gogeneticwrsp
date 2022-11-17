package model

import (
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
				CPUClock:   50,
				Memory:     50,
				Storage:    50,
				NetLatency: 50,
			},
			Priority: 500,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock:   15,
				Memory:     15,
				Storage:    15,
				NetLatency: 15,
			},
			Priority: 1000,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock:   135,
				Memory:     135,
				Storage:    135,
				NetLatency: 153,
			},
			Priority: 750,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock:   1352,
				Memory:     1235,
				Storage:    1325,
				NetLatency: 1532,
			},
			Priority: 751,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock:   352,
				Memory:     235,
				Storage:    325,
				NetLatency: 532,
			},
			Priority: 8751,
		},
		Application{
			IsTask: true,
			TaskReq: TaskResources{
				CPUCycle:   32,
				Memory:     25,
				Storage:    35,
				NetLatency: 52,
			},
			Priority: 851,
		},
		Application{
			IsTask: true,
			TaskReq: TaskResources{
				CPUCycle:   832,
				Memory:     825,
				Storage:    835,
				NetLatency: 852,
			},
			Priority: 810,
		},
		Application{
			IsTask: true,
			TaskReq: TaskResources{
				CPUCycle:   83,
				Memory:     82,
				Storage:    83,
				NetLatency: 85,
			},
			Priority: 111,
		},
	}

	var apps2 []Application = make([]Application, len(apps))
	copy(apps2, apps)

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
				CPUClock:   352,
				Memory:     235,
				Storage:    325,
				NetLatency: 532,
			},
			Priority: 8751,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock:   15,
				Memory:     15,
				Storage:    15,
				NetLatency: 15,
			},
			Priority: 1000,
		},
		Application{
			IsTask: true,
			TaskReq: TaskResources{
				CPUCycle:   32,
				Memory:     25,
				Storage:    35,
				NetLatency: 52,
			},
			Priority: 851,
		},
		Application{
			IsTask: true,
			TaskReq: TaskResources{
				CPUCycle:   832,
				Memory:     825,
				Storage:    835,
				NetLatency: 852,
			},
			Priority: 810,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock:   1352,
				Memory:     1235,
				Storage:    1325,
				NetLatency: 1532,
			},
			Priority: 751,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock:   135,
				Memory:     135,
				Storage:    135,
				NetLatency: 153,
			},
			Priority: 750,
		},
		Application{
			IsTask: false,
			SvcReq: ServiceResources{
				CPUClock:   50,
				Memory:     50,
				Storage:    50,
				NetLatency: 50,
			},
			Priority: 500,
		},
		Application{
			IsTask: true,
			TaskReq: TaskResources{
				CPUCycle:   83,
				Memory:     82,
				Storage:    83,
				NetLatency: 85,
			},
			Priority: 111,
		},
	}

	assert.Equal(t, apps, sortResult, fmt.Sprintf("result is not expected"))
}

func TestHaveCirDep(t *testing.T) {
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
