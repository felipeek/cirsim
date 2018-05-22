package internal

import (
	"fmt"
	"math"
	"os"
)

func mnaSolveLinear(elementList *Element, nodesMap map[string]int) {
	//elementListPrint(elementList)

	//H1 := [][]float64{{1, 0, 2}, {2, 1, 0}, {2, 1, 1}}
	//B1 := []float64{1, 2, 1}

	//mnaLUFactorization(H1, B1)

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

	// Create H Matrix
	H := make([][]float64, len(nodesMap)+len(currentNodes)-1)
	for i, _ := range H {
		H[i] = make([]float64, len(nodesMap)+len(currentNodes)-1)
	}

	// Create B Array
	B := make([]float64, len(nodesMap)+len(currentNodes)-1)

	mnaBuildMatrices(elementList, currentNodes, H, B)
	X := mnaLUFactorization(H, B)

	//fmt.Println("X: %+v\n", X)
	mnaPrintMatrices(H, B, X)
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

func mnaLUFactorization(H [][]float64, B []float64) []float64 {
	LU := make([][]float64, len(H))

	for i := 0; i < len(H); i++ {
		LU[i] = make([]float64, len(H))

		for j := 0; j < len(H); j++ {
			LU[i][j] = H[i][j]
		}
	}
	for k := 0; k < len(H); k++ {

		// partial pivoting
		absValue := math.Abs(LU[k][k])
		absIndex := k
		for i := k + 1; i < len(H); i++ {
			if math.Abs(LU[k][i]) > absValue {
				absIndex = i
				absValue = math.Abs(LU[k][i])
			}
		}

		//switch rows
		tmp := LU[k]
		LU[k] = LU[absIndex]
		LU[absIndex] = tmp

		tmp2 := B[k]
		B[k] = B[absIndex]
		B[absIndex] = tmp2

		for i := k + 1; i < len(H); i++ {
			LU[i][k] = H[i][k] / H[k][k]
			for j := k + 1; j < len(H); j++ {
				LU[i][j] = H[i][j] - H[i][k]*H[k][j]
			}
		}
	}

	//printMatrix(LU)

	Y := make([]float64, len(B))
	for k := 0; k < len(H); k++ {
		Y[k] = B[k]
		for j := 0; j < k-1; j++ {
			Y[k] = Y[k] - LU[k][j]*Y[j]
		}
		Y[k] = Y[k] / LU[k][k]
	}

	X := make([]float64, len(B))
	for k := len(H) - 1; k >= 0; k-- {
		X[k] = B[k]
		for j := k + 1; j < len(H); j++ {
			X[k] = X[k] - LU[k][j]*X[j]
		}
		X[k] = X[k] / LU[k][k]
	}

	return X
}

func mnaPrintMatrices(H [][]float64, B []float64, X []float64) {
	for i, l := range H {
		for j, v := range l {
			fmt.Printf("H(%d,%d) = %f\n", i, j, v)
		}
	}

	for i, v := range B {
		fmt.Printf("B(%d) = %f\n", i, v)
	}

	//for i, v := range X {
	//	fmt.Printf("X(%d) = %f\n", i, v)
	//}
}

func printMatrix(matrix [][]float64) {
	for _, l := range matrix {
		fmt.Printf("[")
		for j, k := range l {
			fmt.Printf("%.3f", k)
			if j != len(l)-1 {
				fmt.Printf(",")
			}
		}
		fmt.Printf("]\n")
	}
}
