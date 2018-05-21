package internal

import (
	"fmt"
	"os"
)

func mnaSolveLinear(elementList *Element, nodesMap map[string]int) {
	elementListPrint(elementList)

	// identify groups
	currentElement := elementList

	for currentElement != nil {
		if currentElement.ElementType == ElementVoltageSource || currentElement.ElementType == ElementVCVS ||
			currentElement.ElementType == ElementCCVS {
			currentElement.PreserveCurrent = true
		}

		if currentElement.ElementType == ElementCCVS || currentElement.ElementType == ElementCCCS {
			e := elementListFindByLabel(elementList, currentElement.Extra)
			if e == nil {
				fmt.Fprintf(os.Stderr, "MNA Error: VCCS or CCCS has Control Element with invalid label\n")
				os.Exit(1)
			}
			e.PreserveCurrent = true
		}

		currentElement = currentElement.Next
	}

	// assign indices to current nodes
	currentNodes := make(map[string]int, 1)
	startingIndex := len(nodesMap)

	currentElement = elementList
	for currentElement != nil {
		if currentElement.PreserveCurrent {
			currentNodes[currentElement.Label] = startingIndex
			startingIndex = startingIndex + 1
		}

		currentElement = currentElement.Next
	}

	fmt.Println("NodesMap:")
	for k, v := range nodesMap {
		fmt.Printf("%+v: %+v\n", k, v)
	}

	fmt.Println("CurrentMap:")
	for k, v := range currentNodes {
		fmt.Printf("%+v: %+v\n", k, v)
	}

	// Create H Matrix
	H := make([][]float64, len(nodesMap)+len(currentNodes)-1)
	for i, _ := range H {
		H[i] = make([]float64, len(nodesMap)+len(currentNodes)-1)
	}

	// Create B Array
	B := make([]float64, len(nodesMap)+len(currentNodes)-1)

	mnaBuildMatrices(elementList, currentNodes, H, B)

	fmt.Printf("H:\n\n%+v\n\nB:\n\n%+v\n\n", H, B)
}

