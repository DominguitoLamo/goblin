package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const EMPTYTOKEN = "<empty>"
const ENDTOKEN = "$end"

type Parser struct {
	lexer *Lexer
	grammar *grammar
	table *lrTable
}

// This struct implements the LR table generation algorithm.
type lrTable struct {
	grammar *grammar
	addCount int // Internal counter used to detect cycles
	closures [][]*LRItem
	closureMap map[int]int // map hash of lr closure to index of lr closure
	lrAction map[int]map[string]string
	lrGoto map[int]map[string]int
	lrProductions []*production
	actionProductions map[int]map[string]*LRItem
	// Cache of computed gotos
	lrGotoCache map[string][]*LRItem
	symbolGotoCache map[string]*symbolCache
}

type symbolCache struct {
	transfer map[int]int
	end []*LRItem
}

type looked struct {
	state int
	item *LRItem
}

type production struct{
	id int
	name string
	prod []string
	prodSize int
	symSet *StrSet
	precLevel int
	pFunc func(Parser) error
	lrItems []*LRItem
	lrNext *LRItem
	lr0Added int
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
	precedence   map[string]int // Tokentype: level
	usedPrecedence *StrSet
	start        string
}

type Precedence struct {
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
	lrNext *LRItem
	lrAfter []*production
	lrBefore string
	lookaheads map[int]*StrSet
	len int
	symSet *StrSet
}

func CreateParser(lrules map[string]string, ignore []string, srules []*SyntaxRule, precedence []*Precedence) *Parser {
	lexer := CreateLexer(lrules, ignore)
	grammar := CreateGrammar(lexer, srules, precedence)
	table := createLRTable(grammar)
	return &Parser{
		lexer: lexer,
		grammar: grammar,
		table: table,
	}
}

func (p *Parser) Tokenize(s string) ([]*Token, error) {
	return p.lexer.Tokenize(s)
}

// write the info about lalr parse to markdown format, and store it
func (p *Parser) WriteMDInfo(name string, path string) {
	result := ""
	result = p.lexMD()
	result += p.grammarMD()
	result += p.lrTableMD()

	// write string to file
	var mdPath string
	if strings.Contains(name, ".md") {
		mdPath = filepath.Join(path, name)
	} else {
		mdPath = filepath.Join(path, name + ".md")
	}

	file, fileErr := os.Create(mdPath)
	if fileErr != nil {
		panic(fileErr)
	}
	defer file.Close()
	_, writeErr := file.WriteString(result)
	if writeErr != nil {
		panic(writeErr)
	}

	fmt.Printf("file %s is written successfully in %s\n", name, path)
}

func (p *Parser) lrTableMD() string {
	result := "# LR Table\n"
	result += "\n"

	// states
	result += "## States\n"
	result += "\n"

	for i, closure := range p.table.closures {
		result += fmt.Sprintf("# <a id=S%d></a>S%d\n", i, i)
		result += "\n"
		for _, item := range closure {
			result += fmt.Sprintf("- %s \n", item.String())
			// lookahead
			if item.lookaheads != nil {
				if heads, ok := item.lookaheads[i]; ok {
					result += "\n    lookahead: "
					result += heads.string() + "\n"
					result += "\n"
				}
			}
		}
		result += "\n"
	}
		
	// action table
	result += "## Action Table\n"
	result += "\n"

	map2Arr := func(m map[string][]int) []string {
		a := make([]string, 0)
		for key := range m {
			a = append(a, key)
		}

		return a
	}

	// table header
	tArr := map2Arr(p.grammar.terminals)
	result += "| State/Terminates"
	for _, term := range tArr {
		result += fmt.Sprintf(" | %s ", term)
	}
	result += "|\n"
	result += "| --- "
	for _, _ = range tArr {
		result += " | --- "
	}
	result += "|\n"

	// table body
	for i := range p.table.closures {
		result += fmt.Sprintf("| [S%d](#S%d) ", i, i)
		for _, term := range tArr {
			action, ok := p.table.lrAction[i][term]
			if !ok {
				result += "| none "
				continue
			}
			// shift
			if action[0] == 's' {
				id := turnAction2id(action)
				result += fmt.Sprintf("| [%s](#S%d) ", action, id)
			} else if action[0] == 'r' {
				// reduce
				id := turnAction2id(action)
				result += fmt.Sprintf("| [%s](#P%d) ", action, id)
			} else {
				// other case
				result += fmt.Sprintf("| %s ", action)
			}
		}
		result += "|\n"
	}

	// goto table
	result += "## Goto Table\n"
	result += "\n"

	// table header
	nArr := map2Arr(p.grammar.nonterminals)
	result += "| State/Nonterminates"
	for _, non := range nArr {
		result += fmt.Sprintf(" | %s ", non)
	}
	result += "|\n"
	result += "| --- "
	for _, _ = range nArr {
		result += " | --- "
	}
	result += "|\n"

	// table body
	for i := range p.table.closures {
		result += fmt.Sprintf("| [S%d](#S%d) ", i, i)
		for _, non := range nArr {
			lgoto, ok := p.table.lrGoto[i][non]
			if ok {
				result += fmt.Sprintf("| [s%d](#S%d) ", lgoto, lgoto)
			} else {
				result += "| none "
			}
		}
		result += "|\n"
	}

	return result
}

