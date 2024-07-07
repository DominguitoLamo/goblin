package main

import (
	"fmt"
	"regexp"
)

type Token struct {
	Type   string
	Value  string
	Lineno int
	Index  int
	End    int
}

func (t *Token) String() string {
	s := fmt.Sprintf("Token(%s, %s, index: %d, end: %d, lineNO: %d)", t.Type, t.Value, t.Index, t.End, t.Lineno)
	return s
}

type Lexer struct {
	pattern *regexp.Regexp
	rules map[string]string
	ignore *regexp.Regexp
	newLineChar *regexp.Regexp
}

func CreateLexer(rules map[string]string, ignore []string, newLineChar string) *Lexer {
	ignorePattern := ""
	for _, pattern := range ignore {
		ignorePattern += pattern + "|"
	}
	// remove the last |
	ignoreReg := regexp.MustCompile(ignorePattern[:len(ignorePattern)-1])


	pattern := ""
	for key, value := range rules {
		pattern += fmt.Sprintf("(?P<%s>%s)|", key, value)
	}
	patternReg := regexp.MustCompile(pattern[:len(pattern)-1])

	return &Lexer{
		rules: rules,
		pattern: patternReg,
		ignore: ignoreReg,
		newLineChar: regexp.MustCompile(newLineChar),
	}
}

func (l *Lexer) Tokenize(text string) []*Token {
	tokens := []*Token{}
	lineno := 1
	index := 0

	for index < len(text) {
		// handle ignore case
		if l.ignore.MatchString(text[index:index+1]) {
			index++
			continue
		}

		// handle new line case
		if l.newLineChar.MatchString(text[index:index+1]) {
			lineno++
			index++
			continue
		}

		
	}
}