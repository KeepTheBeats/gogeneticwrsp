package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/wcharczuk/go-chart"

	"gogeneticwrsp/algorithms"
	"gogeneticwrsp/experiments"
	"gogeneticwrsp/model"
)

func main() {
	// set the log to show line number and file name
	log.SetFlags(0 | log.Lshortfile)

	log.Println("Hello World!")

	var numCloud, numApp int = 10, 80

	// generate clouds and apps, and write to files
	experiments.GenerateCloudsApps(numCloud, numApp)

	// read clouds and apps from files
	var clouds []model.Cloud
	var apps []model.Application
	clouds, apps = experiments.ReadCloudsApps(numCloud, numApp)

	for i := 0; i < numCloud; i++ {
		log.Println(clouds[i])
	}

	for i := 0; i < numApp; i++ {
		log.Println(apps[i])
	}

	geneticAlgorithm := algorithms.NewGenetic(100, 500, 0.7, 0.007, clouds, apps)

	solution, err := geneticAlgorithm.Schedule(clouds, apps)
	if err != nil {
		log.Printf("geneticAlgorithm.Schedule(clouds, apps), error: %s", err.Error())
	}
	log.Println("solution:", solution)
	for i := 0; i < len(geneticAlgorithm.FitnessRecordIterationBest); i++ {
		log.Printf("Iteration %d: FitnessRecordIterationBest: %f\n", i, geneticAlgorithm.FitnessRecordIterationBest[i])
	}
	log.Println()
	if len(geneticAlgorithm.FitnessRecordBestUntilNow) != len(geneticAlgorithm.BestUntilNowUpdateIterations) {
		log.Panicf("len(geneticAlgorithm.FitnessRecordBestUntilNow): %d, len(geneticAlgorithm.BestUntilNowUpdateIterations): %d\n", len(geneticAlgorithm.FitnessRecordBestUntilNow), len(geneticAlgorithm.BestUntilNowUpdateIterations))
	}
	for i := 0; i < len(geneticAlgorithm.FitnessRecordBestUntilNow); i++ {
		log.Printf("Iteration %d: FitnessRecordBestUntilNow: %f\n", int(geneticAlgorithm.BestUntilNowUpdateIterations[i]), geneticAlgorithm.FitnessRecordBestUntilNow[i])
	}

	// draw geneticAlgorithm.FitnessRecordIterationBest and geneticAlgorithm.FitnessRecordBestUntilNow on a line chart
	var drawChartFunc func(http.ResponseWriter, *http.Request) = func(res http.ResponseWriter, r *http.Request) {
		var xValuesIterationBest []float64
		for i, _ := range geneticAlgorithm.FitnessRecordIterationBest {
			xValuesIterationBest = append(xValuesIterationBest, float64(i))
		}

		graph := chart.Chart{
			Title: "Evolution",
			XAxis: chart.XAxis{
				Name:      "Iteration Number",
				NameStyle: chart.StyleShow(),
				Style:     chart.StyleShow(),
				ValueFormatter: func(v interface{}) string {
					return strconv.FormatInt(int64(v.(float64)), 10)
				},
			},
			YAxis: chart.YAxis{
				AxisType:  chart.YAxisSecondary,
				Name:      "Fitness",
				NameStyle: chart.StyleShow(),
				Style:     chart.StyleShow(),
			},
			Background: chart.Style{
				Padding: chart.Box{
					Top:  50,
					Left: 20,
				},
			},
			Series: []chart.Series{
				chart.ContinuousSeries{
					Name:    "Best Fitness in each iteration",
					XValues: xValuesIterationBest,
					YValues: geneticAlgorithm.FitnessRecordIterationBest,
				},
				chart.ContinuousSeries{
					Name: "Best Fitness in all iterations",
					// the first value (iteration -1) is much different with others, which will cause that we cannot observer the trend of the evolution
					XValues: geneticAlgorithm.BestUntilNowUpdateIterations[1:],
					YValues: geneticAlgorithm.FitnessRecordBestUntilNow[1:],
				},
			},
		}

		graph.Elements = []chart.Renderable{
			chart.LegendThin(&graph),
		}

		res.Header().Set("Content-Type", "image/png")
		err := graph.Render(chart.PNG, res)
		if err != nil {
			log.Println("Error: graph.Render(chart.PNG, res)", err)
		}
	}

	http.HandleFunc("/", drawChartFunc)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("Error: http.ListenAndServe(\":8080\", nil)", err)
	}
}
