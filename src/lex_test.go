package main

import (
	"fmt"
	"regexp"
	"testing"
)

func TestReg(t *testing.T) {
    re := regexp.MustCompile(`(?P<variable>[a-zA-Z_][a-zA-Z0-9_]*) = (?P<value>.*)`)
    text := "name = John Doe\nage = 30"

    matches := re.FindStringSubmatch(text)
    if len(matches) > 0 {
        for i, name := range re.SubexpNames() {
            if i != 0 && name != "" { // Skip the entire match
                fmt.Printf("%s: %s\n", name, matches[i])
            }
        }
    }
}

func TestSimpleLex(t *testing.T) {
    symbols := map[string]string {
        "NAME": "[a-zA-Z_][a-zA-Z0-9_]*",
        "NUMBER": "[0-9]+",

        "PLUS": "\\+",
        "MINUS": "\\-",
        "MULTIPLY": "\\*",
        "DIVIDE": "/",
        "ASSIGN": "=",

        "LPAREN": "\\(",
        "RPAREN": "\\)",
    }

    ignores := []string{
        "\\s+",
        "\t",
    }

    l := CreateLexer(symbols, ignores, "\n+")
    tokens, err := l.Tokenize("a = 1 + 2 * (3 - 4)")
    if err != nil {
        t.Error(err)
    }

    for _, token := range tokens {
        fmt.Printf("%s", token.String())
    }
}