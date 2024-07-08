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
	redefine map[string]map[string]string
	rules map[string]string
	ignore *regexp.Regexp
}

func CreateLexer(rules map[string]string, ignore []string) *Lexer {
	ignorePattern := ""
	redefine := map[string]map[string]string{}

	for _, pattern := range ignore {
		ignorePattern += pattern + "|"
	}
	// remove the last |
	ignoreReg := regexp.MustCompile(ignorePattern[:len(ignorePattern)-1])


	pattern := ""
	for key, value := range rules {
		if isIn, tokenType, keywords := isRedefine(key); isIn {
			_, ok := redefine[tokenType]
			if !ok {
				redefine[tokenType] = map[string]string{
					value: keywords,
				}
			} else {
				redefine[tokenType][value] = keywords
			}
			continue
		}

		pattern += fmt.Sprintf("(?P<%s>%s)|", key, value)
	}
	patternReg := regexp.MustCompile(pattern[:len(pattern)-1])

	return &Lexer{
		rules: rules,
		redefine: redefine,
		pattern: patternReg,
		ignore: ignoreReg,
	}
}

func isRedefine(key string) (bool, string, string) {
	reg := regexp.MustCompile(`([A-Z]+)\[([A-Z]+)\]`)
	match := reg.FindStringSubmatch(key)
	if match == nil {
		return false, "", ""
	}

	if len(match) != 3 {
		return false, "", ""
	}
	return true, match[1], match[2]
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
		if text[index:index+1] == "\n" {

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

		// handle redefine case
		if typeMap, ok := l.redefine[token.Type]; ok {
			keyword, valOk := typeMap[token.Value]
			if valOk {
				token.Type = keyword
			}
		}

		index += longestLen
		tokens = append(tokens, token)
	}
	return tokens, nil
}