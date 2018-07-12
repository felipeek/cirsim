package internal

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

func ParserInit(netListPath string) {
	var token Token
	lexer := LexerInit(netListPath)
	nodesMap := make(map[string]int)
	nodesQuantity := 1
	nodesMap["0"] = 0
	var elementList *Element = nil
	opCommand := false
	tranCommand := false
	tStep := 0.0
	tStop := 0.0

	for !lexer.eof {
		token = LexerNextToken(&lexer)
		switch token.TokenType {
		case TokenLineBreak:
			{

			}
		case TokenCommand:
			{
				if token.TokenValue == ".op" {
					opCommand = true
				} else if token.TokenValue == ".tran" {
					tranCommand = true
					// Get Value
					nodeToken := LexerNextToken(&lexer)
					err, step := parserParseNumber(nodeToken.TokenValue)

					if err {
						fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", lexer.lineNumber)
						return
					}

					nodeToken = LexerNextToken(&lexer)
					err, stop := parserParseNumber(nodeToken.TokenValue)

					if err {
						fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", lexer.lineNumber)
						return
					}

					tStep = step
					tStop = stop
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
						elementListAppend(elementList, e)
					} else {
						elementList = e
					}
				}
			}
		}
	}

	if opCommand {
		mnaSolveLinear(elementList, nodesMap)
	}
	if tranCommand {
		mnaSolveDynamic(elementList, nodesMap, tStep, tStop)
	}
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

	e.Label = elementArray
	e.Next = nil
	e.PreserveCurrent = false

	if e.ElementType == ElementResistor || e.ElementType == ElementCapacitor ||
		e.ElementType == ElementInductor || e.ElementType == ElementDiode {
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

		// If the element is a capactor or an inductor, there is an optional parameter
		if e.ElementType == ElementCapacitor ||
			e.ElementType == ElementInductor {
			// Get Value
			nodeToken = LexerNextToken(lexer)
			if nodeToken.TokenType != TokenLineBreak {
				err, e.Extra = parserParseIC(nodeToken.TokenValue)
				if err {
					fmt.Fprintf(os.Stderr, "Parser Error: IC format error at line %d\n", currentLine)
					return true, e
				}
			} else {
				e.Extra = 0.0
			}
		}
	}

	if e.ElementType == ElementVCCS || e.ElementType == ElementVCVS {
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

	if e.ElementType == ElementVoltageSource ||
		e.ElementType == ElementCurrentSource {
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

		switch nodeToken.TokenValue[0] {
		case 's':
			desc := sinDescriptor{}
			if nodeToken.TokenValue[:3] != "sin" {
				fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
				return true, e
			}
			if len(nodeToken.TokenValue) > 3 {
				if nodeToken.TokenValue[3] != '(' {
					fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
					return true, e
				}
				if len(nodeToken.TokenValue) > 4 {
					err, desc.v0 = parserParseNumber(nodeToken.TokenValue[4:])
					if err {
						fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
						return true, e
					}
				} else {
					nodeToken = LexerNextToken(lexer)
					err, desc.v0 = parserParseNumber(nodeToken.TokenValue)
					if err {
						fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
						return true, e
					}
				}
			} else {
				nodeToken = LexerNextToken(lexer)
				if nodeToken.TokenValue[0] != '(' {
					fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
					return true, e
				}
				if len(nodeToken.TokenValue) > 1 {
					err, desc.v0 = parserParseNumber(nodeToken.TokenValue[1:])
					if err {
						fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
						return true, e
					}
				} else {
					nodeToken = LexerNextToken(lexer)
					err, desc.v0 = parserParseNumber(nodeToken.TokenValue)
					if err {
						fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
						return true, e
					}
				}
			}

			nodeToken = LexerNextToken(lexer)
			err, desc.va = parserParseNumber(nodeToken.TokenValue)
			if err {
				fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
				return true, e
			}

			nodeToken = LexerNextToken(lexer)
			err, desc.freq = parserParseNumber(nodeToken.TokenValue)
			if err {
				fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
				return true, e
			}

			nodeToken = LexerNextToken(lexer)
			if nodeToken.TokenValue[len(nodeToken.TokenValue)-1] != ')' {
				err, desc.td = parserParseNumber(nodeToken.TokenValue)
				if err {
					fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
					return true, e
				}

				nodeToken = LexerNextToken(lexer)
				if nodeToken.TokenValue != ")" {
					fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
					return true, e
				}
			} else {
				err, desc.td = parserParseNumber(nodeToken.TokenValue[:len(nodeToken.TokenValue)-1])
				if err {
					fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
					return true, e
				}
			}
			e.Extra = desc
		case 'p':
			ar := make([]pwlDescriptor, 0)
			desc := pwlDescriptor{}
			if nodeToken.TokenValue[:3] != "pwl" {
				fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
				return true, e
			}
			if len(nodeToken.TokenValue) > 3 {
				if nodeToken.TokenValue[3] != '(' {
					fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
					return true, e
				}
				if len(nodeToken.TokenValue) > 4 {
					err, desc.t = parserParseNumber(nodeToken.TokenValue[4:])
					if err {
						fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
						return true, e
					}
				} else {
					nodeToken = LexerNextToken(lexer)
					err, desc.t = parserParseNumber(nodeToken.TokenValue)
					if err {
						fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
						return true, e
					}
				}
			} else {
				nodeToken = LexerNextToken(lexer)
				if nodeToken.TokenValue[0] != '(' {
					fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
					return true, e
				}
				if len(nodeToken.TokenValue) > 1 {
					err, desc.t = parserParseNumber(nodeToken.TokenValue[1:])
					if err {
						fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
						return true, e
					}
				} else {
					nodeToken = LexerNextToken(lexer)
					err, desc.t = parserParseNumber(nodeToken.TokenValue)
					if err {
						fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
						return true, e
					}
				}
			}

			nodeToken = LexerNextToken(lexer)
			if nodeToken.TokenValue[len(nodeToken.TokenValue)-1] != ')' {
				err, desc.x = parserParseNumber(nodeToken.TokenValue)
				if err {
					fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
					return true, e
				}
				ar = append(ar, desc)

				for {
					nodeToken = LexerNextToken(lexer)
					if nodeToken.TokenValue == ")" {
						break
					}
					err, desc.t = parserParseNumber(nodeToken.TokenValue)
					if err {
						fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
						return true, e
					}

					nodeToken = LexerNextToken(lexer)
					if nodeToken.TokenValue[len(nodeToken.TokenValue)-1] != ')' {
						err, desc.x = parserParseNumber(nodeToken.TokenValue)
						if err {
							fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
							return true, e
						}
						ar = append(ar, desc)
					} else {
						err, desc.x = parserParseNumber(nodeToken.TokenValue[:len(nodeToken.TokenValue)-1])
						if err {
							fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
							return true, e
						}
						ar = append(ar, desc)
						break
					}
				}
			} else {
				err, desc.x = parserParseNumber(nodeToken.TokenValue[:len(nodeToken.TokenValue)-1])
				if err {
					fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
					return true, e
				}
				ar = append(ar, desc)
			}

			e.Extra = ar
		default:
			e.Extra = nil
			err, e.Value = parserParseNumber(nodeToken.TokenValue)

			if err {
				fmt.Fprintf(os.Stderr, "Parser Error: Number format error at line %d\n", currentLine)
				return true, e
			}
		}
	}

	if e.ElementType == ElementCCVS || e.ElementType == ElementCCCS {
		e.Nodes = make([]int, 2)

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

		// Get Control Element
		nodeToken = LexerNextToken(lexer)
		if lexer.eof || nodeToken.TokenType != TokenStr {
			fmt.Fprintf(os.Stderr, "Parser Error: Element format error at line %d\n", currentLine)
			return true, e
		}

		nodeName = nodeToken.TokenValue
		e.Extra = nodeName

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
		e.Extra = nodeToken.TokenValue
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

func parserPrintNodesMap(nodesMap map[string]int) {
	fmt.Printf("Nodes Map:\n")

	for k, v := range nodesMap {
		fmt.Printf("\t%s -> %d\n", k, v)
	}
}

func parserParseIC(icString string) (bool, float64) {
	if icString[:3] != "ic=" {
		return true, 0.0
	}

	return parserParseNumber(icString[3:])
}
