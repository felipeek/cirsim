package internal

import (
	"bytes"
	"io/ioutil"

	chart "github.com/wcharczuk/go-chart"
)

type graphValues struct {
	t []float64
	v []float64
}

func genAllGraphs(currentNodes map[string]int, nodesMap map[string]int, X [][]float64) error {
	// Gen graph of all voltages
	for k, v := range nodesMap {
		if v != 0 {
			err := genGraph("voltage_"+k, X, v-1)
			if err != nil {
				return err
			}
		}
	}

	// Gen graph of all currents
	for k, v := range currentNodes {
		if v != 0 {
			err := genGraph("current_"+k, X, v-1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func genGraph(label string, X [][]float64, xIndex int) error {
	gv := graphValues{
		t: make([]float64, 0),
		v: make([]float64, 0),
	}
	for t, v := range X {
		gv.t = append(gv.t, float64(t))
		gv.v = append(gv.v, v[xIndex])
	}

	return graphRender(label, gv)
}

func graphRender(label string, gv graphValues) error {
	graph := chart.Chart{
		Width: 1920,
		XAxis: chart.XAxis{
			Name:      "t",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		YAxis: chart.YAxis{
			Name:      "value",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					Show: true,
				},
				XValues: gv.t,
				YValues: gv.v,
			},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("res/"+label+".png", buffer.Bytes(), 0644)
}