func (p *Parser) grammarMD() string {
	result := "# Grammar\n"
	result += "\n"

	// terminates
	result += "## Terminates\n"
	result += "\n"
	for term, states := range p.grammar.terminals {
		set := createSet()
		for _, state := range states {
			set.add(fmt.Sprintf("%d", state))
		}
		result += fmt.Sprintf("- %s : %s \n", term, set.string())
	}
	result += "\n"

	// nonterminates
	result += "## Non-Terminates\n"
	result += "\n"
	for term, states := range p.grammar.nonterminals {
		set := createSet()
		for _, state := range states {
			set.add(fmt.Sprintf("%d", state))
		}
		result += fmt.Sprintf("- %s : %s \n", term, set.string())
	}
	result += "\n"

	// Precedence
	result += " Precedence\n"
	result += "\n"
	for term, level := range p.grammar.precedence {
		result += fmt.Sprintf("- %s : %d \n", term, level)
	}
	result += "\n"

	// Productions
	result += "## Productions\n"
	result += "\n"
	for _, prod := range p.grammar.productions {
		result += fmt.Sprintf("### <a id=P%d></a>P%d. %s -> %s \n", prod.id, prod.id, prod.name, strings.Join(prod.prod, " "))
	}

	return result
}

func (p *Parser) lexMD() string {
	result := "# Lexer\n"
	result += "\n"

	for key, val := range p.lexer.rules {
		result += fmt.Sprintf("- %s : %s \n", key, val)
	}
	result += "\n"

	return result
}


