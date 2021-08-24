package ast

import (
	"github.com/fagnercarvalho/redis-lsp/token"
	"strings"
)

type TokenList interface {
	Node
	GetTokens() []Node
	AddToken(Node)
	SetTokens([]Node)
	NextTokenIs([]string, int) bool
	PreviousTokenIs([]string, int) bool
}

type Node interface {
	Type() token.Type
	Start() int
	End() int
	LineStart() int
	LineEnd() int
	Line() int
	String() string
}

type Statement struct {
	CurrentLine int
	Tokens      []Node
}

func (s *Statement) AddToken(n Node) {
	s.Tokens = append(s.Tokens, n)
}

func (s *Statement) SetTokens(nodes []Node) {
	s.Tokens = nodes
}

func (s *Statement) GetTokens() []Node {
	return s.Tokens
}

func (s *Statement) NextTokenIs(expected []string, startIndex int) bool {
	if startIndex >= len(s.Tokens) {
		return false
	}

	tokens := s.Tokens[startIndex:]
	for _, v := range tokens {
		if v.Type() == token.Space {
			continue
		}

		return hasExpectedKeyword(expected, v.String())
	}

	return false
}

func (s *Statement) PreviousTokenIs(expected []string, endIndex int) bool {
	tokens := s.Tokens
	for i := len(tokens) - 1; i >= 0; i-- {
		v := tokens[i]

		if v.Type() == token.Space || v.LineEnd() > endIndex {
			continue
		}

		return hasExpectedKeyword(expected, v.String())
	}

	return false
}

func (s Statement) Type() token.Type {
	return token.Statement
}

func (s Statement) Start() int {
	return s.Tokens[0].Start()
}

func (s Statement) End() int {
	return s.Tokens[len(s.Tokens)-1].End()
}

func (s Statement) LineStart() int {
	return s.Tokens[0].LineStart()
}

func (s Statement) LineEnd() int {
	return s.Tokens[len(s.Tokens)-1].LineEnd()
}

func (s Statement) Line() int {
	return s.CurrentLine
}

func (s Statement) String() string {
	var result []string
	for _, t := range s.Tokens {
		result = append(result, t.String())
	}

	return strings.Join(result, "")
}

type MultiKeyword struct {
	CurrentLine int
	Tokens      []Node
}

func (s MultiKeyword) Type() token.Type {
	return token.MultiKeyword
}

func (s MultiKeyword) Start() int {
	return s.Tokens[0].Start()
}

func (s MultiKeyword) End() int {
	return s.Tokens[0].End()
}

func (s MultiKeyword) LineStart() int {
	return s.Tokens[0].LineStart()
}

func (s MultiKeyword) LineEnd() int {
	return s.Tokens[len(s.Tokens)-1].LineEnd()
}

func (s MultiKeyword) Line() int {
	return s.CurrentLine
}

func (s MultiKeyword) String() string {
	var result []string
	for _, t := range s.Tokens {
		result = append(result, t.String())
	}

	return strings.Join(result, "")
}

type RedisToken struct {
	Node
	TokenType token.Type
	From      int
	To        int
	LineFrom  int
	LineTo    int
	Value     string
}

func (t RedisToken) Type() token.Type {
	return t.TokenType
}

func (t RedisToken) Start() int {
	return t.From
}

func (t RedisToken) End() int {
	return t.To
}

func (t RedisToken) LineStart() int {
	return t.LineFrom
}

func (t RedisToken) LineEnd() int {
	return t.LineTo
}

func (t RedisToken) String() string {
	return t.Value
}

func NewToken(token token.Token) RedisToken {
	return RedisToken{
		TokenType: token.Type,
		From:      token.Start,
		To:        token.End,
		LineFrom:  token.LineStart,
		LineTo:    token.LineEnd,
		Value:     token.Value,
	}
}

func New(tokens []RedisToken) []TokenList {
	return parseMultiKeywords(parseStatements(tokens))
}

