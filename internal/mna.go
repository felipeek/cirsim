package internal

import (
	"fmt"
	"math"
	"os"
)

func retrieveSourceValue(e Element, time float64) float64 {
	if e.ElementType != ElementCurrentSource && e.ElementType != ElementVoltageSource {
		panic("retrieveSourceValue must receive source element")
	}

	switch v := e.Extra.(type) {
	case sinDescriptor:
		c := 2.0*math.Pi*v.freq*time + v.td
		s := math.Sin(c)
		return v.v0 + v.va*s
	case []pwlDescriptor:
		timeBeforeIndex := 0
		for ; timeBeforeIndex < len(v); timeBeforeIndex++ {
			if v[timeBeforeIndex].t > time {
				break
			}
		}
		timeBeforeIndex--
		desc := v[timeBeforeIndex]
		if timeBeforeIndex == len(v)-1 {
			return desc.x
		} else {
			t0 := desc.t
			x0 := desc.x
			t1 := v[timeBeforeIndex+1].t
			x1 := v[timeBeforeIndex+1].x

			return x0 + (x1-x0)*((time-t0)/(t1-t0))
		}
	default:
		return e.Value
	}
}

func mnaSolveLinear(elementList *Element, nodesMap map[string]int) {
	// identify groups
	currentElement := elementList

	for currentElement != nil {
		if currentElement.ElementType == ElementVoltageSource || currentElement.ElementType == ElementVCVS ||
			currentElement.ElementType == ElementCCVS || currentElement.ElementType == ElementCapacitor ||
			currentElement.ElementType == ElementInductor {
			currentElement.PreserveCurrent = true
		}

		if currentElement.ElementType == ElementCCVS || currentElement.ElementType == ElementCCCS {
			e := elementListFindByLabel(elementList, currentElement.Extra.(string))
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

	mnaBuildStaticMatrices(elementList, currentNodes, H, B)
	mnaBuildDynamicMatrices(elementList, currentNodes, H, B, 0)

	LU, P := mnaLUFactorization(H, B)
	Y := mnaProgressiveSubstitution(LU, B, P)
	Xp := mnaRegressiveSubstitution(LU, Y, P)

	X := make([]float64, len(Xp))

	for i := range X {
		X[i] = Xp[P[i]]
	}

	mnaPrintMatrices(H, B, X, nodesMap, currentNodes)
}

func mnaBuildDynamicMatrices(elementList *Element, currentNodes map[string]int, H [][]float64, B []float64, t float64, X []float64, tStep float64) {
	e := elementList

	for e != nil {
		switch e.ElementType {
		case ElementBJT:
		case ElementDiode:
		case ElementMOSFET:
			fmt.Fprintf(os.Stderr, "MNA Error: Element not implemented.\n")
			os.Exit(1)
		case ElementCCCS, ElementCCVS, ElementResistor, ElementVCCS, ElementVCVS:
			// Treated as static
		case ElementCapacitor:
			if e.PreserveCurrent {
				capacitorVoltage := 0.0
				if t == 0 {
					capacitorVoltage = e.Extra.(float64)
				} else {
					lastVoltage := X[e.Nodes[0]-1] - X[e.Nodes[1]-1]
					capacitorVoltage = lastVoltage + (tStep/e.Value)*X[currentNodes[e.Label]-1]
				}

				if e.Nodes[0] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[0]-1][currentNodes[e.Label]-1] += 1.0
					H[currentNodes[e.Label]-1][e.Nodes[0]-1] += 1.0
				}
				if e.Nodes[1] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[1]-1][currentNodes[e.Label]-1] -= 1.0
					H[currentNodes[e.Label]-1][e.Nodes[1]-1] -= 1.0
				}
				if currentNodes[e.Label] != 0 {
					B[currentNodes[e.Label]-1] = capacitorVoltage
				}
			}
		case ElementInductor:
			if e.PreserveCurrent {
				inductorCurrent := 0.0
				if t == 0 {
					inductorCurrent = e.Extra.(float64)
				} else {
					lastCurrent := X[currentNodes[e.Label]-1]
					inductorCurrent = lastCurrent + (tStep/e.Value)*(X[e.Nodes[0]-1]-X[e.Nodes[1]-1])
				}

				if e.Nodes[0] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[0]-1][currentNodes[e.Label]-1] += 1.0
				}
				if e.Nodes[1] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[1]-1][currentNodes[e.Label]-1] -= 1.0
				}
				if currentNodes[e.Label] != 0 {
					H[currentNodes[e.Label]-1][currentNodes[e.Label]-1] += 1.0
					B[currentNodes[e.Label]-1] = inductorCurrent
				}
			}
		case ElementCurrentSource:
			cValue := retrieveSourceValue(*e, t)
			if !e.PreserveCurrent {
				if e.Nodes[0] != 0 {
					B[e.Nodes[0]-1] -= cValue
				}
				if e.Nodes[1] != 0 {
					B[e.Nodes[1]-1] += cValue
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
					B[currentNodes[e.Label]-1] += cValue
				}
			}

		case ElementVoltageSource:
			vValue := retrieveSourceValue(*e, t)
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
					B[currentNodes[e.Label]-1] += vValue
				}
			}
		}

		e = e.Next
	}
}

