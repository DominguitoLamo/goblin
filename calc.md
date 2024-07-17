# Lexer

- PLUS : \+ 
- MINUS : \- 
- MULTIPLY : \* 
- ASSIGN : = 
- NAME : [a-zA-Z_][a-zA-Z0-9_]* 
- DIVIDE : / 
- LPAREN : \( 
- RPAREN : \) 
- NUMBER : [0-9]+ 

# Grammar

## Terminates

- $end : {0, } 
- NUMBER : {9, } 
- RPAREN : {8, } 
- NAME : {10, 1, } 
- PLUS : {3, } 
- MULTIPLY : {5, } 
- ASSIGN : {1, } 
- DIVIDE : {6, } 
- LPAREN : {8, } 
- MINUS : {4, 7, } 

## Non-Terminates

- S' : {} 
- statement : {0, } 
- expr : {8, 1, 2, 3, 4, 5, 6, 7, } 

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

- expr -> LPAREN . expr RPAREN 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S2></a>S2

- S' -> statement . $end 

# <a id=S3></a>S3

- statement -> NAME . ASSIGN expr 
- expr -> NAME . 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, $end, }


# <a id=S4></a>S4

- statement -> expr . 

    lookahead: {$end, }

- expr -> expr . PLUS expr 
- expr -> expr . MINUS expr 
- expr -> expr . MULTIPLY expr 
- expr -> expr . DIVIDE expr 

# <a id=S5></a>S5

- expr -> MINUS . expr 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S6></a>S6

- expr -> NUMBER . 

    lookahead: {DIVIDE, RPAREN, $end, PLUS, MINUS, MULTIPLY, }


# <a id=S7></a>S7

- expr -> NAME . 

    lookahead: {MINUS, MULTIPLY, DIVIDE, RPAREN, $end, PLUS, }


# <a id=S8></a>S8

- expr -> LPAREN expr . RPAREN 
- expr -> expr . PLUS expr 
- expr -> expr . MINUS expr 
- expr -> expr . MULTIPLY expr 
- expr -> expr . DIVIDE expr 

# <a id=S9></a>S9

- S' -> statement $end . 

# <a id=S10></a>S10

- statement -> NAME ASSIGN . expr 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S11></a>S11

- expr -> expr PLUS . expr 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S12></a>S12

- expr -> expr MINUS . expr 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S13></a>S13

- expr -> expr MULTIPLY . expr 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S14></a>S14

- expr -> expr DIVIDE . expr 
- expr -> . expr PLUS expr 
- expr -> . expr MINUS expr 
- expr -> . expr MULTIPLY expr 
- expr -> . expr DIVIDE expr 
- expr -> . MINUS expr 
- expr -> . LPAREN expr RPAREN 
- expr -> . NUMBER 
- expr -> . NAME 

# <a id=S15></a>S15

- expr -> MINUS expr . 

    lookahead: {RPAREN, $end, PLUS, MINUS, MULTIPLY, DIVIDE, }

- expr -> expr . PLUS expr 
- expr -> expr . MINUS expr 
- expr -> expr . MULTIPLY expr 
- expr -> expr . DIVIDE expr 

# <a id=S16></a>S16

- expr -> LPAREN expr RPAREN . 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, RPAREN, $end, }


# <a id=S17></a>S17

- statement -> NAME ASSIGN expr . 

    lookahead: {$end, }

- expr -> expr . PLUS expr 
- expr -> expr . MINUS expr 
- expr -> expr . MULTIPLY expr 
- expr -> expr . DIVIDE expr 

# <a id=S18></a>S18

- expr -> expr PLUS expr . 

    lookahead: {MINUS, MULTIPLY, DIVIDE, RPAREN, $end, PLUS, }

- expr -> expr . PLUS expr 

    lookahead: {DIVIDE, RPAREN, $end, PLUS, MINUS, MULTIPLY, }

- expr -> expr . MINUS expr 

    lookahead: {RPAREN, $end, PLUS, MINUS, MULTIPLY, DIVIDE, }

- expr -> expr . MULTIPLY expr 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, RPAREN, $end, }

- expr -> expr . DIVIDE expr 

    lookahead: {MULTIPLY, DIVIDE, RPAREN, $end, PLUS, MINUS, }


# <a id=S19></a>S19

- expr -> expr MINUS expr . 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, RPAREN, $end, }

- expr -> expr . PLUS expr 

    lookahead: {RPAREN, $end, PLUS, MINUS, MULTIPLY, DIVIDE, }

- expr -> expr . MINUS expr 

    lookahead: {MINUS, MULTIPLY, DIVIDE, RPAREN, $end, PLUS, }

- expr -> expr . MULTIPLY expr 

    lookahead: {$end, PLUS, MINUS, MULTIPLY, DIVIDE, RPAREN, }

- expr -> expr . DIVIDE expr 

    lookahead: {$end, PLUS, MINUS, MULTIPLY, DIVIDE, RPAREN, }


# <a id=S20></a>S20

- expr -> expr MULTIPLY expr . 

    lookahead: {MULTIPLY, DIVIDE, RPAREN, $end, PLUS, MINUS, }

- expr -> expr . PLUS expr 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, RPAREN, $end, }

- expr -> expr . MINUS expr 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, RPAREN, $end, }

- expr -> expr . MULTIPLY expr 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, RPAREN, $end, }

- expr -> expr . DIVIDE expr 

    lookahead: {PLUS, MINUS, MULTIPLY, DIVIDE, RPAREN, $end, }


# <a id=S21></a>S21

- expr -> expr DIVIDE expr . 

    lookahead: {DIVIDE, RPAREN, $end, PLUS, MINUS, MULTIPLY, }

- expr -> expr . PLUS expr 

    lookahead: {MULTIPLY, DIVIDE, RPAREN, $end, PLUS, MINUS, }

