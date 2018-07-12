package internal

import (
	"bytes"
	"fmt"
	"io/ioutil"

	chart "github.com/wcharczuk/go-chart"
)

var (
	graphs map[string]*graphValues = make(map[string]*graphValues)
)

type graphValues struct {
	t []float64
	v []float64
}

func graphCollect(e Element, t float64, v float64) {
	graph, ok := graphs[e.Label]

	if !ok {
		graphs[e.Label] = &graphValues{
			t: make([]float64, 0),
			v: make([]float64, 0),
		}
		graph = graphs[e.Label]
	}

	graph.t = append(graph.t, t)
	graph.v = append(graph.v, v)
}

func graphRender(e Element) error {
	v, ok := graphs[e.Label]
	if !ok {
		return fmt.Errorf("trying to render graph that does not exist")
	}
	graph := chart.Chart{
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
				XValues: v.t,
				YValues: v.v,
			},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("res/"+e.Label+".png", buffer.Bytes(), 0644)
}
