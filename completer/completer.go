package completer

import (
	"github.com/fagnercarvalho/redis-lsp/ast"
	"github.com/fagnercarvalho/redis-lsp/token"
)

type Completer struct {
	Users []string
	Keys []string
}

func (c Completer) Complete(text string, line int, position int) []string {
	tokenizer := token.Tokenizer{}
	tokens := tokenizer.Tokenize(text)

	var redisTokens []ast.RedisToken
	for _, t := range tokens {
		token := ast.NewToken(t)
		redisTokens = append(redisTokens, token)
	}

	statements := ast.New(redisTokens)
	statement, endIndex := ast.GetSelectedStatement(statements, line, position-1)

	if statement.PreviousTokenIs([]string{"ACL GETUSER"}, position-1) {
		return c.Users
	} else if  statement.PreviousTokenIs([]string{"GET", "SET"}, position-1) {
		return c.Keys
	} else {
		return getCommands(statement.String()[:endIndex])
	}
}
