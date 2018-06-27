package pathtree

import (
	"errors"
	"fmt"
	"strings"
)

// Node is a data structure, inspired by prefix tree, used for routing HTTP requests in the Event Gateway. It's used for creating tree structure
// of segments in HTTP paths. Each segments is stored in separate node.
type Node struct {
	segment     string
	children    map[string]*Node
	parameter   string
	isStatic    bool
	isParameter bool
	isWildcard  bool

	value interface{}
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
// nolint: gocyclo
func (n *Node) AddRoute(route string, value interface{}) error {
	if route == "/" {
		if n.value == nil {
			n.value = value
			return nil
		}
		return fmt.Errorf("route / conflicts with existing route")
	}

	segments := toSegments(route)
	currentNode := n

	for i, segment := range segments {
		if currentNode.isWildcard {
			return errors.New("wildcard must be the last parameter")
		}

		// look for static route
		child, exists := currentNode.children[segment]
		if !exists {
			// look for param
			child, exists = first(currentNode.children)
			if !exists {
				// empty children, create node and go to the next segment
				currentNode.children[segment] = createNode(segment)
				if i == len(segments)-1 {
					currentNode.children[segment].value = value
					return nil
				}
				currentNode = currentNode.children[segment]
				continue
			}
		}

		segmentIsParam := strings.HasPrefix(segment, ":")
		segmentIsWildcard := strings.HasPrefix(segment, "*")
		segmentIsStatic := !segmentIsParam && !segmentIsWildcard

		if child.isWildcard || segmentIsWildcard {
			return fmt.Errorf("wildcard with different name (%q) already defined: for route: %s", child.parameter, route)
		}

		if child.isParameter && child.segment != segment {
			return fmt.Errorf("parameter with different name (%q) already defined: for route: %s", child.parameter, route)
		}

		if child.isStatic && !segmentIsStatic {
			return fmt.Errorf("static route already defined for route: %s", route)
		}

		if currentNode.children[segment] == nil {
			currentNode.children[segment] = createNode(segment)
		}

		if i == len(segments)-1 {

			if currentNode.children[segment].value != nil {
				return fmt.Errorf("route %s conflicts with existing route", route)
			}
			currentNode.children[segment].value = value
			return nil
		}
		currentNode = currentNode.children[segment]
	}

	return nil
}

// DeleteRoute deletes route from the tree. This function will panic in case of removing non-existing node.
func (n *Node) DeleteRoute(route string) error {
	if route == "/" {
		n.value = nil
		return nil
	}

	segments := toSegments(route)
	currentNode := n

	for i, segment := range segments {
		if i == len(segments)-1 {
			_, exists := currentNode.children[segment]
			if !exists {
				return errors.New("unable to delete node non existing node")
			}

			if len(currentNode.children[segment].children) == 0 {
				delete(currentNode.children, segment)
			} else {
				currentNode.children[segment].value = nil
			}

			return nil
		}

		currentNode = currentNode.children[segment]
	}

	return nil
}

// Resolve takes request URL path and traverse the tree trying find corresponding route.
// nolint: gocyclo
func (n *Node) Resolve(path string) (interface{}, Params) {
	if path == "/" {
		if n.value != nil {
			return n.value, nil
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
			if !exists || !(child.isParameter || child.isWildcard) {
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
			return currentNode.value, params
		}

		if i == len(segments)-1 {
			return currentNode.value, params
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

func createNode(segment string) *Node {
	isParam := strings.HasPrefix(segment, ":")
	isWildcard := strings.HasPrefix(segment, "*")
	isStatic := !isParam && !isWildcard

	child := NewNode()
	child.segment = segment

	child.isStatic = isStatic
	child.isParameter = isParam
	child.isWildcard = isWildcard

	if isParam {
		child.parameter = strings.TrimPrefix(segment, ":")
	} else if isWildcard {
		child.parameter = strings.TrimPrefix(segment, "*")
	}

	return child
}
