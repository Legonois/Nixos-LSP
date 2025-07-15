package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

func main() {
	ctx := context.Background()

	stream := jsonrpc2.NewStream(os.Stdin)
	conn := jsonrpc2.NewConn(stream)
	handler := &server{conn: conn, files: make(map[protocol.URI]string)}
	conn.Go(ctx, handler.Handler)
	<-conn.Done()
}

type server struct {
	conn  jsonrpc2.Conn
	files map[protocol.URI]string // map of file URIs to their contents
}

func (s *server) Handler(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	switch req.Method() {

	// initialize
	// handle initialize request
	case protocol.MethodInitialize:
		var params protocol.InitializeParams
		if err := json.Unmarshal(req.Params(), &params); err != nil {
			return err
		}
		reply(ctx, protocol.InitializeResult{
			Capabilities: protocol.ServerCapabilities{
				TextDocumentSync: protocol.TextDocumentSyncKindIncremental,
				SemanticTokensProvider: protocol.SemanticTokensOptions{
					WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
						WorkDoneProgress: true,
					},
				},
			},
		}, nil)
		return nil

	// initialized
	// client is ready; no reply expected
	case protocol.MethodInitialized:
		return nil

	// shutdown
	// handle shutdown request
	case protocol.MethodShutdown:
		reply(ctx, nil, nil)
		return nil

	// exit
	// handle exit notification
	case protocol.MethodExit:
		os.Exit(0)
		return nil

	// textDocument/didOpen
	// handle open document
	case protocol.MethodTextDocumentDidOpen:
		var params protocol.DidOpenTextDocumentParams
		if err := json.Unmarshal(req.Params(), &params); err != nil {
			return err
		}

		s.files[params.TextDocument.URI] = params.TextDocument.Text
		log.Printf("Opened %s", params.TextDocument.URI)
		return nil

	// textDocument/didChange
	// handle change in document
	case protocol.MethodTextDocumentDidChange:
		var params protocol.DidChangeTextDocumentParams
		if err := json.Unmarshal(req.Params(), &params); err != nil {
			return err
		}

		// store full document text (since TextDocumentSyncKindFull)
		if len(params.ContentChanges) > 0 {
			s.files[params.TextDocument.URI] = params.ContentChanges[0].Text
		}
		log.Printf("Changed %s (version %d)",
			params.TextDocument.URI, params.TextDocument.Version)
		return nil

	// textDocument/completion
	// handle completion request
	case protocol.MethodTextDocumentCompletion:
		var params protocol.CompletionParams
		if err := json.Unmarshal(req.Params(), &params); err != nil {
			return err
		}

		items := []protocol.CompletionItem{{
			Label: "HelloWorld",
			Kind:  protocol.CompletionItemKindText,
		}}

		reply(ctx, protocol.CompletionList{
			IsIncomplete: false,
			Items:        items,
		}, nil)
		return nil

	// textDocument/hover
	// handle hover request
	case protocol.MethodTextDocumentHover:
		var params protocol.HoverParams
		if err := json.Unmarshal(req.Params(), &params); err != nil {
			return err
		}

		reply(ctx, &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.Markdown,
				Value: "**Hover** example",
			},
			Range: &protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 0, Character: 5},
			},
		}, nil)
		return nil

	// textDocument/definition
	// handle go-to-definition request
	case protocol.MethodTextDocumentDefinition:
		var params protocol.DefinitionParams
		if err := json.Unmarshal(req.Params(), &params); err != nil {
			return err
		}

		loc := protocol.Location{
			URI: params.TextDocument.URI,
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 0, Character: 5},
			},
		}

		reply(ctx, []protocol.Location{loc}, nil)
		return nil

	default:
		// unhandled method
		log.Printf("Unhandled method: %s", req.Method())
		return jsonrpc2.MethodNotFoundHandler(ctx, reply, req)
	}
}
