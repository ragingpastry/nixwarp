package types

import "fmt"

type NodeError struct {
	Node    string
	Err     error
	Command string
}

func (r *NodeError) Error() string {
	return fmt.Sprintf("Node: %s Error: %s Command: %s", r.Node, r.Err.Error(), r.Command)
}
