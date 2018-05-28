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
	LU, P := mnaLUFactorization(H, B)
	Y := mnaProgressiveSubstitution(LU, B, P)
	Xp := mnaRegressiveSubstitution(LU, Y, P)

	X := make([]float64, len(Xp))

	for i := range X {
		X[i] = Xp[P[i]]
	}

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

func mnaRegressiveSubstitution(LU [][]float64, Y []float64, P []int) []float64 {
	X := make([]float64, len(Y))

	for k := len(LU[0]) - 1; k >= 0; k-- {
		X[P[k]] = Y[P[k]]

		for j := k + 1; j < len(LU[0]); j++ {
			X[P[k]] = X[P[k]] - LU[j][P[k]]*X[P[j]]
		}

		X[P[k]] = X[P[k]] / LU[k][P[k]]
	}

	return X
}

func mnaProgressiveSubstitution(LU [][]float64, B []float64, P []int) []float64 {
	Y := make([]float64, len(B))
	L := make([][]float64, len(LU))
	for i := range LU {
		L[i] = make([]float64, len(LU[i]))
		for j := range LU[i] {
			L[i][j] = LU[i][j]
		}
	}

	for i := range L {
		L[i][P[i]] = 1.0
	}

	for k := range L[0] {
		Y[P[k]] = B[P[k]]
		for j := 0; j < k; j++ {
			Y[P[k]] = Y[P[k]] - L[j][P[k]]*Y[P[j]]
		}

		Y[P[k]] = Y[P[k]] / L[k][P[k]]
	}

	return Y
}

func mnaLUFactorization(H [][]float64, B []float64) ([][]float64, []int) {
	P := make([]int, len(H[0])) // permutation vector
	for i := range P {
		P[i] = i
	}

	LU := make([][]float64, len(H))
	for i := range H {
		LU[i] = make([]float64, len(H[i]))
		for j := range H[i] {
			LU[i][j] = H[i][j]
		}
	}

	for k := range P {
		kMax := k

		for l := k + 1; l < len(LU[0]); l++ {
			if math.Abs(LU[k][P[l]]) > math.Abs(LU[k][P[k]]) {
				kMax = l
			}
		}

		aux := P[k]
		P[k] = P[kMax]
		P[kMax] = aux

		for i := k + 1; i < len(LU[0]); i++ {
			LU[k][P[i]] = LU[k][P[i]] / LU[k][P[k]]

			for j := k + 1; j < len(LU); j++ {
				LU[j][P[i]] = LU[j][P[i]] - LU[k][P[i]]*LU[j][P[k]]
			}
		}
	}
	return LU, P
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

	for i, v := range X {
		fmt.Printf("X(%d) = %f\n", i, v)
	}
}
