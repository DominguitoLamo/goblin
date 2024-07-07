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
	s := fmt.Sprintf("Token(%s, %s, index: %d, end: %d, lineNO: %d)\n", t.Type, t.Value, t.Index, t.End, t.Lineno)
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

func (l *Lexer) Tokenize(text string) ([]*Token, error) {
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

		match := l.pattern.FindStringSubmatch(text[index:])
		if match == nil {
			return nil, fmt.Errorf("invalid token at index %d, line %d", index, lineno)
		}

		token := &Token{}
		longestLen := 0
		for i, name := range l.pattern.SubexpNames() {
			if i != 0 && name != "" && len(match[i]) > 0 {

				token.Type = name
				token.Value = match[i]
				token.Index = index
				token.End = index + len(match[i])
				token.Lineno = lineno

				if len(match[i]) > longestLen {
					longestLen = len(match[i])
				}
			}
		}
		index += longestLen
		tokens = append(tokens, token)
	}
	return tokens, nil
}