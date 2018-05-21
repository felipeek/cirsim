package internal

import "fmt"

type ElementType int

const (
	ElementResistor      ElementType = 0
	ElementCapacitor     ElementType = 1
	ElementInductor      ElementType = 2
	ElementVoltageSource ElementType = 3
	ElementCurrentSource ElementType = 4
	ElementVCVS          ElementType = 5
	ElementCCCS          ElementType = 6
	ElementVCCS          ElementType = 7
	ElementCCVS          ElementType = 8
	ElementDiode         ElementType = 9
	ElementBJT           ElementType = 10
	ElementMOSFET        ElementType = 11
)

type Element struct {
	ElementType     ElementType
	Label           string
	Nodes           []int
	Value           float64
	Extra           string // model or control element (CCCS CCVS)
	PreserveCurrent bool   // used by MNA algorithm
	Next            *Element
}

func elementListAppend(elementList *Element, e *Element) {
	tmp := elementList

	for tmp.Next != nil {
		tmp = tmp.Next
	}

	tmp.Next = e
}

func elementListFindByLabel(elementList *Element, label string) *Element {
	tmp := elementList

	for tmp != nil {
		if tmp.Label == label {
			return tmp
		}
		tmp = tmp.Next
	}

	return nil
}

func elementListPrint(elementList *Element) {
	e := elementList
	count := 1

	for e != nil {
		fmt.Printf("Element %d: \n", count)
		elementPrint(e)
		count = count + 1
		e = e.Next
	}
}

func elementPrint(e *Element) {
	switch e.ElementType {
	case ElementBJT:
		fmt.Printf("\tType: BJT\n")
	case ElementCapacitor:
		fmt.Printf("\tType: Capacitor\n")
	case ElementCCCS:
		fmt.Printf("\tType: CCCS\n")
	case ElementCCVS:
		fmt.Printf("\tType: CCVS\n")
	case ElementCurrentSource:
		fmt.Printf("\tType: Current Source\n")
	case ElementDiode:
		fmt.Printf("\tType: Diode\n")
	case ElementInductor:
		fmt.Printf("\tType: Inductor\n")
	case ElementMOSFET:
		fmt.Printf("\tType: MOSFET\n")
	case ElementResistor:
		fmt.Printf("\tType: Resistor\n")
	case ElementVCCS:
		fmt.Printf("\tType: VCCS\n")
	case ElementVCVS:
		fmt.Printf("\tType: VCVS\n")
	case ElementVoltageSource:
		fmt.Printf("\tType: Voltage Source\n")
	}

	fmt.Printf("\tLabel: %s\n", e.Label)

	fmt.Printf("\tNodes:\n")
	for i, n := range e.Nodes {
		fmt.Printf("\t\tNode %d: [%d]\n", i, n)
	}

	if e.ElementType == ElementCCCS || e.ElementType == ElementCCVS {
		fmt.Printf("\tControl Element: %s\n", e.Extra)
	}

	if e.ElementType == ElementBJT || e.ElementType == ElementMOSFET {
		fmt.Printf("\tModel: %s\n", e.Extra)
	} else {
		fmt.Printf("\tValue: %f\n", e.Value)
	}
}
