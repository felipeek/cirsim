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

func genGraphs(elementList *Element, X [][]float64) {
	e := elementList
	for e != nil {
		gv := graphValues{
			t: make([]float64, 0),
			v: make([]float64, 0),
		}
		for t, xx := range X {
			v1 := 0.0
			v2 := 0.0
			if e.Nodes[0] != 0 {
				v1 = xx[e.Nodes[0]-1]
			}
			if e.Nodes[1] != 0 {
				v2 = xx[e.Nodes[1]-1]
			}

			voltage := v1 - v2
			gv.t = append(gv.t, float64(t))
			gv.v = append(gv.v, voltage)
		}
		err := graphRender(e.Label, gv)
		if err != nil {
			panic(err)
		}
		e = e.Next
	}

}

func graphRender(label string, gv graphValues) error {
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
