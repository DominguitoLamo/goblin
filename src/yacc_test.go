package main

import (
	"fmt"
	"testing"
)

func TestCreateGrammar(t *testing.T) {
	symbols := map[string]string {
		"NUMBER": "[0-9]+",
		"PLUS": "\\+",
	}
    ignores := []string{
        "\t"," ",
    }

	l := CreateLexer(symbols, ignores)

	rules := []*SyntaxRule {
		{
			Name: "expr",
			Expend: []*RuleOps {
				 {
					Ops: "terms PLUS terms",
				},
			},
		},
		{
			Name: "terms",
			Expend: []*RuleOps {
				 {
					Ops: "NUMBER",
				},
			},
		},
	}

	g := CreateGrammar(l, rules, []*Precedence{})
	result := g.string()
	fmt.Println(result)
}