func createLRTable(g *grammar) *lrTable {
	table := &lrTable {
		grammar: g,
		addCount: 0,
		closureMap: make(map[int]int),
		lrAction: make(map[int]map[string]string),
		lrGoto: make(map[int]map[string]int),
		lrProductions: g.productions,
		actionProductions: make(map[int]map[string]*LRItem),
		lrGotoCache: make(map[string][]*LRItem),
		symbolGotoCache: make(map[string]*symbolCache),
	}

	// Step 1: Construct C = { I0, I1, ... IN}, collection of LR(0) items
	// This determines the number of states
	closures := table.lr0Items()
	table.closures = closures
	table.addLalrLookheads()
	debugPrintClosures(table)

	// Let's build LR Table!
	// build the parser table, state by state
	for cIndex, closure := range closures {
		// loop over each production in I
		stAction := make(map[string]string)
		stActionItem := make(map[string]*LRItem)

		for _, lrItem := range closure {
			// dotIndex to the end of the production. Reduce
			if (lrItem.lrIndex + 1) == lrItem.len {
				// Start symbol. Accept!
				if lrItem.name == "S'" {
					stAction[ENDTOKEN] = "accepted"
					stActionItem[ENDTOKEN] = lrItem
				} else {
					// We are at the end of a production.  Reduce!
					laHeads := lrItem.lookaheads[cIndex]
					laHeads.forEach(func(head string) {
						r, isHead := stAction[head]
						if isHead {
							// shift/ reduce conflict
							if r[0] == 's' {
								// precdence is the key to make decision. shift is favored.
								sLevel := g.precedence[head]
								rLevel := g.productions[lrItem.number].precLevel
								// reduce
								if rLevel > sLevel {
									stAction[head] = fmt.Sprintf("r%d", lrItem.number)
									stActionItem[head] = lrItem
								}
							} else {
								// reduce/reduce conflict. Invoke panic!
								oldl := stActionItem[head]
								panic(fmt.Sprintf("reduce/reduce conflict between %s and %s in state %d",
								 oldl.String(), lrItem.String(), cIndex))
							}
						} else {
							// just reduce
							stAction[head] = fmt.Sprintf("r%d", lrItem.number)
							stActionItem[head] = lrItem
						}
					})
				}
			} else {
				// We are not at the end of a production.  Shift
				i := lrItem.lrIndex
				front := (*lrItem.prod)[i + 1] // get symbol right after "."
				if _, ok := g.terminals[front]; ok {
					sGoto := table.lr0Goto(closure, front)
					var stateId int
					if s, ok := table.closureMap[hashLRItems(sGoto)]; ok {
						stateId = s
					} else {
						panic(fmt.Sprintf("LR0 goto state of %s not found in state %d", lrItem.String(), cIndex))
					}

					if stateId >= 0 {
						// shift state
						if shift, ok := stAction[front]; ok {
							// shift/shift conflict!
							if shift[0] == 's' {
								oldId := turnAction2id(shift)
								if oldId != stateId {
									panic(fmt.Sprintf("shift conflict between states %d and %d", cIndex, oldId))
								}
							} else if shift[0] == 'r' {
								// reduce/shift conflict
								oldl := g.productions[turnAction2id(shift)]
								oldPrec := oldl.precLevel
								prec := g.precedence[front]

								if prec >= oldPrec  {
									stAction[front] = fmt.Sprintf("s%d", stateId)
									stActionItem[front] = lrItem
								}
							}

						} else {
							stAction[front] = fmt.Sprintf("s%d", stateId)
							stActionItem[front] = lrItem
						}
					}
				}
			}
		}

		// construct goto table
		stGoto := make(map[string]int)
		nonTerminals := createSet()
		for _, l := range closure {
			l.symSet.forEach(func(s string){
				if _, ok := g.nonterminals[s]; ok {
					nonTerminals.add(s)
				}
			})
		}

		nonTerminals.forEach(func(nt string) {
			gotoItems := table.lr0Goto(closure, nt)
			gotoId := table.closureMap[hashLRItems(gotoItems)]
			if gotoId >= 0 {
				stGoto[nt] = gotoId
			}
		})

		table.lrAction[cIndex] = stAction
		table.lrGoto[cIndex] = stGoto
		table.actionProductions[cIndex] = stActionItem
	}

	return table
}

func turnAction2id(action string) int {
	sStr := action[1:]
	state, err := strconv.Atoi(sStr)
	if err != nil {
		panic(err)
	}
	return state
}

func (self *lrTable) addLalrLookheads() {
	nullable := self.computeNullableNonterminals()

	trans := self.findNonterminalTransition()

	// get the next terminal in each trans
	readsets := self.computeReadSets(trans)

	lookd, included := self.computeLookbackIncludes(trans, nullable)

	followSets := self.computeFollowSets(trans, readsets, included)

	self.addLookaheads(lookd, followSets)
}

func (self *lrTable) addLookaheads(lookd map[string][]*looked, followSets map[string]*StrSet) {
	for tran, lookb := range lookd {
		for _, l := range lookb {
			state := l.state
			prod := l.item
			if _, ok := prod.lookaheads[state]; !ok {
				prod.lookaheads[state] = createSet()
			}

			follow := followSets[tran]
			follow.forEach(func(f string){
				if !prod.lookaheads[state].contains(f) {
					prod.lookaheads[state].add(f)
				}
			})
		}
	}
}

