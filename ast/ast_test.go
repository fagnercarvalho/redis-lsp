package ast

import (
	"github.com/fagnercarvalho/redis-lsp/token"
	"testing"
)

func TestParseStatements(t *testing.T) {
	tests := []struct {
		Name               string
		Statements         string
		ExpectedCount      int
		ExpectedStatements []string
	}{
		{
			"Multiple statements",
			"SET test \"testing\";GET test",
			2,
			[]string{"SET test \"testing\";", "GET test"},
		},
		{
			"Single statement without semicolon",
			"GET test",
			1,
			[]string{"GET test"},
		},
		{
			"Single statement with semicolon",
			"GET test;",
			1,
			[]string{"GET test;"},
		},
		{
			"Three statements",
			"GET test;GET test2;GET test3;",
			3,
			[]string{"GET test;", "GET test2;", "GET test3;"},
		},
		{
			"Two statements with newline",
			"GET test\nGET test2",
			2,
			[]string{"GET test<newline>", "GET test2"},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			tokenizer := token.Tokenizer{}
			tokens := tokenizer.Tokenize(test.Statements)

			var redisTokens []RedisToken
			for _, t := range tokens {
				token := NewToken(t)
				redisTokens = append(redisTokens, token)
			}

			statements := parseStatements(redisTokens)

			if len(statements) != test.ExpectedCount {
				t.Errorf("%v - Unexpected amount of statements: %v (expected %v)", test.Name, len(statements), test.ExpectedCount)
			}

			for i, s := range statements {
				if s.String() != test.ExpectedStatements[i] {
					t.Errorf("%v - Unexpected statement: %v (expected %v)", test.Name, s.String(), test.ExpectedStatements[i])
				}
			}
		})
	}

	lineTests := []struct {
		Name string
		Statements string
		ExpectedLines []int
	}{
		{
			"1 statement",
			"GET test1",
			[]int {0},
		},
		{
			"2 statements",
			"GET test1\nGET test2",
			[]int {0, 1},
		},
		{
			"3 statements",
			"GET test1\nGET test2\nGET test3",
			[]int {0, 1, 2},
		},
	}

	for _, test := range lineTests {
		t.Run(test.Name, func(t *testing.T) {
			tokenizer := token.Tokenizer{}
			tokens := tokenizer.Tokenize(test.Statements)

			var redisTokens []RedisToken
			for _, t := range tokens {
				token := NewToken(t)
				redisTokens = append(redisTokens, token)
			}

			statements := parseStatements(redisTokens)
			result := parseMultiKeywords(statements)
			for i, s := range result {
				if s.Line() != test.ExpectedLines[i] {
					t.Errorf("%v - Unexpected line number: %v (expected %v)", test.Name, s.Line(), test.ExpectedLines[i])
				}
			}
		})
	}
}

func TestParseKeywords(t *testing.T) {
	tests := []struct {
		Name                 string
		Statements           string
		ExpectedCount        int
		ExpectedMultiKeyword []string
	}{
		{
			"Single multikeyword with identifier",
			"ACL GETUSER default",
			1,
			[]string{"ACL GETUSER"},
		},
		{
			"Multiple multikeyword",
			"ACL GETUSER default;ACL LIST",
			2,
			[]string{"ACL GETUSER", "ACL LIST"},
		},
		{
			"Single multikeyword with hyphens",
			"CLUSTER COUNT-FAILURE-REPORTS",
			1,
			[]string{"CLUSTER COUNT-FAILURE-REPORTS"},
		},
		{
			"Single multikeyword",
			"LATENCY LATEST",
			1,
			[]string{"LATENCY LATEST"},
		},
		{
			"No multikeyword",
			"GET test",
			0,
			nil,
		},
		{
			"Single keyword",
			"HELLO",
			0,
			nil,
		},
		{
			"Incomplete multikeyword",
			"COMMAND",
			0,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			tokenizer := token.Tokenizer{}
			tokens := tokenizer.Tokenize(test.Statements)

			var redisTokens []RedisToken
			for _, t := range tokens {
				token := NewToken(t)
				redisTokens = append(redisTokens, token)
			}

			statements := parseStatements(redisTokens)
			result := parseMultiKeywords(statements)
			count := 0
			for _, s := range result {
				for _, v := range s.GetTokens() {
					if v.Type() == token.MultiKeyword {
						if v.String() != test.ExpectedMultiKeyword[count] {
							t.Errorf("%v - Unexpected multikeyword token: %v (expected %v)", test.Name, v.String(), test.ExpectedMultiKeyword[count])
						}

						count++
					}
				}
			}

			if count != test.ExpectedCount {
				t.Errorf("%v - Unexpected amount of multikeyword tokens: %v (expected %v)", test.Name, count, test.ExpectedCount)
			}
		})
	}
}

func TestPreviousTokenIs(t *testing.T) {
	tests := []struct {
		Name          string
		Statements    string
		PreviousToken string
	}{
		{
			"Should return previous multi keyword token",
			"LATENCY HELP",
			"LATENCY HELP",
		},
		{
			"Should return previous keyword token",
			"GET",
			"GET",
		},
		{
			"Should return previous keyword token",
			"LATENCY RESET;",
			";",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			tokenizer := token.Tokenizer{}
			tokens := tokenizer.Tokenize(test.Statements)

			var redisTokens []RedisToken
			for _, t := range tokens {
				token := NewToken(t)
				redisTokens = append(redisTokens, token)
			}

			statements := New(redisTokens)

			if !statements[len(statements)-1].PreviousTokenIs([]string{test.PreviousToken}, len(statements[len(statements)-1].String())) {
				t.Errorf("expected %v keyword in statements %v", test.PreviousToken, test.Statements)
			}
		})
	}
}

func TestGetSelectedStatement(t *testing.T) {
	tests := []struct {
		Name string
		Statements string
		Position int
		Line int
		ExpectedStatement string
	}{
		{
			"1 statement",
			"GET test1",
			5,
			1,
			"GET test1",
		},
		{
			"2 statements",
			"GET test1\nGET test2",
			1,
			2,
			"GET test2",
		},
		{
			"3 statements",
			"GET test1\nGET test2\nGET test3",
			3,
			3,
			"GET test3",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			tokenizer := token.Tokenizer{}
			tokens := tokenizer.Tokenize(test.Statements)

			var redisTokens []RedisToken
			for _, t := range tokens {
				token := NewToken(t)
				redisTokens = append(redisTokens, token)
			}

			statements := parseStatements(redisTokens)
			result := parseMultiKeywords(statements)
			statement, _ := GetSelectedStatement(result, test.Line, test.Position)
			if statement.String() != test.ExpectedStatement {
				t.Fatalf("%v - Unexpected statement: %v (expected %v)", test.Name, statement.String(), test.ExpectedStatement)
			}
		})
	}
}