func mnaBuildStaticMatrices(elementList *Element, currentNodes map[string]int, H [][]float64, B []float64) {
	e := elementList

	for e != nil {
		switch e.ElementType {
		case ElementBJT, ElementDiode, ElementMOSFET:
			fmt.Fprintf(os.Stderr, "MNA Error: Element not implemented.\n")
			os.Exit(1)
		case ElementCapacitor, ElementInductor, ElementCurrentSource, ElementVoltageSource:
			// Treated as dynamic
		case ElementCCCS:
			if !e.PreserveCurrent {
				controlElement := elementListFindByLabel(elementList, e.Extra.(string))
				if e.Nodes[0] != 0 && currentNodes[controlElement.Label] != 0 {
					H[e.Nodes[0]-1][currentNodes[controlElement.Label]-1] += e.Value
				}
				if e.Nodes[1] != 0 && currentNodes[controlElement.Label] != 0 {
					H[e.Nodes[1]-1][currentNodes[controlElement.Label]-1] -= e.Value
				}
			} else {
				controlElement := elementListFindByLabel(elementList, e.Extra.(string))
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
				controlElement := elementListFindByLabel(elementList, e.Extra.(string))
				if e.Nodes[0] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[0]-1][currentNodes[e.Label]-1] += 1.0
					H[currentNodes[e.Label]-1][e.Nodes[0]-1] += 1.0
				}
				if e.Nodes[1] != 0 && currentNodes[e.Label] != 0 {
					H[e.Nodes[1]-1][currentNodes[e.Label]-1] -= 1.0
					H[currentNodes[e.Label]-1][e.Nodes[1]-1] -= 1.0
				}
				if currentNodes[e.Label] != 0 && currentNodes[controlElement.Label] != 0 {
					H[currentNodes[e.Label]-1][currentNodes[controlElement.Label]-1] -= e.Value
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
				if e.Nodes[0] != 0 && currentNodes[e.Label]-1 != 0 {
					H[e.Nodes[0]-1][currentNodes[e.Label]-1] += 1.0
				}
				if e.Nodes[1] != 0 && currentNodes[e.Label]-1 != 0 {
					H[e.Nodes[1]-1][currentNodes[e.Label]-1] -= 1.0
				}
				if currentNodes[e.Label]-1 != 0 && e.Nodes[2] != 0 {
					H[currentNodes[e.Label]-1][e.Nodes[2]] -= e.Value
				}
				if currentNodes[e.Label]-1 != 0 && e.Nodes[3] != 0 {
					H[currentNodes[e.Label]-1][e.Nodes[3]] += e.Value
				}
				if currentNodes[e.Label]-1 != 0 {
					H[currentNodes[e.Label]-1][currentNodes[e.Label]-1] += 1.0
				}
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
		}

		e = e.Next
	}
}

func mnaRegressiveSubstitution(LU [][]float64, Y []float64, P []int) []float64 {
	X := make([]float64, len(Y))

	for k := len(LU[0]) - 1; k >= 0; k-- {
		X[P[k]] = Y[P[k]]

		for j := k + 1; j < len(LU[0]); j++ {
			X[P[k]] = X[P[k]] - LU[P[k]][j]*X[P[j]]
		}

		X[P[k]] = X[P[k]] / LU[P[k]][k]
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
		L[P[i]][i] = 1.0
	}

	for k := range L[0] {
		Y[P[k]] = B[P[k]]
		for j := 0; j < k; j++ {
			Y[P[k]] = Y[P[k]] - L[P[k]][j]*Y[P[j]]
		}

		Y[P[k]] = Y[P[k]] / L[P[k]][k]
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
			if math.Abs(LU[P[l]][k]) > math.Abs(LU[P[k]][k]) {
				kMax = l
			}
		}

		aux := P[k]
		P[k] = P[kMax]
		P[kMax] = aux

		for i := k + 1; i < len(LU[0]); i++ {
			LU[P[i]][k] = LU[P[i]][k] / LU[P[k]][k]

			for j := k + 1; j < len(LU); j++ {
				LU[P[i]][j] = LU[P[i]][j] - LU[P[i]][k]*LU[P[k]][j]
			}
		}
	}
	return LU, P
}

func mnaPrintMatrices(H [][]float64, B []float64, X []float64, nodesMap map[string]int, currentNodes map[string]int) {
	fmt.Printf("Matrix H (Row-Major):\n\n")
	for i, l := range H {
		for j, v := range l {
			fmt.Printf("\tH(%d,%d) = %.3f\n", i, j, v)
		}
	}

	fmt.Printf("\nVector s:\n\n")
	for i, v := range B {
		fmt.Printf("\ts(%d) = %.3f\n", i, v)
	}

	fmt.Printf("\nResults:\n\n")
	fmt.Printf("\tVoltages:\n")
	for k, v := range nodesMap {
		if v != 0 {
			fmt.Printf("\tV(%s) = %.3f V\n", k, X[v-1])
		}
	}
	fmt.Printf("\n\tCurrents:\n")
	for k, v := range currentNodes {
		if v != 0 {
			fmt.Printf("\tI(%s) = %.3f A\n", k, X[v-1])
		}
	}
}

func mnaSolveDynamic(elementList *Element, nodesMap map[string]int, tStep float64, tStop float64) {
	elementListPrint(elementList)

	e := elementList
	for e != nil {
		if e.ElementType == ElementVoltageSource || e.ElementType == ElementCurrentSource {
			for t := 0.0; t <= tStop; t += tStep {
				v := retrieveSourceValue(*e, t)
				graphCollect(*e, t, v)
			}
			err := graphRender(*e)
			if err != nil {
				panic(err)
			}
		}
		e = e.Next
	}
}
