# Goblin (Go Bite Linguistics)

*lex and yacc written in go*

## Introduction

This project is the implement of LALR(1) for the instruction use of compiler course.

## Quick Start

```golang
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
	calc.parse("a = 1 + 2")
	calc.parse("b = a + 2")
	calc.parse("a + b")
}
```

## Write LR Table into Markdown

To study the process of LR table generation, you can use the following code to write the LR table into a markdown file.

```golang
func TestCreateParser(t *testing.T) {
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
	
	p := CreateParser(symbols, ignores, rules, precedences)
	p.WriteMDInfo("calc", "../")
}
```

The result is:

```markdown
# Lexer

- MINUS : \- 
- LPAREN : \( 
- RPAREN : \) 
- NAME : [a-zA-Z_][a-zA-Z0-9_]* 
- NUMBER : [0-9]+ 
- PLUS : \+ 
- MULTIPLY : \* 
- DIVIDE : / 
- ASSIGN : = 

# Grammar

## Terminates

- NUMBER : {9, } 
- PLUS : {3, } 
- DIVIDE : {6, } 
- ASSIGN : {1, } 
- RPAREN : {8, } 
- $end : {0, } 
- NAME : {1, 10, } 
- MULTIPLY : {5, } 
- LPAREN : {8, } 
- MINUS : {4, 7, } 

## Non-Terminates

- S' : {} 
- statement : {0, } 
- expr : {2, 3, 4, 5, 6, 7, 8, 1, } 

 Precedence

- PLUS : 1 
- MINUS : 1 
- MULTIPLY : 2 
- DIVIDE : 2 
- UMINUS : 3 

## Productions

### <a id=P0></a>P0. S' -> statement $end 
### <a id=P1></a>P1. statement -> NAME ASSIGN expr 
### <a id=P2></a>P2. statement -> expr 
### <a id=P3></a>P3. expr -> expr PLUS expr 
### <a id=P4></a>P4. expr -> expr MINUS expr 
### <a id=P5></a>P5. expr -> expr MULTIPLY expr 
### <a id=P6></a>P6. expr -> expr DIVIDE expr 
### <a id=P7></a>P7. expr -> MINUS expr 
### <a id=P8></a>P8. expr -> LPAREN expr RPAREN 
### <a id=P9></a>P9. expr -> NUMBER 
### <a id=P10></a>P10. expr -> NAME 
# LR Table

## States

# <a id=S0></a>S0

- S' -> . statement $end 
- statement -> . NAME ASSIGN expr 
- statement -> . expr 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S1></a>S1

- expr -> NUMBER . 

    lookahead: {MINUS, MULTIPLY, DIVIDE, $end, RPAREN, PLUS, }


# <a id=S2></a>S2

- statement -> NAME . ASSIGN expr 
- expr -> NAME . 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, $end, }


# <a id=S3></a>S3

- statement -> expr . 

    lookahead: {$end, }

- expr -> expr . PLUS expr 
- expr -> expr . MINUS expr 
- expr -> expr . MULTIPLY expr 
- expr -> expr . DIVIDE expr 

# <a id=S4></a>S4

- expr -> MINUS . expr 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S5></a>S5

- S' -> statement . $end 

# <a id=S6></a>S6

- expr -> LPAREN . expr RPAREN 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S7></a>S7

- statement -> NAME ASSIGN . expr 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S8></a>S8

- expr -> expr PLUS . expr 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S9></a>S9

- expr -> expr MINUS . expr 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S10></a>S10

- expr -> expr MULTIPLY . expr 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S11></a>S11

- expr -> expr DIVIDE . expr 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S12></a>S12

- expr -> MINUS expr . 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, $end, RPAREN, }

- expr -> expr . PLUS expr 
- expr -> expr . MINUS expr 
- expr -> expr . MULTIPLY expr 
- expr -> expr . DIVIDE expr 

# <a id=S13></a>S13

- expr -> NAME . 

    lookahead: {MINUS, RPAREN, MULTIPLY, DIVIDE, $end, PLUS, }


# <a id=S14></a>S14

- S' -> statement $end . 

# <a id=S15></a>S15

- expr -> LPAREN expr . RPAREN 
- expr -> expr . PLUS expr 
- expr -> expr . MINUS expr 
- expr -> expr . MULTIPLY expr 
- expr -> expr . DIVIDE expr 

# <a id=S16></a>S16

- statement -> NAME ASSIGN expr . 

    lookahead: {$end, }

- expr -> expr . PLUS expr 
- expr -> expr . MINUS expr 
- expr -> expr . MULTIPLY expr 
- expr -> expr . DIVIDE expr 

# <a id=S17></a>S17

- expr -> expr PLUS expr . 

    lookahead: {RPAREN, PLUS, MINUS, MULTIPLY, DIVIDE, $end, }

- expr -> expr . PLUS expr 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, $end, RPAREN, }

- expr -> expr . MINUS expr 

    lookahead: {MINUS, MULTIPLY, RPAREN, DIVIDE, $end, PLUS, }

- expr -> expr . MULTIPLY expr 

    lookahead: {DIVIDE, $end, RPAREN, PLUS, MINUS, MULTIPLY, }

- expr -> expr . DIVIDE expr 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, $end, RPAREN, }


# <a id=S18></a>S18

- expr -> expr MINUS expr . 

    lookahead: {MINUS, RPAREN, MULTIPLY, DIVIDE, $end, PLUS, }

- expr -> expr . PLUS expr 

    lookahead: {MINUS, MULTIPLY, DIVIDE, $end, RPAREN, PLUS, }

- expr -> expr . MINUS expr 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, $end, RPAREN, }

- expr -> expr . MULTIPLY expr 

    lookahead: {$end, PLUS, MINUS, RPAREN, MULTIPLY, DIVIDE, }

- expr -> expr . DIVIDE expr 

    lookahead: {MULTIPLY, RPAREN, DIVIDE, $end, PLUS, MINUS, }


# <a id=S19></a>S19

- expr -> expr MULTIPLY expr . 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, RPAREN, $end, }

- expr -> expr . PLUS expr 

    lookahead: {RPAREN, PLUS, MINUS, MULTIPLY, DIVIDE, $end, }

- expr -> expr . MINUS expr 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, $end, RPAREN, }

- expr -> expr . MULTIPLY expr 

    lookahead: {MULTIPLY, DIVIDE, $end, PLUS, MINUS, RPAREN, }

- expr -> expr . DIVIDE expr 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, $end, RPAREN, }


# <a id=S20></a>S20

- expr -> expr DIVIDE expr . 

    lookahead: {RPAREN, PLUS, MINUS, MULTIPLY, DIVIDE, $end, }

- expr -> expr . PLUS expr 

    lookahead: {DIVIDE, $end, RPAREN, PLUS, MINUS, MULTIPLY, }

- expr -> expr . MINUS expr 

    lookahead: {PLUS, MINUS, MULTIPLY, RPAREN, DIVIDE, $end, }

- expr -> expr . MULTIPLY expr 

    lookahead: {MINUS, MULTIPLY, DIVIDE, $end, RPAREN, PLUS, }

- expr -> expr . DIVIDE expr 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, $end, RPAREN, }


# <a id=S21></a>S21

- expr -> LPAREN expr RPAREN . 

    lookahead: {RPAREN, PLUS, MINUS, MULTIPLY, DIVIDE, $end, }


## Action Table

| State/Terminates | DIVIDE  | ASSIGN  | RPAREN  | $end  | NUMBER  | PLUS  | LPAREN  | MINUS  | NAME  | MULTIPLY |
| ---  | ---  | ---  | ---  | ---  | ---  | ---  | ---  | ---  | ---  | --- |
| [S0](#S0) | none | none | none | none | [s1](#S1) | none | [s6](#S6) | [s4](#S4) | [s2](#S2) | none |
| [S1](#S1) | [r9](#P9) | none | [r9](#P9) | [r9](#P9) | none | [r9](#P9) | none | [r9](#P9) | none | [r9](#P9) |
| [S2](#S2) | [r10](#P10) | [s7](#S7) | none | [r10](#P10) | none | [r10](#P10) | none | [r10](#P10) | none | [r10](#P10) |
| [S3](#S3) | [s11](#S11) | none | none | [r2](#P2) | none | [s8](#S8) | none | [s9](#S9) | none | [s10](#S10) |
| [S4](#S4) | none | none | none | none | [s1](#S1) | none | [s6](#S6) | [s4](#S4) | [s13](#S13) | none |
| [S5](#S5) | none | none | none | [s14](#S14) | none | none | none | none | none | none |
| [S6](#S6) | none | none | none | none | [s1](#S1) | none | [s6](#S6) | [s4](#S4) | [s13](#S13) | none |
| [S7](#S7) | none | none | none | none | [s1](#S1) | none | [s6](#S6) | [s4](#S4) | [s13](#S13) | none |
| [S8](#S8) | none | none | none | none | [s1](#S1) | none | [s6](#S6) | [s4](#S4) | [s13](#S13) | none |
| [S9](#S9) | none | none | none | none | [s1](#S1) | none | [s6](#S6) | [s4](#S4) | [s13](#S13) | none |
| [S10](#S10) | none | none | none | none | [s1](#S1) | none | [s6](#S6) | [s4](#S4) | [s13](#S13) | none |
| [S11](#S11) | none | none | none | none | [s1](#S1) | none | [s6](#S6) | [s4](#S4) | [s13](#S13) | none |
| [S12](#S12) | [r7](#P7) | none | [r7](#P7) | [r7](#P7) | none | [r7](#P7) | none | [r7](#P7) | none | [r7](#P7) |
| [S13](#S13) | [r10](#P10) | none | [r10](#P10) | [r10](#P10) | none | [r10](#P10) | none | [r10](#P10) | none | [r10](#P10) |
| [S14](#S14) | none | none | none | accepted | none | none | none | none | none | none |
| [S15](#S15) | [s11](#S11) | none | [s21](#S21) | none | none | [s8](#S8) | none | [s9](#S9) | none | [s10](#S10) |
| [S16](#S16) | [s11](#S11) | none | none | [r1](#P1) | none | [s8](#S8) | none | [s9](#S9) | none | [s10](#S10) |
| [S17](#S17) | [s11](#S11) | none | [r3](#P3) | [r3](#P3) | none | [s8](#S8) | none | [s9](#S9) | none | [s10](#S10) |
| [S18](#S18) | [s11](#S11) | none | [r4](#P4) | [r4](#P4) | none | [s8](#S8) | none | [s9](#S9) | none | [s10](#S10) |
| [S19](#S19) | [s11](#S11) | none | [r5](#P5) | [r5](#P5) | none | [r5](#P5) | none | [r5](#P5) | none | [s10](#S10) |
| [S20](#S20) | [s11](#S11) | none | [r6](#P6) | [r6](#P6) | none | [r6](#P6) | none | [r6](#P6) | none | [s10](#S10) |
| [S21](#S21) | [r8](#P8) | none | [r8](#P8) | [r8](#P8) | none | [r8](#P8) | none | [r8](#P8) | none | [r8](#P8) |
## Goto Table

| State/Nonterminates | S'  | statement  | expr |
| ---  | ---  | ---  | --- |
| [S0](#S0) | none | [s5](#S5) | [s3](#S3) |
| [S1](#S1) | none | none | none |
| [S2](#S2) | none | none | [s0](#S0) |
| [S3](#S3) | none | none | [s0](#S0) |
| [S4](#S4) | none | none | [s12](#S12) |
| [S5](#S5) | none | [s0](#S0) | none |
| [S6](#S6) | none | none | [s15](#S15) |
| [S7](#S7) | none | none | [s16](#S16) |
| [S8](#S8) | none | none | [s17](#S17) |
| [S9](#S9) | none | none | [s18](#S18) |
| [S10](#S10) | none | none | [s19](#S19) |
| [S11](#S11) | none | none | [s20](#S20) |
| [S12](#S12) | none | none | [s0](#S0) |
| [S13](#S13) | none | none | none |
| [S14](#S14) | none | [s0](#S0) | none |
| [S15](#S15) | none | none | [s0](#S0) |
| [S16](#S16) | none | none | [s0](#S0) |
| [S17](#S17) | none | none | [s0](#S0) |
| [S18](#S18) | none | none | [s0](#S0) |
| [S19](#S19) | none | none | [s0](#S0) |
| [S20](#S20) | none | none | [s0](#S0) |
| [S21](#S21) | none | none | [s0](#S0) |

```