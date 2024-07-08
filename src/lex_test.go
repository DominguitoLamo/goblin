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

func TestRedefineReg(t *testing.T) {
    is, tokenType, keywords := isRedefine("NAME[INT]")
    fmt.Printf("%v %s %s\n", is, tokenType, keywords)
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
        "\t"," ",
    }

    l := CreateLexer(symbols, ignores)
    tokens, err := l.Tokenize("a\na = 1 + 2 * (3 - 4)\nb = 2 - 1\nc = 3 * 4 / 2")
    if err != nil {
        t.Error(err)
    }

    for _, token := range tokens {
        fmt.Printf("%s", token.String())
    }
}

func TestConflict(t *testing.T) {
    symbols := map[string]string {
        "NAME": "[a-zA-Z_][a-zA-Z0-9_]*",
        "NAME[INT]": "int",
        "NAME[IF]": "if",

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
        " ",
        "\t",
    }

    l := CreateLexer(symbols, ignores)
    // int a = 12\nif a == 12\na + 12\n
    tokens, err := l.Tokenize("if a == 12")

    if err != nil {
        t.Error(err)
    }

    for _, token := range tokens {
        fmt.Printf("%s", token.String())
    }
    fmt.Printf("\n")

    tokens1, err := l.Tokenize("int a = 12\nif a == 12\na + 12\n")

    if err != nil {
        t.Error(err)
    }

    for _, token := range tokens1 {
        fmt.Printf("%s", token.String())
    }
}