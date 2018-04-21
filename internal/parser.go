package internal

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

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
	ElementType ElementType
	Label       string
	Nodes       []int
	Value       float64
	Model       string
	next        *Element
}

func ParserInit(netListPath string) {
	var token Token
	lexer := LexerInit(netListPath)
	eof := false
	nodesMap := make(map[string]int)
	nodesQuantity := 1
	nodesMap["0"] = 0
	var elementList *Element = nil

	for !lexer.eof {
		token = LexerNextToken(&lexer)
		switch token.TokenType {
		case TokenLineBreak:
			{

			}
		case TokenCommand:
			{
				// This is not implemented yet.
				// Ignore until a line break is received
				for !eof && token.TokenType != TokenLineBreak {
					token = LexerNextToken(&lexer)
				}
			}
		case TokenStr:
			{
				// Parse "Element" Line
				var (
					err bool
					e   *Element
				)

				err, e = parserParseElement(&lexer, token.TokenValue, nodesMap, &nodesQuantity)
				if err {
					return
				} else {
					if elementList != nil {
						parserAppendElement(elementList, e)
					} else {
						elementList = e
					}
				}
			}
		}
	}

	parserPrintElementList(elementList)
	fmt.Println()
	parserPrintNodesMap(nodesMap)
}

func parserParseElement(lexer *Lexer, elementArray string,
	nodesMap map[string]int, nodesQuantity *int) (bool, *Element) {
	var e = new(Element)
	var err bool
	var nodeToken Token
	currentLine := lexer.lineNumber

	switch elementArray[0] {
	case 'r':
		e.ElementType = ElementResistor
	case 'c':
		e.ElementType = ElementCapacitor
	case 'l':
		e.ElementType = ElementInductor
	case 'v':
		e.ElementType = ElementVoltageSource
	case 'i':
		e.ElementType = ElementCurrentSource
	case 'e':
		e.ElementType = ElementVCVS
	case 'f':
		e.ElementType = ElementCCCS
	case 'g':
		e.ElementType = ElementVCCS
	case 'h':
		e.ElementType = ElementCCVS
	case 'd':
		e.ElementType = ElementDiode
	case 'q':
		e.ElementType = ElementBJT
	case 'm':
		e.ElementType = ElementMOSFET
	}

	e.Label = elementArray[1:]
	e.next = nil

	if e.ElementType == ElementResistor || e.ElementType == ElementCapacitor ||
		e.ElementType == ElementInductor || e.ElementType == ElementVoltageSource ||
		e.ElementType == ElementCurrentSource || e.ElementType == ElementDiode {
		e.Nodes = make([]int, 2)

		// Get First Node
		nodeToken = LexerNextToken(lexer)
		if lexer.eof || nodeToken.TokenType != TokenStr {
			fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
			return true, e
		}

		nodeName := nodeToken.TokenValue
		nodeNumber, exists := nodesMap[nodeName]

		if exists {
			e.Nodes[0] = nodeNumber
		} else {
			e.Nodes[0] = *nodesQuantity
			nodesMap[nodeName] = *nodesQuantity
			*nodesQuantity = *nodesQuantity + 1
		}

		// Get Second Node
		nodeToken = LexerNextToken(lexer)
		if lexer.eof || nodeToken.TokenType != TokenStr {
			fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
			return true, e
		}

		nodeName = nodeToken.TokenValue
		nodeNumber, exists = nodesMap[nodeName]

		if exists {
			e.Nodes[1] = nodeNumber
		} else {
			e.Nodes[1] = *nodesQuantity
			nodesMap[nodeName] = *nodesQuantity
			*nodesQuantity = *nodesQuantity + 1
		}

		// Get Value
		nodeToken = LexerNextToken(lexer)
		err, e.Value = parserParseNumber(nodeToken.TokenValue)

		if err {
			fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
			return true, e
		}
	}

	if e.ElementType == ElementVCCS || e.ElementType == ElementVCVS ||
		e.ElementType == ElementCCVS || e.ElementType == ElementCCCS {
		e.Nodes = make([]int, 4)

		// Get First Node
		nodeToken := LexerNextToken(lexer)
		if lexer.eof || nodeToken.TokenType != TokenStr {
			fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
			return true, e
		}

		nodeName := nodeToken.TokenValue
		nodeNumber, exists := nodesMap[nodeName]

		if exists {
			e.Nodes[0] = nodeNumber
		} else {
			e.Nodes[0] = *nodesQuantity
			nodesMap[nodeName] = *nodesQuantity
			*nodesQuantity = *nodesQuantity + 1
		}

		// Get Second Node
		nodeToken = LexerNextToken(lexer)
		if lexer.eof || nodeToken.TokenType != TokenStr {
			fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
			return true, e
		}

		nodeName = nodeToken.TokenValue
		nodeNumber, exists = nodesMap[nodeName]

		if exists {
			e.Nodes[1] = nodeNumber
		} else {
			e.Nodes[1] = *nodesQuantity
			nodesMap[nodeName] = *nodesQuantity
			*nodesQuantity = *nodesQuantity + 1
		}

		// Get Third Node
		nodeToken = LexerNextToken(lexer)
		if lexer.eof || nodeToken.TokenType != TokenStr {
			fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
			return true, e
		}

		nodeName = nodeToken.TokenValue
		nodeNumber, exists = nodesMap[nodeName]

		if exists {
			e.Nodes[2] = nodeNumber
		} else {
			e.Nodes[2] = *nodesQuantity
			nodesMap[nodeName] = *nodesQuantity
			*nodesQuantity = *nodesQuantity + 1
		}

		// Get Fourth Node
		nodeToken = LexerNextToken(lexer)
		if lexer.eof || nodeToken.TokenType != TokenStr {
			fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
			return true, e
		}

		nodeName = nodeToken.TokenValue
		nodeNumber, exists = nodesMap[nodeName]

		if exists {
			e.Nodes[3] = nodeNumber
		} else {
			e.Nodes[3] = *nodesQuantity
			nodesMap[nodeName] = *nodesQuantity
			*nodesQuantity = *nodesQuantity + 1
		}

		// Get Value
		nodeToken = LexerNextToken(lexer)
		err, e.Value = parserParseNumber(nodeToken.TokenValue)

		if err {
			fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
			return true, e
		}
	}

	if e.ElementType == ElementBJT || e.ElementType == ElementMOSFET {
		e.Nodes = make([]int, 3)

		// Get First Node
		nodeToken := LexerNextToken(lexer)
		if lexer.eof || nodeToken.TokenType != TokenStr {
			fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
			return true, e
		}

		nodeName := nodeToken.TokenValue
		nodeNumber, exists := nodesMap[nodeName]

		if exists {
			e.Nodes[0] = nodeNumber
		} else {
			e.Nodes[0] = *nodesQuantity
			nodesMap[nodeName] = *nodesQuantity
			*nodesQuantity = *nodesQuantity + 1
		}

		// Get Second Node
		nodeToken = LexerNextToken(lexer)
		if lexer.eof || nodeToken.TokenType != TokenStr {
			fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
			return true, e
		}

		nodeName = nodeToken.TokenValue
		nodeNumber, exists = nodesMap[nodeName]

		if exists {
			e.Nodes[1] = nodeNumber
		} else {
			e.Nodes[1] = *nodesQuantity
			nodesMap[nodeName] = *nodesQuantity
			*nodesQuantity = *nodesQuantity + 1
		}

		// Get Third Node
		nodeToken = LexerNextToken(lexer)
		if lexer.eof || nodeToken.TokenType != TokenStr {
			fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
			return true, e
		}

		nodeName = nodeToken.TokenValue
		nodeNumber, exists = nodesMap[nodeName]

		if exists {
			e.Nodes[2] = nodeNumber
		} else {
			e.Nodes[2] = *nodesQuantity
			nodesMap[nodeName] = *nodesQuantity
			*nodesQuantity = *nodesQuantity + 1
		}

		// Get Model
		nodeToken = LexerNextToken(lexer)
		e.Model = nodeToken.TokenValue
	}

	return false, e
}