func (self *lrTable ) computeFollowSets(trans *StrSet, readsets map[string]*StrSet, included map[string]*StrSet) map[string]*StrSet {
	followsets := make(map[string]*StrSet)

	trans.forEach(func(tran string) {
		if _, ok := followsets[tran]; !ok {
			followsets[tran] = createSet()
		}
		followTerminals := readsets[tran]
		followsets[tran].addSet(followTerminals)

		if includedSet, ok := included[tran]; ok {
			includedSet.forEach(func(i string) {
				followsets[tran].addSet(readsets[i])
			})
		}
	})

	return followsets
}
// Determines the lookback and includes relations
//
// LOOKBACK:
//
// This relation is determined by running the LR(0) state machine forward.
// For example, starting with a production "N : . A B C", we run it forward
// to obtain "N : A B C ."   We then build a relationship between this final
// state and the starting state.   These relationships are stored in a dictionary
// lookdict.
//
// INCLUDES:
//
// Computes the INCLUDE() relation (p,A) INCLUDES (p',B).
//
// This relation is used to determine non-terminal transitions that occur
// inside of other non-terminal transition states.   (p,A) INCLUDES (p', B)
// if the following holds:
//
//       B -> LAT, where T -> epsilon and p' -L-> p
//
// L is essentially a prefix (which may be empty), T is a suffix that must be
// able to derive an empty string.  State p' must lead to state p with the string L.
//
func (self *lrTable) computeLookbackIncludes(trans *StrSet, nullable *StrSet) (map[string][]*looked, map[string]*StrSet) {
	lookDict := make(map[string][]*looked)
	includedDict := make(map[string]*StrSet)

	// loop over all transitions and compute lookbacks and includes
	trans.forEach(func(tran string){
		state, nonTerminal := getStateAndNonterminal(tran)
		lookb := make([]*looked, 0)
		included := createSet()

		for _, lrItem := range self.closures[state] {
			if lrItem.name != nonTerminal {
				continue
			}

			// we have a name match.  We now follow the production all the way
            // through the state machine until we get the . on the right hand side
			lrIndex := lrItem.lrIndex
			currentState := state
			
			for lrIndex < lrItem.len - 1 {
				lrIndex++
				token := (*lrItem.prod)[lrIndex]
				newTran := fmt.Sprintf("%d-%s", currentState, token)
				// Check to see if this token and state are a non-terminal transition
				if trans.contains(newTran) {
					i := lrIndex + 1
					isIncluded := true
					for i < lrItem.len {
						t := (*lrItem.prod)[i]
						// if we have a terminal, we can't include it
						if _, ok := self.grammar.terminals[t]; ok {
							isIncluded = false
							break
						}
						// if we have a non-terminal, we need to check if it is nullable
						if !nullable.contains(t) {
							isIncluded = false
							break
						}
						i++
					}
					// if the prods of the productions are all nullable or all non-terminal, we can include it.
					// or the last one prod in the lritem of the current state is the non-terminal indicated, we can include it.
					if isIncluded {
						included.add(newTran)
					}
				}

				// go to next set
				cGoto := self.lr0Goto(self.closures[currentState], token)
				currentState = self.closureMap[hashLRItems(cGoto)]
			}

			// When we get here, currentState is the final state, now we have to locate the production
			for _, csItem := range self.closures[currentState] {
				if csItem.name != lrItem.name {
					continue
				}
				if csItem.len != lrItem.len {
					continue
				}

				i := 0
				isLooked := true
				for i < csItem.lrIndex {
					if (*csItem.prod)[i] != (*lrItem.prod)[i + 1] {
						isLooked = false
						break
					}
					i++
				}
				if isLooked {
					lookb = append(lookb, &looked{
						state: currentState,
						item: csItem,
					})

				}
			}
		}

		lookDict[tran] = lookb
		included.forEach(func(item string) {
			if _, ok := includedDict[item]; !ok {
				includedDict[item] = createSet()
			}
			includedDict[item].add(tran)
		})
	})

	return lookDict, includedDict
}

func (self *lrTable) computeReadSets(trans *StrSet) map[string]*StrSet {
	readset := make(map[string]*StrSet)

	trans.forEach(func(tran string) {
		state, nonTerminal := getStateAndNonterminal(tran)
		readset[tran] = createSet()

		if state == 0 {
			readset[tran].add(ENDTOKEN)
		}

		cGoto := self.lr0Goto(self.closures[state], nonTerminal)
		for _, lrItem := range cGoto {
			if lrItem.lrIndex < (lrItem.len - 1) {
				a := (*lrItem.prod)[lrItem.lrIndex + 1]
				if _, ok := self.grammar.terminals[a]; ok {
					readset[tran].add(a)
				}
			}
		}
	})

	return readset
}

func getStateAndNonterminal(s string) (int, string) {
	arr := strings.Split(s, "-")
	if len(arr) != 2 {
		panic("invalid transition string")
	}
	nonTerminal := arr[1]
	state, _ := strconv.Atoi(arr[0])

	return state, nonTerminal
}

// Creates a dictionary containing all of the non-terminals that might produce an empty production.
func (self *lrTable) computeNullableNonterminals() *StrSet {
	nullable := createSet()
	numNullable := 0

	for {
		for _, p := range self.grammar.productions[1:] {
			if p.prodSize == 0 {
				nullable.add(p.name)
				continue
			}

			isAllNullable := true
			for _, t := range p.prod {
				if !nullable.contains(t) {
					isAllNullable = false
					break
				}
			}
			if isAllNullable {
				nullable.add(p.name)
			}
		}

		// is nullable changed
		if nullable.size() == numNullable {
			break
		}
		numNullable = nullable.size()
	}

	return nullable
}