func GetSelectedStatement(tokens []TokenList, line int, position int) (TokenList, int) {
	for _, s := range tokens {
		if s.Line() != line {
			continue
		}

		if position >= s.LineStart() && position <= s.LineEnd() {
			return s, position - s.LineStart() + 1
		}
	}

	return tokens[len(tokens)-1], 0
}

func parseStatements(tokens []RedisToken) []TokenList {
	var result []TokenList

	var line int
	tokenList := &Statement{CurrentLine: line}
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		tokenList.AddToken(t)
		if (t.TokenType == token.Semicolon || t.TokenType == token.Newline) && i+1 < len(tokens) {
			result = append(result, tokenList)
			if t.TokenType == token.Newline {
				line++
			}
			tokenList = &Statement{CurrentLine: line}
		}
	}

	result = append(result, tokenList)

	return result
}

var multiKeywords = map[string][]string{
	"ACL":     {"LOAD", "SAVE", "LIST", "USERS", "GETUSER", "SETUSER", "DELUSER", "CAT", "GENPASS", "WHOAMI", "LOG", "HELP"},
	"CLIENT":  {"CACHING", "ID", "INFO", "KILL", "LIST", "GETNAME", "GETREDIR", "UNPAUSE", "PAUSE", "REPLY", "SETNAME", "TRACKING", "TRACKINGINFO", "UNBLOCK"},
	"CLUSTER": {"ADDSLOTS", "BUMPEPOCH", "COUNT-FAILURE-REPORTS", "COUNTKEYSINSLOT", "DELSLOTS", "FAILOVER", "FLUSHSLOTS", "FORGET", "GETKEYSINSLOT", "INFO", "KEYSLOT", "MEET", "MYID", "NODES", "REPLICATE", "RESET", "SAVECONFIG", "SET-CONFIG-EPOCH", "SETSLOT", "SLAVES", "REPLICAS", "SLOTS"},
	"COMMAND": {"COUNT", "GETKEYS", "INFO"},
	"CONFIG":  {"GET", "REWRITE", "SET", "RESETSTAT"},
	"DEBUG":   {"OBJECT", "SEGFAULT"},
	"MEMORY":  {"DOCTOR", "HELP", "MALLOC-STATS", "PURGE", "STATS", "USAGE"},
	"MODULE":  {"LIST", "LOAD", "UNLOAD"},
	"SCRIPT":  {"DEBUG", "EXISTS", "FLUSH", "KILL", "LOAD"},
	"LATENCY": {"DOCTOR", "GRAPH", "HISTORY", "LATEST", "RESET", "HELP"},
}

func parseMultiKeywords(tokens []TokenList) []TokenList {
	for _, t := range tokens {
		innerTokens := t.GetTokens()
		var newTokens []Node
		multiKeyword := MultiKeyword{CurrentLine: t.Line()}

		for i := 0; i < len(innerTokens); i++ {
			innerToken := innerTokens[i]

			if innerToken.Type() == token.Space {
				if len(multiKeyword.Tokens) > 0 {
					multiKeyword.Tokens = append(multiKeyword.Tokens, innerToken)
				} else {
					newTokens = append(newTokens, innerToken)
				}

				continue
			}

			expectedKeywords, ok := multiKeywords[innerToken.String()]
			if ok {
				if len(multiKeyword.Tokens) == 0 {
					if !t.NextTokenIs(expectedKeywords, i+1) {
						newTokens = append(newTokens, innerToken)
					} else {
						multiKeyword.Tokens = append(multiKeyword.Tokens, innerToken)
					}

					continue
				}
			} else {
				if len(multiKeyword.Tokens) > 0 {
					multiKeyword.Tokens = append(multiKeyword.Tokens, innerToken)
					newTokens = append(newTokens, multiKeyword)
					multiKeyword = MultiKeyword{CurrentLine: t.Line()}
					continue
				}
			}

			newTokens = append(newTokens, innerToken)
		}

		t.SetTokens(newTokens)
	}

	return tokens
}

func hasExpectedKeyword(expected []string, actual string) bool {
	for _, v := range expected {
		if v == actual {
			return true
		}
	}

	return false
}