func parserParseNumber(numberValue string) (bool, float64) {
	if parserIsNumberOnSINotation(numberValue) {
		base, _ := strconv.ParseFloat(numberValue[0:len(numberValue)-1], 64)
		var exp float64
		switch numberValue[len(numberValue)-1] {
		case 'f':
			exp = -15.0
		case 'p':
			exp = -12.0
		case 'n':
			exp = -9.0
		case 'u':
			exp = -6.0
		case 'm':
			exp = -3.0
		case 'k':
			exp = 3.0
		case 't':
			exp = 12.0
		case 'g':
			{
				if numberValue[len(numberValue)-2] == 'e' {
					exp = 6.0
					// readjust base
					base, _ = strconv.ParseFloat(numberValue[0:len(numberValue)-3], 64)
				} else {
					exp = 9.0
				}
			}
		}

		return false, base * math.Pow(10.0, exp)
	} else if parserIsNumberOnSignificandExpoentNotation(numberValue) {
		baseNumber := 0
		for numberValue[baseNumber] != 'e' {
			baseNumber = baseNumber + 1
		}
		base, _ := strconv.ParseFloat(numberValue[0:baseNumber], 64)
		exp, _ := strconv.ParseFloat(numberValue[baseNumber+1:], 64)

		return false, math.Pow(base, exp)
	} else if parserIsNumberOnRegularNotation(numberValue) {
		f, _ := strconv.ParseFloat(numberValue, 64)
		return false, f
	} else {
		return true, 0.0
	}
}

func parserIsByteNumber(b byte) bool {
	if b == '0' || b == '1' || b == '2' || b == '3' || b == '4' || b == '5' || b == '6' ||
		b == '7' || b == '8' || b == '9' {
		return true
	}
	return false
}

