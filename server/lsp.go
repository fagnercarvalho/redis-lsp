package server

// initialize

type InitializeResult struct {
	Capabilities Capabilities `json:"capabilities"`
}

type TextDocumentSyncKind int

const (
	KindFull = 1
)

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters"`
	ResolveProvider   bool     `json:"resolveProvider"`
}

type ExecuteCommandOptions struct {
	Commands []string `json:"commands"`
}

type Capabilities struct {
	TextDocumentSync       TextDocumentSyncKind  `json:"textDocumentSync"`
	CompletionProvider     CompletionOptions     `json:"completionProvider"`
	ExecuteCommandProvider ExecuteCommandOptions `json:"executeCommandProvider"`
	HoverProvider          bool                  `json:"hoverProvider"`
	SelectionRangeProvider bool                  `json:"selectionRangeProvider"`
}

// completion

type CompletionItem struct {
	Label         string             `json:"label"`
	Kind          CompletionItemKind `json:"kind"`
	Documentation interface{}        `json:"documentation,omitempty"`
}

type MarkupContent struct {
	Kind  MarkupKind `json:"kind"`
	Value string     `json:"value"`
}

type MarkupKind string

const (
	Markdown  = "markdown"
	PlainText = "plaintext"
)

type CompletionItemKind int

const (
	Text    = 1
	Keyword = 14
)

type CompletionParams struct {
	TextDocumentPositionParams
}

type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type TextDocumentIdentifier struct {
	Uri string `json:"uri"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// didOpen

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type TextDocumentItem struct {
	Uri        string `json:"uri"`
	LanguageId string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// didChange

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type VersionedTextDocumentIdentifier struct {
	Uri string `json:"uri"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

// logMessage

type LogMessageParams struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

type MessageType float64

var (
	Error   MessageType = 1
	Warning MessageType = 2
	Info    MessageType = 3
	Log     MessageType = 4
)

// workspace/executeCommand

type ExecuteCommandParams struct {
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments"`
}

// showMessage

type ShowMessageParams struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}
