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
			Expand: []*RuleOps {
				 {
					Ops: "terms PLUS terms",
				},
			},
		},
		{
			Name: "terms",
			Expand: []*RuleOps {
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

func TestCheckGrammar(t *testing.T) {
	symbols := map[string]string {
		"NUMBER": "[0-9]+",
		"PLUS": "\\+",
		"MINUS": "-",
	}
    ignores := []string{
        "\t"," ",
    }

	l := CreateLexer(symbols, ignores)

	rules := []*SyntaxRule {
		{
			Name: "s",
			Expand: []*RuleOps {
				 {
					Ops: "e PLUS e",
				},
			},
		},
		{
			Name: "e",
			Expand: []*RuleOps {
				 {
					Ops: "r",
				},
			},
		},
		{
			Name: "r",
			Expand: []*RuleOps {
				 {
					Ops: "s",
				},
			},
		},
		{
			Name: "t",
			Expand: []*RuleOps {
				 {
					Ops: "NUMBER",
				},
			},
		},
	}

	CreateGrammar(l, rules, []*Precedence{})
}

func TestFirstAndFollow(t *testing.T) {
	g := createCalcGrammar()
	result := g.string()
	fmt.Println(result)
}

func createCalcGrammar() *grammar {
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

	precedences := []*Precedence {
		{
			TokenType: []string {
				"PLUS",
				"MINUS",
			},
			Level: 1,
		},
		{
			TokenType: []string {
				"MULTIPLY",
				"DIVIDE",
			},
			Level: 2,
		},
		{
			TokenType: []string {
				"UMINUS",
			},
			Level: 3,
		},
	}

	rules := []*SyntaxRule {
		{
			Name: "statement",
			Expand: []*RuleOps {
				{
					Ops: "NAME ASSIGN expr",
				},
				{
					Ops: "expr",
				},
			},
		},
		{
			Name: "expr",
			Expand: []*RuleOps {
				{
					Ops: "expr PLUS expr",
				},
				{
					Ops: "expr MINUS expr",
				},
				{
					Ops: "expr MULTIPLY expr",
				},
				{
					Ops: "expr DIVIDE expr",
				},
				{
					Ops: "MINUS expr %prec UMINUS",
				},
				{
					Ops: "LPAREN expr RPAREN",
				},
				{
					Ops: "NUMBER",
				},
				{
					Ops: "NAME",
				},
			},
		},
	}
	

	g := CreateGrammar(l, rules, precedences)

	return g
}

func TestLRTable(t *testing.T) {
	g := createCalcGrammar()
	CreateLRTable(g)
}