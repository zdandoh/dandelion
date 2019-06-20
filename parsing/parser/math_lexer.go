// Code generated from Math.g4 by ANTLR 4.7.2. DO NOT EDIT.

package parser

import (
	"fmt"
	"unicode"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = unicode.IsLetter

var serializedLexerAtn = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 2, 11, 50, 8,
	1, 4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7, 9,
	7, 4, 8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 3, 2, 3, 2, 3, 3, 3, 3, 3, 4,
	3, 4, 3, 5, 3, 5, 3, 6, 3, 6, 3, 7, 3, 7, 3, 8, 6, 8, 35, 10, 8, 13, 8,
	14, 8, 36, 3, 9, 5, 9, 40, 10, 9, 3, 9, 3, 9, 3, 10, 6, 10, 45, 10, 10,
	13, 10, 14, 10, 46, 3, 10, 3, 10, 2, 2, 11, 3, 3, 5, 4, 7, 5, 9, 6, 11,
	7, 13, 8, 15, 9, 17, 10, 19, 11, 3, 2, 4, 3, 2, 50, 59, 4, 2, 11, 11, 34,
	34, 2, 52, 2, 3, 3, 2, 2, 2, 2, 5, 3, 2, 2, 2, 2, 7, 3, 2, 2, 2, 2, 9,
	3, 2, 2, 2, 2, 11, 3, 2, 2, 2, 2, 13, 3, 2, 2, 2, 2, 15, 3, 2, 2, 2, 2,
	17, 3, 2, 2, 2, 2, 19, 3, 2, 2, 2, 3, 21, 3, 2, 2, 2, 5, 23, 3, 2, 2, 2,
	7, 25, 3, 2, 2, 2, 9, 27, 3, 2, 2, 2, 11, 29, 3, 2, 2, 2, 13, 31, 3, 2,
	2, 2, 15, 34, 3, 2, 2, 2, 17, 39, 3, 2, 2, 2, 19, 44, 3, 2, 2, 2, 21, 22,
	7, 44, 2, 2, 22, 4, 3, 2, 2, 2, 23, 24, 7, 49, 2, 2, 24, 6, 3, 2, 2, 2,
	25, 26, 7, 45, 2, 2, 26, 8, 3, 2, 2, 2, 27, 28, 7, 47, 2, 2, 28, 10, 3,
	2, 2, 2, 29, 30, 7, 42, 2, 2, 30, 12, 3, 2, 2, 2, 31, 32, 7, 43, 2, 2,
	32, 14, 3, 2, 2, 2, 33, 35, 9, 2, 2, 2, 34, 33, 3, 2, 2, 2, 35, 36, 3,
	2, 2, 2, 36, 34, 3, 2, 2, 2, 36, 37, 3, 2, 2, 2, 37, 16, 3, 2, 2, 2, 38,
	40, 7, 15, 2, 2, 39, 38, 3, 2, 2, 2, 39, 40, 3, 2, 2, 2, 40, 41, 3, 2,
	2, 2, 41, 42, 7, 12, 2, 2, 42, 18, 3, 2, 2, 2, 43, 45, 9, 3, 2, 2, 44,
	43, 3, 2, 2, 2, 45, 46, 3, 2, 2, 2, 46, 44, 3, 2, 2, 2, 46, 47, 3, 2, 2,
	2, 47, 48, 3, 2, 2, 2, 48, 49, 8, 10, 2, 2, 49, 20, 3, 2, 2, 2, 6, 2, 36,
	39, 46, 3, 8, 2, 2,
}

var lexerDeserializer = antlr.NewATNDeserializer(nil)
var lexerAtn = lexerDeserializer.DeserializeFromUInt16(serializedLexerAtn)

var lexerChannelNames = []string{
	"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
}

var lexerModeNames = []string{
	"DEFAULT_MODE",
}

var lexerLiteralNames = []string{
	"", "'*'", "'/'", "'+'", "'-'", "'('", "')'",
}

var lexerSymbolicNames = []string{
	"", "", "", "", "", "", "", "INT", "NEWLINE", "WS",
}

var lexerRuleNames = []string{
	"T__0", "T__1", "T__2", "T__3", "T__4", "T__5", "INT", "NEWLINE", "WS",
}

type MathLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var lexerDecisionToDFA = make([]*antlr.DFA, len(lexerAtn.DecisionToState))

func init() {
	for index, ds := range lexerAtn.DecisionToState {
		lexerDecisionToDFA[index] = antlr.NewDFA(ds, index)
	}
}

func NewMathLexer(input antlr.CharStream) *MathLexer {

	l := new(MathLexer)

	l.BaseLexer = antlr.NewBaseLexer(input)
	l.Interpreter = antlr.NewLexerATNSimulator(l, lexerAtn, lexerDecisionToDFA, antlr.NewPredictionContextCache())

	l.channelNames = lexerChannelNames
	l.modeNames = lexerModeNames
	l.RuleNames = lexerRuleNames
	l.LiteralNames = lexerLiteralNames
	l.SymbolicNames = lexerSymbolicNames
	l.GrammarFileName = "Math.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// MathLexer tokens.
const (
	MathLexerT__0    = 1
	MathLexerT__1    = 2
	MathLexerT__2    = 3
	MathLexerT__3    = 4
	MathLexerT__4    = 5
	MathLexerT__5    = 6
	MathLexerINT     = 7
	MathLexerNEWLINE = 8
	MathLexerWS      = 9
)