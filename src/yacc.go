package main

import (
	"fmt"
	"strings"
)

type Parser interface {
}

type production struct{}

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
	productions []*production
	// A dictionary mapping the names of nonterminals to a list of all
	// productions of that nonterminal.
	prodNames    map[string][]*production
	prodMap      map[string]int
	terminals    map[string][]int
	nonterminals map[string][]int
	first        map[string]*StrSet
	follow       map[string]*StrSet
	precedence   map[string]string // Tokentype:acc-level
	usedPrecedence *StrSet
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
		productions:  make([]*production, 0),
		prodNames:    make(map[string][]*production),
		prodMap:      make(map[string]int),
		terminals:    make(map[string][]int),
		nonterminals: make(map[string][]int),
		first:        make(map[string]*StrSet),
		follow:       make(map[string]*StrSet),
		precedence:   make(map[string]string), // Tokentype:acc-level
		usedPrecedence: createSet(),
	}

	// identify keywords in lexer
	for key := range l.rules {
		if isIn, _, keywords := isRedefine(key); isIn {
			grammar.terminals[keywords] = []int{}
			continue
		}
		grammar.terminals[key] = []int{}
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
	precInfo, opsArr := g.getPrecedence(name, rOps)
	var ops []string
	if opsArr != nil {
		ops = opsArr
	} else {
		ops = rOps
	}

	// see if the rule is already defined
	ruleId := fmt.Sprintf("%s->%s", name, strings.Join(ops, " "))
	if _, ok := g.prodMap[ruleId]; ok {
		panic(fmt.Sprintf("duplicate production %s", ruleId))
	}

	// create a new production instance
	pnumber := len(g.productions)
	if _,ok := g.nonterminals[name]; !ok {
		g.nonterminals[name] = []int{}
	}

	// add the production number to Terminals and NonTerminals
	for _, item := range ops {
		if _, ok := g.terminals[item]; ok {
			g.terminals[item] = append(g.terminals[item], pnumber)
		} else {
			if _,ok := g.nonterminals[item]; !ok {
				g.nonterminals[item] = []int{}
			}
			g.nonterminals[item] = append(g.nonterminals[item], pnumber)
		}
	}

	// create a production and add it to the list of productions
	p := createProduction(pnumber, name, ops, precInfo, rFunc)
	g.productions = append(g.productions, p)
	g.prodMap[ruleId] = pnumber


	// add the production to the list of productions for this nonterminal
	if _, ok := g.prodNames[name]; !ok {
		g.prodNames[name] = []*production{}
	}
	g.prodNames[name] = append(g.prodNames[name], p)
}

func (g *grammar) getPrecedence(name string, rOps []string) (string, []string) {
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

		precName := rOps[len(rOps)-1]
		precInfo, isIn := g.precedence[precName]
		if !isIn {
			panic(fmt.Sprintf("Nothing known about the precedence of %s", precName))
		}
		g.usedPrecedence.add(precName)
		return precInfo, rOps[:len(rOps)-2]
	}

	precName := g.rightMostTerminal(rOps)

	if precName == "" {
		return fmt.Sprintf("%s-%s", PRIGHT, 0), nil
	} else {
		return g.precedence[precName], nil
	}
}

func (g *grammar) rightMostTerminal(ops []string) string {
	for i := len(ops) - 1; i >= 0; i-- {
		if _, ok := g.terminals[ops[i]]; ok {
			return ops[i]
		}
	}
	return ""
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

func createProduction(pnumber int, name string, ops []string, precInfo string, rFunc func( Parser) error) {
	panic("unimplemented")
}