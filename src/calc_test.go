package main

import (
	"fmt"
	"strconv"
	"testing"
)

type calcParser struct {
	vars   map[string]int
	parser *Parser
}

func createCalc() *calcParser {
	vars := make(map[string]int)

	// lexer rules
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
					RFunc: func(pvals []PValue) (PValue, error) {
						key := string(pvals[0].GetValue())
						// string to int
						num, valErr := tokenValue2Int(pvals[2].GetValue())
						if valErr != nil {
							return nil, valErr
						}
						vars[key] = num

						return &Token {
							Type: "NUMBER",
							Value: "0",
						}, nil
					},
				},
				{
					Ops: "expr",
					RFunc: func(pvals []PValue) (PValue, error) {
						return pvals[0], nil
					},
				},
			},
		},
		{
			Name: "expr",
			Expand: []*RuleOps {
				{
					Ops: "expr PLUS expr",
					RFunc: func(pvals []PValue) (PValue, error) {
						// string to int
						num1, valErr := tokenValue2Int(pvals[0].GetValue())
						if valErr != nil {
							return nil, valErr
						}
						num2, valErr := tokenValue2Int(pvals[2].GetValue())
						if valErr != nil {
							return nil, valErr
						}
						num := num1 + num2
						return &Token {
							Type: "NUMBER",
							Value: fmt.Sprintf("%d", num),
						}, nil
					},
				},

				{
					Ops: "expr MINUS expr",
					RFunc: func(pvals []PValue) (PValue, error) {
						// string to int
						num1, valErr := tokenValue2Int(pvals[0].GetValue())
						if valErr != nil {
							return nil, valErr
						}
						num2, valErr := tokenValue2Int(pvals[2].GetValue())
						if valErr != nil {
							return nil, valErr
						}
						num := num1 - num2
						return &Token {
							Type: "NUMBER",
							Value: fmt.Sprintf("%d", num),
						}, nil
					},
				},

				{
					Ops: "expr MULTIPLY expr",
					RFunc: func(pvals []PValue) (PValue, error) {
						// string to int
						num1, valErr := tokenValue2Int(pvals[0].GetValue())
						if valErr != nil {
							return nil, valErr
						}
						num2, valErr := tokenValue2Int(pvals[2].GetValue())
						if valErr != nil {
							return nil, valErr
						}
						num := num1 * num2
						return &Token {
							Type: "NUMBER",
							Value: fmt.Sprintf("%d", num),
						}, nil
					},
				},

				{
					Ops: "expr DIVIDE expr",
					RFunc: func(pvals []PValue) (PValue, error) {
						// string to int
						num1, valErr := tokenValue2Int(pvals[0].GetValue())
						if valErr != nil {
							return nil, valErr
						}
						num2, valErr := tokenValue2Int(pvals[2].GetValue())
						if valErr != nil {
							return nil, valErr
						}
						num := num1 / num2
						return &Token {
							Type: "NUMBER",
							Value: fmt.Sprintf("%d", num),
						}, nil
					},
				},

				{
					Ops: "MINUS expr %prec UMINUS",
					RFunc: func(pvals []PValue) (PValue, error) {
						// string to int
						num, valErr := tokenValue2Int(pvals[2].GetValue())
						if valErr != nil {
							return nil, valErr
						}
						num = -num
						return &Token {
							Type: "NUMBER",
							Value: fmt.Sprintf("%d", num),
						}, nil
					},
				},

				{
					Ops: "LPAREN expr RPAREN",
					RFunc: func(pvals []PValue) (PValue, error) {
						return pvals[1], nil
					},
				},

				{
					Ops: "NUMBER",
					RFunc: func(pvals []PValue) (PValue, error) {
						return pvals[0], nil
					},
				},

				{
					Ops: "NAME",
					RFunc: func(pvals []PValue) (PValue, error) {
						value := string(pvals[0].GetValue())
						num, ok := vars[value]
						if !ok {
							return nil, fmt.Errorf("undefined variable: %s", value)
						}
						return &Token {
							Type: "NUMBER",
							Value: fmt.Sprintf("%d", num),
						}, nil
					},
				},
				
			},
		},
	}

	return &calcParser{
		vars: vars,
		parser: CreateParser(symbols, ignores, rules, precedences),
	}
}

func (c *calcParser) parse(input string)  {
	result, err := c.parser.Parse(input)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// string to int
	num, valErr := tokenValue2Int(result.GetValue())
	if valErr != nil {
		fmt.Printf("Error: %v\n", valErr)
	}

	fmt.Printf("result is %d", num)
}

func tokenValue2Int(val []byte) (int, error) {
	// Use strconv.Atoi to convert the string to an integer.
	// It returns the converted integer and a potential error.
	num, err := strconv.Atoi(string(val))
	if err != nil {
		// Handle the error, for example, by returning 0 and the error.
		return 0, fmt.Errorf("error converting string to int: %w", err)
	}
	return num, nil
}

func TestCalc(t *testing.T) {
	calc := createCalc()
	calc.parse("1 + 2 * 3")
}
