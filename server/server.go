package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fagnercarvalho/redis-lsp/ast"
	"github.com/fagnercarvalho/redis-lsp/client"
	"github.com/fagnercarvalho/redis-lsp/completer"
	"github.com/fagnercarvalho/redis-lsp/token"
	"github.com/go-redis/redis/v8"
	"github.com/sourcegraph/jsonrpc2"
	"log"
	"strings"
)

// JSON RPC 2 server with handlers for LSP initialization and completion

// https://microsoft.github.io/language-server-protocol/specifications/specification-current/
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#initialize
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#textDocument_completion
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#completionItem_resolve

// https://www.jsonrpc.org/specification

type Server struct {
	files map[string]string
	redis client.Redis
	completer completer.Completer
}

func New(address string, username string, password string, db int, dbCache bool) (Server, error) {
	client, err := client.New(address, username, password, db, dbCache)
	if err != nil {
		return Server{}, err
	}

	completer := completer.Completer{Users: client.Users, Keys: client.Keys}

	return Server{files: map[string]string{}, redis: client, completer: completer}, nil
}

func (s Server) Handle(ctx context.Context, conn *jsonrpc2.Conn, request *jsonrpc2.Request) (result interface{}, err error) {
	log.Printf("handling %v \n", request.Method)

	switch request.Method {
	case "initialize":
		return handleInitialize()
	case "textDocument/completion":
		return s.handleCompletion(request.Params)
	case "textDocument/didOpen":
		return s.handleOpen(request.Params)
	case "textDocument/didChange":
		return s.handleChange(request.Params)
	case "completionItem/resolve":
		return s.handleCompletionResolve(request.Params)
	case "workspace/executeCommand":
		return s.handleWorkspaceExecuteCommand(ctx, request.Params, conn)
	case "initialized":
		// do nothing
		return nil, nil
	}
	// TODO: implement hover, setTraceNotification (for settings changes)
	errorMessage := fmt.Sprintf("method not handled: %v", request.Method)
	log.Println(errorMessage)

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: errorMessage}
}

func handleInitialize() (interface{}, error) {
	return InitializeResult{
		Capabilities: Capabilities{
			TextDocumentSync: KindFull,
			CompletionProvider: CompletionOptions{
				ResolveProvider:   true,
			},
			ExecuteCommandProvider: ExecuteCommandOptions{
				Commands: []string { "server.executeCommand" },
			},
			HoverProvider: false,
			SelectionRangeProvider: true,
		},
	}, nil
}

func (s Server) handleOpen(params *json.RawMessage) (interface{}, error) {
	var request DidOpenTextDocumentParams
	err := json.Unmarshal(*params, &request)
	if err != nil {
		return nil, err
	}

	s.files[request.TextDocument.Uri] = request.TextDocument.Text

	return nil, nil
}

func (s Server) handleChange(params *json.RawMessage) (interface{}, error) {
	var request DidChangeTextDocumentParams
	err := json.Unmarshal(*params, &request)
	if err != nil {
		return nil, err
	}

	s.files[request.TextDocument.Uri] = request.ContentChanges[0].Text

	return nil, nil
}

func (s Server) handleCompletion(params *json.RawMessage) (interface{}, error) {
	var request CompletionParams
	err := json.Unmarshal(*params, &request)
	if err != nil {
		return nil, err
	}

	text := s.files[request.TextDocument.Uri]
	commands := s.completer.Complete(text, request.Position.Line, request.Position.Character)

	var items []CompletionItem
	for _, c := range commands {
		items = append(items, CompletionItem{Label: c, Kind: Text})
	}

	return items, nil
}

func (s Server) handleCompletionResolve(params *json.RawMessage) (interface{}, error) {
	var request CompletionItem
	err := json.Unmarshal(*params, &request)
	if err != nil {
		return nil, err
	}

	bytes, err := completer.GetDocumentation(request.Label)
	if err != nil {
		// if there is no documentation it probably means this is not a keyword
		// let's swallow this error
		log.Printf("error while getting documentation for label %v: %v", request.Label, err)
		return request, nil
	}

	request.Documentation = MarkupContent{Kind: Markdown, Value: fmt.Sprintf("### %v \n %v", request.Label, string(bytes))}

	return request, nil
}

func (s Server) handleWorkspaceExecuteCommand(ctx context.Context, params *json.RawMessage, conn *jsonrpc2.Conn) (interface{}, error) {
	var request ExecuteCommandParams
	err := json.Unmarshal(*params, &request)
	if err != nil {
		return nil, err
	}

	//tokens :=  // strings.Split(request.Arguments[0].(string), " ")

	tokenizer := token.Tokenizer{}
	tokens := tokenizer.Tokenize(request.Arguments[0].(string))

	var redisTokens []ast.RedisToken
	for _, t := range tokens {
		token := ast.NewToken(t)
		redisTokens = append(redisTokens, token)
	}

	statements := ast.New(redisTokens)

	var commands [][]interface{}
	for _, statement := range statements {
		stmtTokens := statement.GetTokens()
		var new []interface{}
		for _, t := range stmtTokens {
			if t.Type() == token.Semicolon || t.Type() == token.Space {
				continue
			}

			if t.Type() == token.MultiKeyword {
				split :=  strings.Split(t.String(), " ")
				for _, s := range split {
					new = append(new, s)
				}

				continue
			}

			new = append(new, t.String())
		}

		commands = append(commands, new)
	}

	for _, command := range commands {
		val, err := s.redis.ExecuteCommand(ctx, command)
		if err != nil {
			if err == redis.Nil {
				logMessage := LogMessageParams{
					Message: err.Error(),
					Type: Error,
				}

				err = conn.Notify(context.Background(), "window/logMessage", logMessage)
				if err != nil {
					return nil, err
				}

				return nil, nil
			}

			logMessage := ShowMessageParams{
				Message: err.Error(),
				Type: Error,
			}

			err = conn.Notify(context.Background(), "window/showMessage", logMessage)
			if err != nil {
				return nil, err
			}

			return nil, nil
		}

		logMessage := LogMessageParams{
			Message: fmt.Sprintf("%v", val),
			Type: Log,
		}

		err = conn.Notify(context.Background(), "window/logMessage", logMessage)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}
