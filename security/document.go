package security

import (
	"encoding/json"
	"strings"
)

// CatServerCode marks a command document that asks the server to execute
// JavaScript, the usual vehicle of NoSQL injection.
const CatServerCode Category = "server-side-code"

// serverCodeOperators run caller-supplied JavaScript on a MongoDB server.
var serverCodeOperators = map[string]bool{
	"$where":       true,
	"$function":    true,
	"$accumulator": true,
	"mapreduce":    true,
	"$eval":        true,
}

// IsDocument reports whether query is a JSON command document (for example a
// MongoDB command such as {"find": "users", "filter": {...}}) rather than SQL.
func IsDocument(query string) bool {
	t := strings.TrimSpace(query)
	return strings.HasPrefix(t, "{") && json.Valid([]byte(t))
}

// AnalyzeDocument inspects a JSON command document. SQL invariants do not
// apply to documents; instead the document is decoded and rejected if any key,
// at any depth, invokes server-side JavaScript ($where, $function,
// $accumulator, mapReduce). Documents that are not valid JSON are rejected.
func (e *SecurityEngine) AnalyzeDocument(doc string) Decision {
	var v any
	if err := json.Unmarshal([]byte(doc), &v); err != nil {
		return block(CatUnparseable, "invalid command document: %v", err)
	}
	if op := findServerCode(v); op != "" {
		return block(CatServerCode, "command document uses %s, which executes JavaScript on the server", op)
	}
	return Decision{}
}

func findServerCode(v any) string {
	switch t := v.(type) {
	case map[string]any:
		for k, child := range t {
			if serverCodeOperators[strings.ToLower(k)] {
				return k
			}
			if op := findServerCode(child); op != "" {
				return op
			}
		}
	case []any:
		for _, child := range t {
			if op := findServerCode(child); op != "" {
				return op
			}
		}
	}
	return ""
}