- expr -> expr . MINUS expr 

    lookahead: {MULTIPLY, DIVIDE, RPAREN, $end, PLUS, MINUS, }

- expr -> expr . MULTIPLY expr 

    lookahead: {MINUS, MULTIPLY, DIVIDE, RPAREN, $end, PLUS, }

- expr -> expr . DIVIDE expr 

    lookahead: {$end, PLUS, MINUS, MULTIPLY, DIVIDE, RPAREN, }


## Action Table

| State/Terminates | PLUS  | MULTIPLY  | ASSIGN  | $end  | NUMBER  | RPAREN  | NAME  | DIVIDE  | LPAREN  | MINUS |
| ---  | ---  | ---  | ---  | ---  | ---  | ---  | ---  | ---  | ---  | --- |
| [S0](#S0) | hello | none | [s6](#S6) | none | [s3](#S3) | none | none | none | [s1](#S1) | [s5](#S5) |
| [S1](#S1) | [s6](#S6) | none | [s7](#S7) | none | none | none | none | none | [s1](#S1) | [s5](#S5) |
| [S2](#S2) | none | none | none | none | none | none | none | [s9](#S9) | none | none |
| [S3](#S3) | [r10](#P10) | [r10](#P10) | [s10](#S10) | [r10](#P10) | none | none | none | [r10](#P10) | none | [r10](#P10) |
| [S4](#S4) | none | [r2](#P2) | none | none | none | [s11](#S11) | [s13](#S13) | [s14](#S14) | none | [s12](#S12) |
| [S5](#S5) | none | [s6](#S6) | none | [s7](#S7) | none | none | none | none | [s1](#S1) | [s5](#S5) |
| [S6](#S6) | none | [r9](#P9) | [r9](#P9) | [r9](#P9) | none | [r9](#P9) | [r9](#P9) | none | [r9](#P9) | none |
| [S7](#S7) | none | [r10](#P10) | none | [r10](#P10) | [r10](#P10) | none | [r10](#P10) | [r10](#P10) | none | [r10](#P10) |
| [S8](#S8) | [s12](#S12) | [s14](#S14) | none | none | [s11](#S11) | [s13](#S13) | none | none | none | [s16](#S16) |
| [S9](#S9) | none | none | none | none | none | none | none | none | none |
| [S10](#S10) | [s5](#S5) | none | [s1](#S1) | [s7](#S7) | none | none | none | none | [s6](#S6) | none |
| [S11](#S11) | [s1](#S1) | [s5](#S5) | none | none | [s7](#S7) | none | none | none | none | [s6](#S6) |
| [S12](#S12) | none | none | [s6](#S6) | none | [s7](#S7) | none | none | none | [s1](#S1) | [s5](#S5) |
| [S13](#S13) | none | [s7](#S7) | none | none | none | none | [s6](#S6) | [s1](#S1) | [s5](#S5) | none |
| [S14](#S14) | none | [s1](#S1) | [s5](#S5) | none | none | none | [s6](#S6) | none | [s7](#S7) | none |
| [S15](#S15) | [r7](#P7) | none | [r7](#P7) | none | [r7](#P7) | none | [r7](#P7) | [r7](#P7) | none | [r7](#P7) |
| [S16](#S16) | [r8](#P8) | none | [r8](#P8) | [r8](#P8) | none | [r8](#P8) | none | [r8](#P8) | none | [r8](#P8) |
| [S17](#S17) | [s14](#S14) | none | [s12](#S12) | none | none | none | [s11](#S11) | [s13](#S13) | none | [r1](#P1) |
| [S18](#S18) | [s13](#S13) | none | [r3](#P3) | none | [r3](#P3) | none | [s11](#S11) | [s14](#S14) | none | [s12](#S12) |
| [S19](#S19) | none | [s11](#S11) | [s13](#S13) | none | [r4](#P4) | none | [r4](#P4) | [s12](#S12) | [s14](#S14) | none |
| [S20](#S20) | [s14](#S14) | none | [r5](#P5) | [r5](#P5) | [s13](#S13) | none | [r5](#P5) | none | [r5](#P5) | none |
| [S21](#S21) | [s14](#S14) | none | [r6](#P6) | none | [r6](#P6) | none | [r6](#P6) | [s13](#S13) | none | [r6](#P6) |
## Goto Table

| State/Nonterminates | expr  | S'  | statement |
| ---  | ---  | ---  | --- |
| [S0](#S0) | [s4](#S4) | none | [s2](#S2) |
| [S1](#S1) | none | none | [s8](#S8) |
| [S2](#S2) | none | [s0](#S0) | none |
| [S3](#S3) | none | [s0](#S0) | none |
| [S4](#S4) | [s0](#S0) | none | none |
| [S5](#S5) | none | none | [s15](#S15) |
| [S6](#S6) | none | none | none |
| [S7](#S7) | none | none | none |
| [S8](#S8) | none | none | [s0](#S0) |
| [S9](#S9) | none | [s0](#S0) | none |
| [S10](#S10) | none | [s17](#S17) | none |
| [S11](#S11) | none | none | [s18](#S18) |
| [S12](#S12) | none | none | [s19](#S19) |
| [S13](#S13) | none | none | [s20](#S20) |
| [S14](#S14) | none | none | [s21](#S21) |
| [S15](#S15) | none | none | [s0](#S0) |
| [S16](#S16) | none | none | [s0](#S0) |
| [S17](#S17) | none | [s0](#S0) | none |
| [S18](#S18) | none | none | [s0](#S0) |
| [S19](#S19) | none | none | [s0](#S0) |
| [S20](#S20) | none | none | [s0](#S0) |
| [S21](#S21) | none | none | [s0](#S0) |
