package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const ENDTOKEN = "$end"

type Parser interface {
}

type production struct{
	id int
	name string
	prod []string
	prodSize int
	symSet *StrSet
	precDirect int
	precLevel int
	pFunc func(Parser) error
}

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

func CreateGrammar(l *Lexer, r []*SyntaxRule, p []*Precedence) *grammar {
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
	grammar.start = r[0].Name

	// check unused, undefined, unreachable, cycles
	grammar.checkGrammar()

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
			panic(fmt.Sprintf("Syntax error in %s. Nothing follows %%prec", name))
		}

		if rOps[len(rOps)-2] == PREC {
			panic(fmt.Sprintf("Syntax error in %s. %%prec can only appear at the end of a grammar rule", name))
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
		return fmt.Sprintf("%d-%d", PRIGHT, 0), nil
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

func (g *grammar) checkGrammar() {
	g.undefinedSymbols()
	g.unusedTerminals()
	g.unusedRules()
	g.unreachableRules()
	g.cyclicRules()
}


func (g *grammar) cyclicRules() {
	terminates := make(map[string]bool)

	// terminals
	for t := range g.terminals {
		terminates[t] = true
	}
	terminates[ENDTOKEN] = true

	// nonterminals
	for n, _ := range g.nonterminals {
		terminates[n] = false
	}

	for {
		changed := false
		for n, products := range g.prodNames {
			pTerminates := true
			// nonterminals in terminates if any of its productions terminates
			for _, p := range products {
				for _, s := range p.prod {
					// the symbol s is not terminate, so production p does not terminate
					if isTerminate := terminates[s]; !isTerminate {
						pTerminates = false
						break
					}
				}
			}

			// all productions of nonterminal n terminate, so n terminates
			if pTerminates {
				if !terminates[n] {
					terminates[n] = true
					changed = true
				}
				break
			}
		}

		if !changed {
			break
		}
	}

	infinite := make([]string, 0)
	for s, t := range terminates {
		// consider unused case
		// unused terminal
		if _, ok := g.terminals[s]; !ok {
			continue
		}

		// unused rules
		if _, ok := g.prodNames[s]; !ok {
			continue
		}

		if !t {
			infinite = append(infinite, s)
		}
	}

	if len(infinite) > 0 {
		panic(fmt.Sprintf("cyclic rule(s) made by %s", strings.Join(infinite, ",")))
	}
}

func (g *grammar) unreachableRules() {
	reachable := createSet()
	g.makeReachable(g.start, reachable)

	for s, _ := range g.nonterminals {
		if !reachable.contains(s) {
			fmt.Printf("unreachable rule %s!! \n", s)
		}
	}
}

func (g *grammar) makeReachable(s string, reachable *StrSet) {
	if reachable.contains(s) {
		return
	}
	reachable.add(s)
	for _, p := range g.prodNames[s] {
		for _, item := range p.prod {
			g.makeReachable(item, reachable)
		}
	}
}

func (g *grammar) unusedRules() {
	for s, n := range g.nonterminals {
		if n != nil && len(n) == 0 {
			fmt.Printf("unused rule %s!! \n", s)
		}
	}

}

func (g *grammar) unusedTerminals() {
	for s, t := range g.terminals {
		if t != nil && len(t) == 0 {
			fmt.Printf("unused terminal %s!! \n", s)
		}
	}
}

func (g *grammar) undefinedSymbols() {
	for _, p := range g.productions {
		for _, item := range p.prod {
			if _, ok := g.terminals[item]; !ok {
				if _, ok := g.nonterminals[item]; !ok {
					fmt.Printf("undefined symbol %s in %s \n", item, p.name)
				}
			}
		}
	}
}

func (g *grammar) string() string {
	result := ""

	result += "Grammar:\n"
	result += "\n"

	result += "Terminals:\n"

	for t := range g.terminals {
		result += fmt.Sprintf("%s\n", t)
	}
	result += "\n"

	result += "Nonterminals:\n"
	for n := range g.nonterminals {
		result += fmt.Sprintf("%s\n", n)
	}
	result += "\n"

	result += fmt.Sprintf("start:%s \n", g.start)
	result += "\n"

	result += "Productions:\n"
	for _, p := range g.productions {
		result += fmt.Sprintf("%s -> %s\n", p.name, strings.Join(p.prod, " "))
	}

	return result
}


func expStr2Arr(s string) []string {
  // write regexp to get word from string and convert them to array
  reg := regexp.MustCompile(`\w+`)
  return reg.FindAllString(s, -1)
}

func createProduction(pnumber int, name string, ops []string, precInfo string, pfunc func(Parser) error) *production {
	p := &production{
		id: pnumber,
		name: name,
		prod: ops,
		prodSize: len(ops),
		// get the unique symbols in production
		symSet: createSet(),
		precDirect: PRIGHT,
		precLevel: 0,
		pFunc: pfunc,
	}

	if precInfo != "" {
		precArr := strings.Split(precInfo, "-")
		if len(precArr) != 2 {
			panic("invalid precedence info")
		}
	
		if num, err := strconv.Atoi(precArr[0]); err != nil {
			panic("invalid precedence info")
		} else {
			p.precDirect = num
		}
	
		if num, err := strconv.Atoi(precArr[1]); err != nil {
			panic("invalid precedence info")
		} else {
			p.precLevel = num
		}
	}

	for _, item := range ops {
		p.symSet.add(item)
	}

	return p
}