// Given a set of LR(0) items, this functions finds all of the non-terminal
// transitions.
func (self *lrTable) findNonterminalTransition() *StrSet {
	closures := self.closures
	trans := createSet()

	for cIndex, closure := range closures {
		for _, lrItem := range closure {
			lrIndex := lrItem.lrIndex
			if lrIndex < (lrItem.len - 1) {
				nonTerm := (*lrItem.prod)[lrIndex + 1]
				if _, ok := self.grammar.nonterminals[nonTerm]; ok {
					t := fmt.Sprintf("%d-%s", cIndex, nonTerm)
					if !trans.contains(t) {
						trans.add(t)
					}
				}
			}
		}
	}

	return trans
}

// get all the states of LR(0) closures
func (self *lrTable) lr0Items() [][]*LRItem {
	closures := make([][]*LRItem, 0)
	closures = append(closures, self.lr0Closure(&[]*LRItem{
		self.grammar.productions[0].lrNext,
	}))
	i := 0
	for _, item := range closures {
		self.closureMap[hashLRItems(item)] = i
		i++
	}

	// Loop over the items in C and each grammar symbols
	i = 0
	// len has to be invoked each iteration since the length of the closure is increasing
	for i < len(closures) {
		cItem := closures[i]
		i++

		allSymbols := createSet()
		for _, lrItem := range cItem {
			allSymbols.addSet(lrItem.symSet)
		}

		allSymbols.forEach(func(symbol string){
			cGoto := self.lr0Goto(cItem, symbol)
			if len(cGoto) == 0 || self.closureMap[hashLRItems(cGoto)] != 0 {
				// continue
			} else {
				self.closureMap[hashLRItems(cGoto)] = len(closures)
				closures = append(closures, cGoto)
			}
		})
	}

	return closures
}

func hashLRItem(lr *LRItem) int {
	result := lr.String()
	hash := md5.Sum([]byte(result))

	return int(binary.LittleEndian.Uint32(hash[:]))
}

// compute hash with the concat of 
func hashLRItems(lr []*LRItem) int {
	result := ""
	for _, item := range lr {
		result += item.String()
	}

	hash := md5.Sum([]byte(result))
	
	return int(binary.LittleEndian.Uint32(hash[:]))
}

// Compute the LR(0) closure operation on items, where items is a array of LR(0) items.
func (self *lrTable) lr0Closure(items *[]*LRItem) []*LRItem {
	self.addCount++

	result := make([]*LRItem, 0)
	result = append(result, *items...)

	didAdd := true
	for didAdd {
		didAdd = false
		for _, item := range result {
			for _, after := range item.lrAfter {
				if after.lr0Added == self.addCount {
					continue
				}
				result = append(result, after.lrNext)
				after.lr0Added = self.addCount
				didAdd = true
			}
		}
	}

	return result
}

// Compute the LR(0) goto function goto(lrs,symbol) where I is a set
// of LR(0) items and X is a grammar symbol.   This function is written
// in a way that guarantees uniqueness of the generated goto sets
func (self *lrTable) lr0Goto(lrs []*LRItem, symbol string) []*LRItem {
	// First we look for a previously cached entry
	lrCacheKey := fmt.Sprintf("%d-%s", hashLRItems(lrs), symbol)
	if lGoto, ok := self.lrGotoCache[lrCacheKey]; ok { 
		return lGoto
	}

	// Now we generate the goto set in a way that guarantees uniqueness
	// of the result
	s := self.symbolGotoCache[lrCacheKey]
	if s == nil {
		s = &symbolCache{
			transfer: make(map[int]int),
			end: make([]*LRItem, 0),
		}
		self.symbolGotoCache[lrCacheKey] = s
	}

	sGoto := make([]*LRItem, 0)
	var currentId int
	for _, lrItem := range lrs {
		next := lrItem.lrNext
		// the next in front of the dot is the symbol
		if next != nil && next.lrBefore == symbol {
			nextHash := hashLRItem(next)
			_, isExist := s.transfer[nextHash]
			if !isExist {
				if currentId == 0 {
					currentId = nextHash
					s.transfer[currentId] = currentId
				} else {
					s.transfer[currentId] = nextHash
					currentId = nextHash
				}
				sGoto = append(sGoto, next)
			}
		}
	}

	if len(s.end) == 0 {
		if len(sGoto) > 0 {
			s.end = self.lr0Closure(&sGoto)
		} else {
			s.end = make([]*LRItem, 0)
		}
	}
	self.lrGotoCache[lrCacheKey] = s.end
	return self.lrGotoCache[lrCacheKey]
}

