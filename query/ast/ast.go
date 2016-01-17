package ast

import (
	"errors"

	"github.com/dvirsky/timedis/pipeline"
)

const (
	TypeFilter        = "filter"
	TypeMovingAverage = "movingAvg"
	TypeFaucet        = "faucet"
)

var registry map[string]pipeline.SourceFactory

type Node struct {
	Type     string                 `json:"type"`
	Params   map[string]interface{} `json:"params"`
	Children []Node                 `json:"upstream,omitempty"`
}

func (n Node) Eval() (pipeline.Source, error) {

	f, found := registry[n.Type]
	if !found {
		return nil, errors.New("Invalid source type " + n.Type)
	}

	var children []pipeline.Source

	if n.Children != nil {
		children = make([]pipeline.Source, 0, len(n.Children))

		for _, child := range n.Children {
			if s, err := child.Eval(); err != nil {
				return nil, err
			} else {
				children = append(children, s)
			}

		}
	}

	return f(n.Params, children)

}

func init() {
	registry = map[string]pipeline.SourceFactory{
		TypeFilter:        pipeline.NewFilter,
		TypeMovingAverage: pipeline.NewMovingAverage,
		TypeFaucet:        pipeline.NewFaucet,
	}

}
