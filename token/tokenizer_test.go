package token

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		Name           string
		Tokens         string
		ExpectedCount  int
		ExpectedTokens []string
	}{
		{
			"Multiple tokens with string",
			"SET test \"testing\";GET test",
			9,
			[]string{"SET", " ", "test", " ", "\"testing\"", ";", "GET", " ", "test"},
		},
		{
			"Multiple tokens with special characters identifier",
			"GET user:123:comments",
			3,
			[]string{"GET", " ", "user:123:comments"},
		},
		{
			"Tokens with newline",
			"GET user\n",
			4,
			[]string{"GET", " ", "user", "<newline>"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			tokenizer := Tokenizer{}
			tokens := tokenizer.Tokenize(test.Tokens)

			if len(tokens) != test.ExpectedCount {
				t.Errorf("%v - Unexpected amount of tokens: %v (expected %v)", test.Name, len(tokens), test.ExpectedCount)
			}

			for i, token := range tokens {
				if token.Value != test.ExpectedTokens[i] {
					t.Errorf("%v - Unexpected token: %v (expected %v)", test.Name, token.Value, test.ExpectedTokens[i])
				}
			}
		})
	}

	indexTests := []struct{
		Name string
		Tokens string
		Start []int
		End []int
		LineStart []int
		LineEnd []int
	}{
		{
			"Tokens",
			"GET test",
			[]int { 0, 3, 4 },
			[]int { 2, 3, 7 },
			[]int { 0, 3, 4 },
			[]int { 2, 3, 7 },
		},
		{
			"Tokens with newline",
			"GET test\nGET test2",
			[]int { 0, 3, 4, 8, 9, 12, 13 },
			[]int { 2, 3, 7, 8, 11, 12, 17 },
			[]int { 0, 3, 4, 8, 0, 3, 4 },
			[]int { 2, 3, 7, 8, 2, 3, 8 },
		},
	}

	for _, test := range indexTests {
		t.Run(test.Name, func(t *testing.T) {
			tokenizer := Tokenizer{}
			tokens := tokenizer.Tokenize(test.Tokens)

			for i, token := range tokens {
				if token.Start != test.Start[i] {
					t.Errorf("%v - Unexpected start %v for token: %v (expected %v)", test.Name, token.Start, token.Value, test.Start[i])
				}

				if token.End != test.End[i] {
					t.Errorf("%v - Unexpected end %v for token: %v (expected %v)", test.Name, token.End, token.Value, test.End[i])
				}

				if token.LineStart != test.LineStart[i] {
					t.Errorf("%v - Unexpected line start %v for token: %v (expected %v)", test.Name, token.LineStart, token.Value, test.LineStart[i])
				}

				if token.LineEnd != test.LineEnd[i] {
					t.Errorf("%v - Unexpected line end %v for token: %v (expected %v)", test.Name, token.LineEnd, token.Value, test.LineEnd[i])
				}
			}
		})
	}
}
