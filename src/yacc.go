package main

import "fmt"

type Parser interface {
}

type Production interface{}

type RuleOps struct {
	Ops string
	RFunc    func(Parser) error
}

type SyntaxRule struct {
	Name   string
	Expend []*RuleOps
}

type grammar struct {
	// A list of all of the productions.  The first
	// entry is always reserved for the purpose of
	// building an augmented grammar
	productions []any
	// A dictionary mapping the names of nonterminals to a list of all
	// productions of that nonterminal.
	prodNames    map[string]any
	prodMap      map[string]any
	terminals    map[string][]any
	nonterminals map[string][]any
	first        map[string]*StrSet
	follow       map[string]*StrSet
	precedence   map[string]string // Tokentype:acc-level
	start        string
}

const (
	PLEFT = iota
	PRIGHT
)

type Precedence struct {
	Direct    int
	TokenType []string
	Level     int
}

func createGrammar(l *Lexer, r []*SyntaxRule, p []*Precedence) *grammar {
	grammar := &grammar{
		productions:  make([]any, 1),
		prodNames:    make(map[string]any),
		prodMap:      make(map[string]any),
		terminals:    make(map[string][]any),
		nonterminals: make(map[string][]any),
		first:        make(map[string]*StrSet),
		follow:       make(map[string]*StrSet),
		precedence:   make(map[string]string), // Tokentype:acc-level
	}

	for key := range l.rules {
		if isIn, _, keywords := isRedefine(key); isIn {
			grammar.terminals[keywords] = []any{}
			continue
		}
		grammar.terminals[key] = []any{}
	}

	grammar.setPrecedence(p)
	grammar.setRules(r)

	return grammar
}

func (g *grammar) setPrecedence(p []*Precedence) {
	for _, p := range p {
		if p.Direct > 1 {
			panic("precedence direct must be left or right")
		}
		for _, t := range p.TokenType {
			if _, ok := g.precedence[t]; ok {
				panic(fmt.Sprintf("precedence conflict for token type %s", t))
			}

			g.precedence[t] = fmt.Sprintf("%d-%d", p.Direct, p.Level)
		}
	}
}

func (g *grammar) setRules(rules []*SyntaxRule) {
	for _, rule := range rules {
		// valid whether it is terminal type
		if _, ok := g.terminals[rule.Name]; ok {
			panic("duplicate name with tokentype")
		}
		for _, ops := range rule.Expend {
			rOps := expStr2Arr(ops.Ops)
			g.addProduction(rule.Name, rOps, ops.RFunc)
		}
	}
}

func (g *grammar) addProduction(name string, rOps []string, rFunc func(Parser) error) {
	// Determine the precedence level
	const PREC = "%prec"
	isPrecExist := false
	for _, item := range rOps {
		if item == PREC {
			isPrecExist = true
			break
		}
	}

	if isPrecExist {
		if rOps[len(rOps)-1] == PREC {
			panic(fmt.Sprintf("Syntax error in %s. Nothing follows %prec", name))
		}

		if rOps[len(rOps)-2] == PREC {
			panic(fmt.Sprintf("Syntax error in %s. %prec can only appear at the end of a grammar rule", name))
		}

		
	}
}

func expStr2Arr(s string) []string {
	arr := []string{}
	start := 0
	end := 0

	for end < len(s) {
		if s[end] != ' ' && s[end] != '\t' {
			end++
			continue
		} else {
			if s[start] != ' ' {
				arr = append(arr, s[start:end])
			}
			
			start = end + 1
			end = start
		}
	}
	return arr
}