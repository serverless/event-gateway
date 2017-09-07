package pathtree

import (
	"strings"

	"github.com/serverless/event-gateway/functions"
)

// Node is a data structure, inspired by prefix tree, used for routing HTTP requests in the Event Gateway. It's used for creating tree structure
// of segments in HTTP paths. Each segments is stored in separate node.
type Node struct {
	segment     string
	children    map[string]*Node
	functionID  *functions.FunctionID
	parameter   string
	isParameter bool
	isWildcard  bool
}

// NewNode creates new Node.
func NewNode() *Node {
	return &Node{
		children: map[string]*Node{},
	}
}

// Params is array for URL parameter
type Params map[string]string

// AddRoute adds route to the tree. This function will panic in case of adding conflicting parameterized paths.
func (n *Node) AddRoute(route string, functionID functions.FunctionID) {
	if route == "/" {
		n.functionID = &functionID
		return
	}

	segments := toSegments(route)
	currentNode := n

	for i, segment := range segments {
		// look for static route
		child, exists := currentNode.children[segment]
		if !exists {
			// look for param
			child, exists = first(currentNode.children)
			if !exists || !child.isParameter {
				child = NewNode()
				child.segment = segment
				currentNode.children[segment] = child
			}
			if child.segment != segment {
				panic("route " + route + " has a conflicting segment with existing route")
			}
		}

		currentNode = child

		if strings.HasPrefix(segment, ":") {
			currentNode.isParameter = true
			currentNode.parameter = strings.TrimPrefix(segment, ":")
		}

		if strings.HasPrefix(segment, "*") {
			if len(segments) > i+1 {
				panic("wildcard parameter must be the last parameter")
			}

			currentNode.isParameter = true
			currentNode.isWildcard = true
			currentNode.parameter = strings.TrimPrefix(segment, "*")
		}

		if i == len(segments)-1 {
			currentNode.functionID = &functionID
		}
	}
}

// DeleteRoute deletes route from the tree. This function will panic in case of removing non-existing node.
func (n *Node) DeleteRoute(route string) {
	if route == "/" {
		n.functionID = nil
		return
	}

	segments := toSegments(route)
	currentNode := n

	for i, segment := range segments {
		if i == len(segments)-1 {
			_, exists := currentNode.children[segment]
			if !exists {
				panic("unable to delete node non existing node")
			}

			if len(currentNode.children[segment].children) == 0 {
				delete(currentNode.children, segment)
			} else {
				currentNode.children[segment].functionID = nil
			}

			return
		}

		currentNode = currentNode.children[segment]
	}
}

// Resolve takes request URL path and traverse the tree trying find corresponding route.
func (n *Node) Resolve(path string) (*functions.FunctionID, Params) {
	if path == "/" {
		if n.functionID != nil {
			return n.functionID, nil
		}
		return nil, nil
	}

	segments := toSegments(path)
	currentNode := n
	params := Params{}

	for i, segment := range segments {
		// look for static route
		child, exists := currentNode.children[segment]
		if !exists {
			// look for param
			child, exists = first(currentNode.children)
			if !exists || !child.isParameter {
				return nil, nil
			}
		}
		currentNode = child

		if currentNode.isParameter {
			params[currentNode.parameter] = segment
		}

		if currentNode.isWildcard {
			// add missing parts
			params[currentNode.parameter] = strings.Join(segments[i:], "/")
			return currentNode.functionID, params
		}

		if i == len(segments)-1 {
			return currentNode.functionID, params
		}
	}

	return nil, nil
}

func toSegments(route string) []string {
	segments := strings.Split(route, "/")
	// remove first "" element
	_, segments = segments[0], segments[1:]

	return segments
}

func first(m map[string]*Node) (*Node, bool) {
	for _, v := range m {
		return v, true
	}
	return nil, false
}
