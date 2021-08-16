package token

import (
	"unicode"
)

type Token struct {
	Type      Type
	Start     int
	End       int

	// Since the LSP sends the cursor position based on the current line only
	// we need to store the relative start and end of each token
	// For example, if we are in the first position of the 3rd line we are on index 0 and not 2 (if the first two lines were empty).
	LineStart int
	LineEnd   int

	Value     string
}

type Tokenizer struct {
}

func (t *Tokenizer) Tokenize(value string) []Token {
	var tokens []Token

	start := 0
	lineStart := 0
	for {
		token := t.nextToken(value, start, lineStart)
		tokens = append(tokens, token)

		start = token.End + 1
		lineStart = token.LineEnd + 1

		if start == len(value) {
			break
		}

		if token.Type == Newline {
			lineStart = 0
		}
	}

	return tokens
}

func (*Tokenizer) nextToken(value string, start int, lineStart int) Token {
	for i, r := range value[start:] {
		switch {
		case r == '\n':
			return Token{Start: start, End: start + i, LineStart: lineStart, LineEnd: lineStart + i, Type: Newline, Value: "<newline>"}
		case unicode.IsSpace(r):
			return Token{Start: start, End: start + i, LineStart: lineStart, LineEnd: lineStart + i, Type: Space, Value: " "}
		case r == ';':
			return Token{Start: start, End: start + i, LineStart: lineStart, LineEnd: lineStart + i, Type: Semicolon, Value: ";"}
		case r == '"':
			return nextString(value, start+i, lineStart+i)
		case r == '\\':
			continue
		default:
			return nextIdentifier(value, start+i, lineStart+i)
		}
	}

	return Token{Start: start, End: start + 1, Type: Unknown}
}

func nextIdentifier(value string, start int, lineStart int) Token {
	var identifier []rune
	var end int
	for i, r := range value[start:] {
		if unicode.IsSpace(r) || r == ';' || r == '"' || r == '\\' {
			break
		}
		identifier = append(identifier, r)
		end = i
	}

	return Token{Start: start, End: start + end, LineStart: lineStart, LineEnd: lineStart + end, Type: Match(string(identifier)), Value: string(identifier)}
}

func nextString(value string, start int, lineStart int) Token {
	var identifier []rune
	var end int
	for i, r := range value[start:] {
		identifier = append(identifier, r)
		end = i
		if r == '"' && i != 0 {
			break
		}
	}
	return Token{Start: start, End: start + end, LineStart: lineStart, LineEnd: lineStart + end, Type: String, Value: string(identifier)}
}
