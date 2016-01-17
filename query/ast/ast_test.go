package ast

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAST(t *testing.T) {

	tree := Node{
		Type: TypeFilter,
		Params: map[string]interface{}{
			"min": float64(0),
			"max": float64(100),
		},
		Children: []Node{
			{
				Type: TypeFaucet,
				Params: map[string]interface{}{
					"key": "foo.bar",
				},
			},
		},
	}

	source, err := tree.Eval()
	assert.NoError(t, err)

	b, _ := json.Marshal(tree)
	fmt.Println(string(b))
	fmt.Printf("%#v", source)
}
