package internal

import (
	"fmt"
	"io/ioutil"
)

type Lexer struct {
	netlistFile []byte
	position    int
}

type TokenType int

const (
	TokenStr       TokenType = iota
	TokenNumber    TokenType = iota
	TokenLineBreak TokenType = iota
	TokenCommand   TokenType = iota
)

type Token struct {
	TokenType  TokenType
	TokenValue []byte
}

func LexerInit(netlistPath string) Lexer {
	data, err := ioutil.ReadFile(netlistPath)

	if err != nil {
		panic(err)
	}

	fmt.Printf("File contents: %s\n", data)

	lexer := Lexer{
		netlistFile: data,
		position:    0}

	// Ignore file's first line
	lexerIgnoreLine(&lexer)

	return lexer
}

func LexerNextToken(lexer *Lexer) Token {
	var newToken Token

	newToken.TokenType = TokenStr

	// If lexeme starts with '*', we ignore everything until the end of the line
	for lexer.netlistFile[lexer.position] == '*' {
		lexerIgnoreLine(lexer)
	}

	// If lexeme starts with '\n', we assume it is just a line break
	if lexer.netlistFile[lexer.position] == '\n' {
		lexer.position = lexer.position + 1
		newToken.TokenType = TokenLineBreak
		return newToken
	}

	if lexer.netlistFile[lexer.position] == '.' {
		// If lexeme starts with '.', we assume it is a command
		newToken.TokenType = TokenCommand
	} else if lexerLexemeStartsWithNumber(lexer) {
		// If lexeme starts with a number, we assume it is an integer or a float
		newToken.TokenType = TokenNumber
	} else {
		// Last case: We assume lexeme is a string
		newToken.TokenType = TokenStr
	}

	valueStartPosition := lexer.position
	for lexer.netlistFile[lexer.position] != ' ' {
		lexer.position = lexer.position + 1
	}
	newToken.TokenValue = make([]byte, lexer.position-valueStartPosition)
	copy(newToken.TokenValue, lexer.netlistFile[valueStartPosition:lexer.position])
	return newToken
}

func lexerIgnoreLine(lexer *Lexer) {
	for lexer.netlistFile[lexer.position] != '\n' {
		lexer.position = lexer.position + 1
	}
	lexer.position = lexer.position + 1
}

func lexerLexemeStartsWithNumber(lexer *Lexer) bool {
	if lexer.netlistFile[lexer.position] == 0 ||
		lexer.netlistFile[lexer.position] == 1 ||
		lexer.netlistFile[lexer.position] == 2 ||
		lexer.netlistFile[lexer.position] == 3 ||
		lexer.netlistFile[lexer.position] == 4 ||
		lexer.netlistFile[lexer.position] == 5 ||
		lexer.netlistFile[lexer.position] == 6 ||
		lexer.netlistFile[lexer.position] == 7 ||
		lexer.netlistFile[lexer.position] == 8 ||
		lexer.netlistFile[lexer.position] == 9 {
		return true
	}

	return false
}
