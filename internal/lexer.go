package internal

import (
	"io/ioutil"
	"strings"
)

type Lexer struct {
	netlistFile []byte
	position    int
	lineNumber  int
	eof         bool
}

type TokenType int

const (
	TokenStr       TokenType = iota
	TokenLineBreak TokenType = iota
	TokenCommand   TokenType = iota
)

type Token struct {
	TokenType  TokenType
	TokenValue string
}

func LexerInit(netlistPath string) Lexer {
	data, err := ioutil.ReadFile(netlistPath)

	if err != nil {
		panic(err)
	}

	lexer := Lexer{
		netlistFile: data,
		position:    0,
		lineNumber:  1,
		eof:         false}

	// Ignore file's first line
	lexerIgnoreLine(&lexer)

	return lexer
}

func LexerNextToken(lexer *Lexer) Token {
	var newToken Token

	if lexer.eof {
		return newToken
	}

	// Ignore comments and spaces
	lexerJumpCommentsAndSpaces(lexer)

	// If lexeme starts with '\n', we assume it is just a line break
	if lexer.netlistFile[lexer.position] == '\n' || lexer.netlistFile[lexer.position] == '\r' {
		if lexer.netlistFile[lexer.position] == '\r' && lexer.position < len(lexer.netlistFile)-1 &&
			lexer.netlistFile[lexer.position+1] == '\n' {
			lexer.position = lexer.position + 2
		} else {
			lexer.position = lexer.position + 1
		}
		lexer.lineNumber = lexer.lineNumber + 1
		newToken.TokenType = TokenLineBreak
		lexer.eof = lexer.position == len(lexer.netlistFile)
		return newToken
	}

	if lexer.netlistFile[lexer.position] == '.' {
		// If lexeme starts with '.', we assume it is a command
		newToken.TokenType = TokenCommand
	} else {
		// Last case: We assume lexeme is a string
		newToken.TokenType = TokenStr
	}

	valueStartPosition := lexer.position
	for lexer.position < len(lexer.netlistFile) && !lexerCurrentByteIsSpace(lexer) &&
		lexer.netlistFile[lexer.position] != '\n' && lexer.netlistFile[lexer.position] != '\r' {
		lexer.position = lexer.position + 1
	}
	newToken.TokenValue = strings.ToLower(string(lexer.netlistFile[valueStartPosition:lexer.position]))
	lexer.eof = lexer.position == len(lexer.netlistFile)
	return newToken
}

func lexerIgnoreLine(lexer *Lexer) {
	for lexer.position < len(lexer.netlistFile) && lexer.netlistFile[lexer.position] != '\n' &&
		lexer.netlistFile[lexer.position] != '\r' {
		lexer.position = lexer.position + 1
	}

	if lexer.position < len(lexer.netlistFile) && lexer.netlistFile[lexer.position] == '\r' {
		lexer.position = lexer.position + 1
	}

	if lexer.position < len(lexer.netlistFile) && lexer.netlistFile[lexer.position] == '\n' {
		lexer.position = lexer.position + 1
		lexer.lineNumber = lexer.lineNumber + 1
	}

	if lexer.position >= len(lexer.netlistFile) {
		lexer.eof = true
	}
}

func lexerCurrentByteIsSpace(lexer *Lexer) bool {
	if lexer.netlistFile[lexer.position] == ' ' ||
		lexer.netlistFile[lexer.position] == '\t' {
		return true
	}

	return false
}

func lexerJumpCommentsAndSpaces(lexer *Lexer) {
	isComment := lexer.netlistFile[lexer.position] == '*'
	isSpace := lexerCurrentByteIsSpace(lexer)
	for isComment || isSpace {
		if isComment {
			lexerIgnoreLine(lexer)
		} else if isSpace {
			lexer.position = lexer.position + 1
		}

		if lexer.position >= len(lexer.netlistFile) {
			return
		}

		isComment = lexer.netlistFile[lexer.position] == '*'
		isSpace = lexerCurrentByteIsSpace(lexer)
	}
}
