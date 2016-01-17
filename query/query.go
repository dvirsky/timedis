package query

import (
	"encoding/json"

	"github.com/dvirsky/timedis/query/ast"
)

func Parse(query string) (ast.Node, error) {

	var node ast.Node

	err := json.Unmarshal([]byte(query), &node)
	return node, err
}
