package zen

import (
	"fmt"
	"strings"
)

// param is a single URL parameter.
type param struct {
	Key   string
	Value string
}

// node is one path segment in the routing trie. On lookup, static children
// are tried first, then a named parameter (":id"), then a catch-all ("*path").
type node struct {
	children map[string]*node
	param    *node
	wildcard *node
	name     string // parameter name, for param and wildcard nodes
	handlers []HandlerFunc
}

// insert adds a route to the trie. It returns false if the route already exists
// and panics if the pattern conflicts with an existing one.
func (n *node) insert(pattern string, handlers []HandlerFunc) bool {
	segments := strings.FieldsFunc(pattern, func(r rune) bool { return r == '/' })
	for i, seg := range segments {
		switch seg[0] {
		case ':', '*':
			name := seg[1:]
			if name == "" {
				panic(fmt.Sprintf("zen: parameter in %q must have a name", pattern))
			}
			slot := &n.param
			if seg[0] == '*' {
				if i != len(segments)-1 {
					panic(fmt.Sprintf("zen: catch-all in %q must be the last segment", pattern))
				}
				slot = &n.wildcard
			}
			if *slot == nil {
				*slot = &node{name: name}
			} else if (*slot).name != name {
				panic(fmt.Sprintf("zen: %q conflicts with existing parameter %q", pattern, (*slot).name))
			}
			n = *slot
		default:
			if n.children == nil {
				n.children = make(map[string]*node)
			}
			child := n.children[seg]
			if child == nil {
				child = &node{}
				n.children[seg] = child
			}
			n = child
		}
	}
	if n.handlers != nil {
		return false
	}
	n.handlers = handlers
	return true
}

// search finds the handlers for path, appending any matched parameters to params.
// Empty segments are ignored, so "/users/" and "//users" both match "/users".
func (n *node) search(path string, params []param) ([]HandlerFunc, []param) {
	path = strings.TrimLeft(path, "/")
	if path == "" {
		if n.handlers != nil {
			return n.handlers, params
		}
		if n.wildcard != nil {
			return n.wildcard.handlers, append(params, param{n.wildcard.name, "/"})
		}
		return nil, params
	}

	seg, rest := path, ""
	if i := strings.IndexByte(path, '/'); i >= 0 {
		seg, rest = path[:i], path[i:]
	}

	if child := n.children[seg]; child != nil {
		if h, p := child.search(rest, params); h != nil {
			return h, p
		}
	}
	if n.param != nil {
		if h, p := n.param.search(rest, append(params, param{n.param.name, seg})); h != nil {
			return h, p
		}
	}
	if n.wildcard != nil {
		return n.wildcard.handlers, append(params, param{n.wildcard.name, "/" + path})
	}
	return nil, params
}

// countParams returns the number of parameters in a route pattern.
func countParams(pattern string) int {
	return strings.Count(pattern, "/:") + strings.Count(pattern, "/*")
}