func createLRItem(g *grammar,p *production, dotIndex int) *LRItem {
	item := &LRItem{
		name: p.name,
		number: p.id,
		lrIndex: dotIndex,
		lookaheads: make(map[int]*StrSet),
		symSet: p.symSet,
		len: 0,
	}

	item.prod = insertStr2Arr(&p.prod, ".", dotIndex)
	item.len = len(*item.prod)

	// compute the list of productions after following
	item.lrAfter = make([]*production, 0)
	if dotIndex < (item.len - 1) {
		nextIndex := (*item.prod)[dotIndex + 1]
		nextProductions := g.prodNames[nextIndex]
		item.lrAfter = append(item.lrAfter, nextProductions...)
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
		precedence:   make(map[string]int), // Tokentype:acc-level
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
	grammar.terminals[ENDTOKEN] = []int{}

	grammar.setPrecedence(p)
	grammar.setRules(r)
	grammar.start = "S'"

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

	for {
		changed := false
		for n := range g.nonterminals {
			for _, p := range g.prodNames[n] {
				changed = g.setFirstFromProd(n, &p.prod)
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
	nSet := g.first[name]
	changed := false

	for _, x := range *p {
		firsts := g.first[x]
		hasEmpty := false
		firsts.forEach(func(s string) {
			if !nSet.contains(s) {
				nSet.add(s)
				changed = true
			}

			// empty case
			if s == EMPTYTOKEN {
				nSet.add(EMPTYTOKEN)
				hasEmpty = true
			}
		})

		if !hasEmpty {
			break
		}
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
	for _, p := range g.productions {
		var currentlr *LRItem
		i := 0
		lrItems := make([]*LRItem, 0)
		for {
			var item *LRItem
			if i > len(p.prod) {
				item = nil
			} else {
				item = createLRItem(g, p, i)
			}

			if currentlr == nil {
				p.lrNext = item
				currentlr = item
			} else {
				currentlr.lrNext = item
				currentlr = item
			}

			if item == nil {
				break
			}
			lrItems = append(lrItems, item)
			i++
		}
		p.lrItems = lrItems
	}
}

func (g *grammar) setPrecedence(precs []*Precedence) {
	for _, p := range precs {
		for _, t := range p.TokenType {
			if _, ok := g.precedence[t]; ok {
				panic(fmt.Sprintf("precedence conflict for token type %s", t))
			}

			g.precedence[t] = p.Level
		}
	}
}

func (g *grammar) setRules(rules []*SyntaxRule) {
	if  len(rules) == 0 {
		panic("no rules!")
	}

	// add start rule
	g.addProduction("S'", []string{rules[0].Name, ENDTOKEN}, nil)
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

func (g *grammar) getPrecedence(name string, rOps []string) (int, []string) {
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
		return 0, nil
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
	terminates[g.start] = true

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
	result += "\n"
	
	result += "Follow:\n"
	for t, p := range g.follow {
		result += fmt.Sprintf("%s: %s\n", t, p.string())
	}
	result += "\n"

	return result
}


func expStr2Arr(s string) []string {
  // write regexp to get word from string and convert them to array
  // precedence case must be considered
  reg := regexp.MustCompile(`\w+|%prec`)
  return reg.FindAllString(s, -1)
}

func createProduction(pnumber int, name string, ops []string, precInfo int, pfunc func(Parser) error) *production {
	p := &production{
		id: pnumber,
		name: name,
		prod: ops,
		prodSize: len(ops),
		// get the unique symbols in production
		symSet: createSet(),
		precLevel: 0,
		pFunc: pfunc,
		lrItems: make([]*LRItem, 0),
		lrNext: nil,
	}

	if precInfo != 0 {
		p.precLevel = precInfo
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

func debugPrintClosures(table *lrTable) {
	closures := table.closures

	fmt.Printf("Closures:\n")

	for cIndex, closure := range closures {
		fmt.Printf("Closure %d:\n", cIndex)
		for lIndex, lr := range closure {
			fmt.Printf("%d.%d - %s \n", cIndex, lIndex, lr.String())
			if _, ok := lr.lookaheads[cIndex]; ok {
				fmt.Printf("  lookaheads: %s \n", lr.lookaheads[cIndex].string())
			}
		}
	}
}