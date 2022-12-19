package jazzy

import (
	"bytes"
	"strings"
)

const (
	static = iota
	pathParam
	none

	colon = ':'
)

type (
	param struct {
		key   string
		value string
	}

	node struct {
		middlewares []HandleFunc
		handler     HandleFunc
		prefix      string
		handleType  int
		methods     []string
		param       *param
		path        string
		parent      *node
		children    []*node
		leaf        bool
	}

	Router struct {
		tree *node
	}
)

func NewRouter() *Router {
	return &Router{
		tree: &node{
			prefix:   "",
			children: []*node{},
			handler:  nil,
			methods:  []string{},
		},
	}
}

func (r *Router) Insert(method, path string, handler HandleFunc) {
	n := r.tree

	if path == "" {
		path += "/"
	}

	if path == "/" {
		n.handler = handler
		return
	}

	path = path[1:]

	for i := 0; i < len(path); i++ {

	}
}

// slashindex
// return first '/' index
// @return slashindex
func slashindex(suffix string) int {
	for i := 0; i < len(suffix); i++ {
		if suffix[i] == '/' {
			return i
		}
	}
	return 0
}

func suf(suffix string, l int) string {
	if len(suffix) == 1 {
		return suffix
	}
	return suffix[l:]
}

func suffixSlash(suffix string) string {
	if suffix[len(suffix)-1] == '/' {
		return suffix
	}
	return suffix + "/"
}

func newNode(
	middlewares []HandleFunc,
	handler HandleFunc,
	prefix string,
	path string,
	handleType int,
	method string,
	parent *node,
) *node {
	return &node{
		middlewares: middlewares,
		handler:     handler,
		prefix:      prefix,
		path:        path,
		handleType:  handleType,
		methods:     []string{method},
		parent:      parent,
	}
}

func (r *Router) Search(method, path string) (HandleFunc, []*param) {
	// search root
	return nil, nil
}

func paramChild(n []*node) *node {
	for i := 0; i < len(n); i++ {
		if n[i].handleType == pathParam {
			return n[i]
		}
	}
	return nil
}

func paramName(path string) string {
	buf := bytes.NewBuffer(make([]byte, 0, len(path)))
	if len(path) == 0 {
		return ""
	}
	if path[0] == colon {
		for i := 1; i < len(path); i++ {
			if path[i] != '/' {
				buf.WriteString(string(path[i]))
			}
			if path[i] == '/' {
				return buf.String()
			}
		}
	}
	return ""
}

func lcpMinChild(n *node, suffix string) (*node, int) {
	mn := len(suffix)
	next := n

	for i := 0; i < len(n.children); i++ {
		l := lcp(n.children[i].prefix, suffix)
		if l <= mn && l != 0 {
			mn = l
			next = n.children[i]
		}
	}
	return next, mn
}

func lcpMinChildren(ns []*node, suffix string) []*node {
	rns := make([]*node, 0, len(ns))
	l := len(suffix)
	for i := 0; i < len(ns); i++ {
		mnl := min(l, len(ns[i].prefix))
		if ns[i].prefix[:mnl] == suffix {
			rns = append(rns, ns[i])
		}
	}
	return rns
}

func remove(ns []*node, n *node) []*node {
	for i := 0; i < len(ns); i++ {
		if ns[i] == n {
			return append(ns[:i], ns[i+1:]...)
		}
	}
	return ns
}

func handleType(path string) int {
	if strings.Contains(path, string(colon)) {
		return pathParam
	}
	return static
}

func lcp(x, y string) int {
	for i := 0; i < min(len(x), len(y)); i++ {
		if x[i] != y[i] {
			return i
		}
	}
	return min(len(x), len(y))
}

func handleindex(n *node, method string) int {
	for i := 0; i < len(n.methods); i++ {
		if n.methods[i] == method {
			return i
		}
	}
	return -1
}

func min[T ~int | ~int32 | ~int64 | ~float32 | ~float64](x, y T) T {
	if x < y {
		return x
	}
	return y
}

func max[T int | int32 | int64 | float32 | float64](x, y T) T {
	if x > y {
		return x
	}
	return y
}