func parserIsNumberOnRegularNotation(numberValue string) bool {
	stringLength := len(numberValue)
	stringPosition := 0

	if stringLength == 0 {
		return false
	}

	if !(parserIsByteNumber(numberValue[stringPosition]) || numberValue[stringPosition] == '-') {
		return false
	}

	stringPosition = stringPosition + 1

	for stringPosition < stringLength && parserIsByteNumber(numberValue[stringPosition]) {
		stringPosition = stringPosition + 1
	}

	if stringPosition < stringLength && numberValue[stringPosition] == '.' {
		stringPosition = stringPosition + 1

		if stringPosition >= stringLength || !parserIsByteNumber(numberValue[stringPosition]) {
			return false
		}

		stringPosition = stringPosition + 1

		for stringPosition < stringLength && parserIsByteNumber(numberValue[stringPosition]) {
			stringPosition = stringPosition + 1
		}
	}

	if stringPosition == stringLength {
		return true
	} else {
		return false
	}
}

func parserIsNumberOnSINotation(numberValue string) bool {
	stringLength := len(numberValue)
	stringPosition := 0

	if stringLength == 0 {
		return false
	}

	if !(parserIsByteNumber(numberValue[stringPosition]) || numberValue[stringPosition] == '-') {
		return false
	}

	stringPosition = stringPosition + 1

	for stringPosition < stringLength && parserIsByteNumber(numberValue[stringPosition]) {
		stringPosition = stringPosition + 1
	}

	if stringPosition < stringLength && numberValue[stringPosition] == '.' {
		stringPosition = stringPosition + 1

		if stringPosition >= stringLength || !parserIsByteNumber(numberValue[stringPosition]) {
			return false
		}

		stringPosition = stringPosition + 1

		for stringPosition < stringLength && parserIsByteNumber(numberValue[stringPosition]) {
			stringPosition = stringPosition + 1
		}
	}

	if stringPosition >= stringLength || !(numberValue[stringPosition] == 'f' ||
		numberValue[stringPosition] == 'p' || numberValue[stringPosition] == 'n' ||
		numberValue[stringPosition] == 'u' || numberValue[stringPosition] == 'm' ||
		numberValue[stringPosition] == 'k' || numberValue[stringPosition] == 'g' ||
		numberValue[stringPosition] == 't') {
		return false
	}

	stringPosition = stringPosition + 1

	if stringPosition < stringLength-1 && numberValue[stringPosition-1] == 'm' &&
		numberValue[stringPosition] == 'e' && numberValue[stringPosition+1] == 'g' {
		stringPosition = stringPosition + 2
	}

	if stringPosition == stringLength {
		return true
	} else {
		return false
	}
}

func parserIsNumberOnSignificandExpoentNotation(numberValue string) bool {
	stringLength := len(numberValue)
	stringPosition := 0

	if stringLength == 0 {
		return false
	}

	if !parserIsByteNumber(numberValue[stringPosition]) {
		return false
	}

	stringPosition = stringPosition + 1

	for stringPosition < stringLength && parserIsByteNumber(numberValue[stringPosition]) {
		stringPosition = stringPosition + 1
	}

	if stringPosition < stringLength && numberValue[stringPosition] == '.' {
		stringPosition = stringPosition + 1

		if stringPosition >= stringLength || !parserIsByteNumber(numberValue[stringPosition]) {
			return false
		}

		stringPosition = stringPosition + 1

		for stringPosition < stringLength && parserIsByteNumber(numberValue[stringPosition]) {
			stringPosition = stringPosition + 1
		}
	}

	if stringPosition >= stringLength || !(numberValue[stringPosition] == 'e') {
		return false
	}

	stringPosition = stringPosition + 1

	if stringPosition >= stringLength || !(parserIsByteNumber(numberValue[stringPosition]) ||
		numberValue[stringPosition] == '-') {
		return false
	}

	stringPosition = stringPosition + 1

	for stringPosition < stringLength && parserIsByteNumber(numberValue[stringPosition]) {
		stringPosition = stringPosition + 1
	}

	if stringPosition == stringLength {
		return true
	} else {
		return false
	}
}

func parserAppendElement(elementList *Element, e *Element) {
	tmp := elementList

	for tmp.next != nil {
		tmp = tmp.next
	}

	tmp.next = e
}

func parserPrintElementList(elementList *Element) {
	e := elementList
	count := 1

	for e != nil {
		fmt.Printf("Element %d: \n", count)
		parserPrintElement(e)
		count = count + 1
		e = e.next
	}
}

func parserPrintElement(e *Element) {
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

	if e.ElementType == ElementBJT || e.ElementType == ElementMOSFET {
		fmt.Printf("\tModel: %s\n", e.Model)
	} else {
		fmt.Printf("\tValue: %f\n", e.Value)
	}
}

func parserPrintNodesMap(nodesMap map[string]int) {
	fmt.Printf("Nodes Map:\n")

	for k, v := range nodesMap {
		fmt.Printf("\t%s -> %d\n", k, v)
	}
}