func mnaBuildMatrices(elementList *Element, currentNodes map[string]int, H [][]float64, B []float64) {
	e := elementList

	for e != nil {
		switch e.ElementType {
		case ElementBJT:
		case ElementCapacitor:
		case ElementDiode:
		case ElementInductor:
		case ElementMOSFET:
			fmt.Fprintf(os.Stderr, "MNA Error: Element not implemented.\n")
			os.Exit(1)
		case ElementCCCS:
			if !e.PreserveCurrent {
				controlElement := elementListFindByLabel(elementList, e.Extra)
				if e.Nodes[0] != 0 && currentNodes[controlElement.Label] != 0 {
					H[e.Nodes[0]-1][currentNodes[controlElement.Label]-1] += e.Value
				}
				if e.Nodes[1] != 0 && currentNodes[controlElement.Label] != 0 {
					H[e.Nodes[1]-1][currentNodes[controlElement.Label]-1] -= e.Value
				}
			} else {
				controlElement := elementListFindByLabel(elementList, e.Extra)
				if e.Nodes[0] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[0]-1][currentNodes[e.Label]-1] += 1.0
				}
				if e.Nodes[1] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[1]-1][currentNodes[e.Label]-1] -= 1.0
				}
				if currentNodes[e.Label] != 0 {
					H[currentNodes[e.Label]-1][currentNodes[e.Label]-1] += 1.0
				}
				if currentNodes[e.Label] != 0 && currentNodes[controlElement.Label] != 0 {
					H[currentNodes[e.Label]-1][currentNodes[controlElement.Label]-1] -= e.Value
				}
			}
		case ElementCCVS:
			if e.PreserveCurrent {
				controlElement := elementListFindByLabel(elementList, e.Extra)
				if e.Nodes[0] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[0]-1][currentNodes[e.Label]-1] += 1.0
				}
				if e.Nodes[1] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[1]-1][currentNodes[e.Label]-1] -= 1.0
				}
				if currentNodes[controlElement.Label] != 0 && e.Nodes[0] != 0 {
					H[currentNodes[controlElement.Label]-1][e.Nodes[0]-1] += 1.0
				}
				if currentNodes[controlElement.Label] != 0 && e.Nodes[1] != 0 {
					H[currentNodes[controlElement.Label]-1][e.Nodes[1]-1] -= 1.0
				}
				if currentNodes[controlElement.Label] != 0 {
					H[currentNodes[controlElement.Label]-1][currentNodes[controlElement.Label]-1] -= e.Value
				}
			}
		case ElementCurrentSource:
			if !e.PreserveCurrent {
				if e.Nodes[0] != 0 {
					B[e.Nodes[0]-1] -= e.Value
				}
				if e.Nodes[1] != 0 {
					B[e.Nodes[1]-1] += e.Value
				}
			} else {
				if e.Nodes[0] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[0]-1][currentNodes[e.Label]-1] += 1.0
				}
				if e.Nodes[1] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[1]-1][currentNodes[e.Label]-1] -= 1.0
				}
				if currentNodes[e.Label] != 0 {
					H[currentNodes[e.Label]-1][currentNodes[e.Label]-1] += 1.0
					B[currentNodes[e.Label]-1] += e.Value
				}
			}
		case ElementResistor:
			if !e.PreserveCurrent {
				if e.Nodes[0] != 0 {
					H[e.Nodes[0]-1][e.Nodes[0]-1] += 1.0 / e.Value
				}
				if e.Nodes[0] != 0 && e.Nodes[1] != 0 {
					H[e.Nodes[0]-1][e.Nodes[1]-1] -= 1.0 / e.Value
					H[e.Nodes[1]-1][e.Nodes[0]-1] -= 1.0 / e.Value
				}
				if e.Nodes[1] != 0 {
					H[e.Nodes[1]-1][e.Nodes[1]-1] += 1.0 / e.Value
				}
			} else {
				if e.Nodes[0] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[0]-1][currentNodes[e.Label]-1] += 1.0
					H[currentNodes[e.Label]-1][e.Nodes[0]-1] += 1.0
				}
				if e.Nodes[1] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[1]-1][currentNodes[e.Label]-1] -= 1.0
					H[currentNodes[e.Label]-1][e.Nodes[1]-1] -= 1.0
				}
				if currentNodes[e.Label] != 0 {
					H[currentNodes[e.Label]-1][currentNodes[e.Label]-1] -= e.Value
				}
			}
		case ElementVCCS:
			if !e.PreserveCurrent {
				if e.Nodes[0] != 0 && e.Nodes[2] != 0 {
					H[e.Nodes[0]-1][e.Nodes[2]-1] += e.Value
				}
				if e.Nodes[0] != 0 && e.Nodes[3] != 0 {
					H[e.Nodes[0]-1][e.Nodes[3]-1] -= e.Value
				}
				if e.Nodes[1] != 0 && e.Nodes[2] != 0 {
					H[e.Nodes[1]-1][e.Nodes[2]-1] -= e.Value
				}
				if e.Nodes[1] != 0 && e.Nodes[3] != 0 {
					H[e.Nodes[1]-1][e.Nodes[3]-1] += e.Value
				}
			} else {
				fmt.Fprintf(os.Stderr, "VCCS G2 not implemented yet.")
				os.Exit(1)
			}
		case ElementVCVS:
			if e.PreserveCurrent {
				if e.Nodes[0] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[0]-1][currentNodes[e.Label]-1] += 1.0
					H[currentNodes[e.Label]-1][e.Nodes[0]-1] += 1.0
				}
				if e.Nodes[1] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[1]-1][currentNodes[e.Label]-1] -= 1.0
					H[currentNodes[e.Label]-1][e.Nodes[1]-1] -= 1.0
				}
				if currentNodes[e.Label] != 0 && e.Nodes[2] != 0 {
					H[currentNodes[e.Label]-1][e.Nodes[2]-1] -= e.Value
				}
				if currentNodes[e.Label] != 0 && e.Nodes[3] != 0 {
					H[currentNodes[e.Label]-1][e.Nodes[3]-1] += e.Value
				}
			}
		case ElementVoltageSource:
			if e.PreserveCurrent {
				if e.Nodes[0] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[0]-1][currentNodes[e.Label]-1] += 1.0
					H[currentNodes[e.Label]-1][e.Nodes[0]-1] += 1.0
				}
				if e.Nodes[1] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[1]-1][currentNodes[e.Label]-1] -= 1.0
					H[currentNodes[e.Label]-1][e.Nodes[1]-1] -= 1.0
				}
				if currentNodes[e.Label] != 0 {
					B[currentNodes[e.Label]-1] += e.Value
				}
			}
		}

		e = e.Next
	}
}
