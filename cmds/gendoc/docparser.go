package main

import (
	"regexp"
	"sort"
	"strings"
)

const (
	NAME     = "name:"
	USAGE    = "usage:"
	PARAMS   = "params:"
	RETURN   = "return:"
	EXAMPLE  = "example:"
	RESPONSE = "response:"

	PREFIX = "prefix:"
)

type DocFormat struct {
	Usage    string
	Param    string
	Return   string
	Example  string
	Response string
}

type StructDoc struct {
	Name   string
	Tokens map[string]*Token
}

type FuncDoc struct {
	Prefix string
	Name   string
	Tokens map[string]*Token
}

type Token struct {
	Ident  string
	Start  int
	End    int
	Str    string
	Params []string
}

func structParser(source string) *StructDoc {
	setToken := []string{USAGE, NAME, PREFIX}
	idIndex := markTokenPlace(source, setToken)

	tokens := make(map[string]*Token)
	for _, value := range idIndex {
		tokens[value.Ident] = value
	}

	var name string
	nameToken, ok := tokens[NAME]
	if ok {
		name = nameToken.Str
	}

	return &StructDoc{
		Name:   name,
		Tokens: tokens,
	}
}

func funcParser(source string, prefix string) *FuncDoc {
	setToken := []string{USAGE, PARAMS, NAME, RETURN, EXAMPLE, RESPONSE}
	idIndex := markTokenPlace(source, setToken)

	tokens := make(map[string]*Token)
	for _, value := range idIndex {
		tokens[value.Ident] = value
	}

	var name string
	nameToken, ok := tokens[NAME]
	if ok {
		name = nameToken.Str
	}
	paramToken, ok := tokens[PARAMS]
	if ok {
		parserParams(paramToken)
	}

	return &FuncDoc{
		Name:   name,
		Prefix: prefix,
		Tokens: tokens,
	}
}

func markTokenPlace(source string, tokenPrefixSet []string) map[int]*Token {
	idIndex := make(map[int]*Token)
	indexs := []int{}
	for _, tokenPrefix := range tokenPrefixSet {
		index := strings.Index(source, tokenPrefix)
		idIndex[index] = &Token{
			Ident: tokenPrefix,
			Start: index,
		}
		indexs = append(indexs, index)
	}
	delete(idIndex, -1)
	sort.Ints(indexs)

	var preToken *Token
	for _, val := range indexs {
		if val != -1 {
			if preToken != nil {
				preToken.End = idIndex[val].Start
				preToken.Str = source[preToken.Start+len(preToken.Ident) : preToken.End]
				preToken.Str = strings.Trim(preToken.Str, "\t\r\n/ ")
			}
			preToken = idIndex[val]
		}
	}
	if preToken != nil {
		preToken.End = len(source)
		preToken.Str = source[preToken.Start+len(preToken.Ident) : preToken.End]
		preToken.Str = strings.Trim(preToken.Str, "\t\r\n/ ")
	}
	return idIndex
}

func parserParams(token *Token) {
	reg := regexp.MustCompile("\\d\\.")
	sub := reg.Split(token.Str, 40)
	token.Params = []string{}
	for index, paramStr := range sub {
		if index == 0 {
			continue
		}
		paramStr = strings.Trim(paramStr, "\t\r\n/ ")
		token.Params = append(token.Params, paramStr)
	}
}
