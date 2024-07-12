package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const EMPTYTOKEN = "<empty>"
const ENDTOKEN = "$end"

type Parser struct {
	lexer *Lexer
	grammar *grammar
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
	lrItems []*LRItem
	lrNext *LRItem
}

type RuleOps struct {
	Ops string
	RFunc    func(Parser) error
}

type SyntaxRule struct {
	Name   string
	Expand []*RuleOps
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

// This class represents a specific stage of parsing a production rule.  For
// example:
//
//       expr : expr . PLUS term
type LRItem struct {
	name string
	prod *[]string
	number int
	lrIndex int
	lrAfter []*production
	lrBefore string
	lookaheads map[string]int
	len int
	symSet *StrSet
}

func createLRItem(g *grammar,p *production, dotIndex int) *LRItem {
	item := &LRItem{
		name: p.name,
		number: p.id,
		lrIndex: dotIndex,
		lookaheads: make(map[string]int),
		symSet: p.symSet,
		len: 0,
	}

	item.prod = insertStr2Arr(&p.prod, ".", dotIndex)
	item.len = len(*item.prod)

	// compute the list of productions after following
	item.lrAfter = make([]*production, 0)
	if dotIndex < (len(p.prod) - 1) {
		nextProductions := g.prodNames[(*item.prod)[dotIndex + 1]]
		for _, p := range nextProductions {
			item.lrAfter = append(item.lrAfter, p)
		}
	}

	item.lrBefore = ""

	if dotIndex > 0 {
		item.lrBefore = (*item.prod)[dotIndex - 1]
	} else if dotIndex == 0 {
		// get the last one in prod
		item.lrBefore = (*item.prod)[len(*item.prod) - 1]
	}

	return item
}

func (self *LRItem) String() string {
	s := ""
	if self.len != 0 {
		s = fmt.Sprintf("%s -> %s", self.name, strings.Join(*self.prod, " "))
	} else {
		s = fmt.Sprintf("%s -> %s", self.name, EMPTYTOKEN)

	}

	return s
}

func CreateSyntaxParser(l *Lexer, rules []*SyntaxRule, prec []*Precedence) *Parser {
	parser := &Parser{
		lexer: l,
		grammar: CreateGrammar(l, rules, prec),
	}
	parser.buildLRTables()
	return parser
}

// Build the LR Parsing tables from the grammar
func (self *Parser) buildLRTables() {

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

	// prepare for the establishment of LRTable
	grammar.buildLRItems()
	grammar.buildFirst()
	grammar.buildFollow()

	return grammar
}

// Computes all of the follow sets for every non-terminal symbol.  The
// follow set is the set of all symbols that might follow a given
// non-terminal.  See the Dragon book, 2nd Ed. p. 189.
func (g *grammar) buildFollow() {
	if len(g.follow) > 0 {
		return
	}

	if len(g.first) == 0 {
		g.buildFirst()
	}

	// Add '$end' to the follow list of the start symbol
	for n := range g.nonterminals {
		g.follow[n] = createSet()
	}

	g.follow[g.start].add(ENDTOKEN)

	for {
		didAdd := false
		for _, production := range g.productions[1:] {
			for index, symbol := range production.prod {
				if _, ok := g.nonterminals[symbol]; !ok {
					continue
				}
				prodSlice := production.prod[index+1:]
				first := g.getFirstFromProd(&prodSlice)
				hasEmpty := false

				first.forEach(func(f string) {
					if f == EMPTYTOKEN {
						hasEmpty = true
					} else {
						if !g.follow[symbol].contains(f) {
							g.follow[symbol].add(f)
							didAdd = true
						}
					}
				})

				if hasEmpty || index == len(production.prod)-1 {
					follow := g.follow[production.name]
					follow.forEach(func(f string) {
						if !g.follow[symbol].contains(f) {
							g.follow[symbol].add(f)
							didAdd = true
						}
					})
				}
			}
		}

		if !didAdd {
			break
		}
	}
}

// Compute the value of FIRST1(X) for all symbols
func (g *grammar) buildFirst() {
	if len(g.first) > 0 {
		return
	}

	// terminals
	for t := range g.terminals {
		g.first[t] = createSet()
		g.first[t].add(t)
	}

	g.first[ENDTOKEN] = createSet()
	g.first[ENDTOKEN].add(ENDTOKEN)

	// nonterminals
	for n := range g.nonterminals {
		g.first[n] = createSet()
	}

	changed := false
	for {
		for n := range g.nonterminals {
			for _, p := range g.prodNames[n] {
				if !changed {
					changed = g.setFirstFromProd(n, &p.prod)
				}
			}
		}
		if !changed {
			break
		}
	}
}

func (g *grammar) getFirstFromProd(p *[]string) *StrSet {
	result := createSet()

	for _, x := range *p {
		firsts := g.first[x]
		hasEmpty := false
		firsts.forEach(func(s string) {
			if !result.contains(s) {
				result.add(s)
			}

			// empty case
			if s == EMPTYTOKEN {
				result.add(EMPTYTOKEN)
				hasEmpty = true
			}
		})

		if !hasEmpty {
			break
		}
	}

	return result
}

// Compute the value of FIRST1(p) where p is a tuple of symbols.
func (g *grammar) setFirstFromProd(name string, p *[]string) bool {
	result := createSet()
	changed := false

	for _, x := range *p {
		firsts := g.first[x]
		hasEmpty := false
		firsts.forEach(func(s string) {
			if !result.contains(s) {
				result.add(s)
				changed = true
			}

			// empty case
			if s == EMPTYTOKEN {
				result.add(EMPTYTOKEN)
				hasEmpty = true
			}
		})

		if !hasEmpty {
			break
		}
	}

	if result.size() > 0 {
		g.first[name].addSet(result)
	}

	return changed
}

// build_lritems()
//
// This function walks the list of productions and builds a complete set of the
// LR items.  The LR items are stored in two ways:  First, they are uniquely
// numbered and placed in the list _lritems.  Second, a linked list of LR items
// is built for each production.  For example:
//
//   E -> E PLUS E
//
// Creates the list
//
//  [E -> . E PLUS E, E -> E . PLUS E, E -> E PLUS . E, E -> E PLUS E . ]
func (g *grammar) buildLRItems() {
	var last *production
	for _, p := range g.productions {
		last = p
		i := 0
		lrItems := make([]*LRItem, 0)
		for {
			var item *LRItem
			if i > len(p.prod) {
				item = nil
			} else {
				item = createLRItem(g, p, i)
			}
			last.lrNext = item
			if item == nil {
				break
			}
			lrItems = append(lrItems, item)
			i++
		}
		p.lrItems = lrItems
	}
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
	if  len(rules) == 0 {
		panic("no rules!")
	}

	for _, rule := range rules {
		// valid whether it is terminal type
		if _, ok := g.terminals[rule.Name]; ok {
			panic("duplicate name with tokentype")
		}
		for _, ops := range rule.Expand {

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

		if rOps[len(rOps)-2] != PREC {
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

	if predInfo, ok := g.precedence[precName]; ok {
		return predInfo, nil
	} else {
		return fmt.Sprintf("%d-%d", PRIGHT, 0), nil
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
	for n := range g.nonterminals {
		terminates[n] = false
	}

	for {
		changed := false

		for n, products := range g.prodNames {
			// nonterminals in terminates if any of its productions terminates
			for _, p := range products {
				if terminates[n] {
					break
				}
				pTerminates := true
				for _, s := range p.prod {
					// the symbol s is not terminate, so production p does not terminate
					if isTerminate := terminates[s]; !isTerminate {
						pTerminates = false
						break
					}
				}

				// all productions of nonterminal n terminate, so n terminates
				if pTerminates {
					if !terminates[n] {
						terminates[n] = true
						changed = true
					}
				}
			}
		}

		if !changed {
			break
		}
	}

	infinite := make([]string, 0)
	for s, t := range terminates {
		// consider unused case
		_, isTerminalIn := g.terminals[s]
		_, isNonTerminalIn := g.prodNames[s]

		if !isTerminalIn && !isNonTerminalIn {
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

	for s := range g.nonterminals {
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
		if s == g.start {
			continue
		}
		if n != nil && len(n) == 0 {
			fmt.Printf("unused rule %s !! \n", s)
		}
	}

}

func (g *grammar) unusedTerminals() {
	for s, t := range g.terminals {
		if t != nil && len(t) == 0 {
			fmt.Printf("unused terminal %s !! \n", s)
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
	result += "\n"

	result += "First:\n"
	for t, p := range g.first {
		result += fmt.Sprintf("%s: %s\n", t, p.string())
	}

	return result
}


func expStr2Arr(s string) []string {
  // write regexp to get word from string and convert them to array
  // precedence case must be considered
  reg := regexp.MustCompile(`\w+|%prec`)
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
		lrItems: make([]*LRItem, 0),
		lrNext: nil,
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

func insertStr2Arr(arr *[]string, s string, index int) *[]string {
	result := make([]string, len(*arr)+1)
	copy(result, (*arr)[:index])
	result[index] = s
	copy(result[index+1:], (*arr)[index:])
	return &result
}