package query

import (
	"fmt"
	"testing"

	"github.com/dvirsky/timedis/query/ast"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {

	q := `{"type":"movingAvg","params":{"window": 10},"upstream":[{"type":"faucet","params":{"key":"foo.bar"}}]}`

	node, err := Parse(q)
	assert.NoError(t, err)

	assert.Equal(t, node.Type, ast.TypeMovingAverage)
	assert.Equal(t, node.Children[0].Type, ast.TypeFaucet)
	assert.Equal(t, node.Children[0].Params["key"], "foo.bar")

	source, err := node.Eval()
	assert.NoError(t, err)
	assert.NotNil(t, source)
	fmt.Printf("%#v", source)
}
