package main

import (
	"fmt"

	"github.com/felipeek/cirsim/internal"
)

func main() {
	l := internal.LexerInit("C:\\Users\\Felipe\\Development\\go\\src\\github.com\\felipeek\\cirsim\\res\\example.spice")

	flag := false

	for !flag {
		token := internal.LexerNextToken(&l)
		if token.TokenType == internal.TokenCommand {
			fmt.Printf("Conseguiu chegar. Valor: %s", token.TokenValue)
			flag = true
		}
	}
}
