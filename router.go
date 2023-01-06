package jazzy

import (
	"bytes"
	"fmt"
)

const (
	static kind = iota
	pathParam
	none

	colon = ':'
)

type kind int

type (
	param struct {
		key   string
		value string
	}
)

type node struct {
	middlewares  []HandleFunc
	handler      HandleFunc
	prefix       string
	kind         kind
	methods      map[string]HandleFunc
	param        *param
	path         string
	originalPath string
	parent       *node
	children     []*node
	leaf         bool
}

func (n *node) findMaxLengthChild(path string, k kind) *node {
	var maxLengthNode *node
	maxLength := 0
	for _, child := range n.children {
		max := len(child.prefix)
		if max > len(path) {
			max = len(path)
		}
		lcp := 0
		for ; lcp < max && child.prefix[lcp] == path[lcp]; lcp++ {
		}
		if maxLength < lcp {
			maxLength = lcp
			maxLengthNode = child
		}
	}
	return maxLengthNode
}

type Router struct {
	tree *node
}

func NewRouter() *Router {
	return &Router{
		tree: &node{
			prefix:   "",
			children: []*node{},
			handler:  nil,
			methods:  make(map[string]HandleFunc),
		},
	}
}

func (r *Router) Insert(method, path string, handler HandleFunc) {
	originalPath := path

	if path == "" {
		path += "/"
	}

	if path == "/" {
		r.tree.methods[method] = handler
		// r.tree.handler = handler
		return
	}

	path = path[1:]

	for i := 0; i < len(path); i++ {
		if path[i] == ':' {

		}
		r.insert(method, path, originalPath, static, handler)
	}
}

func (r *Router) insert(method, path, originalPath string, k kind, handler HandleFunc) {
	n := r.tree
	for {
		max := len(n.prefix)
		pathl := len(path)

		if max < pathl {
			max = pathl
		}

		lcpIndex := 0
		for ; lcpIndex < max; lcpIndex++ {
			if path[lcpIndex] != n.prefix[lcpIndex] {
				break
			}
		}

		if lcpIndex == 0 {
			n.prefix = path
			n.kind = k
			n.originalPath = originalPath
			n.methods[method] = handler
		}

		if lcpIndex < pathl {
			path = path[lcpIndex:]
			nn := newNode(
				nil,
				nil,
				n.prefix[:lcpIndex],
				originalPath,
				k,
				method,
				n)

			for _, child := range n.children {
				child.parent = nn
			}

			n.update(n.prefix[:lcpIndex], nn)
			nn.children = append(nn.children, n)

			if lcpIndex == len(path) {
				nn.kind = k
				if handler != nil {
					nn.methods[method] = handler
				}
			} else {
				nnn := newNode(
					nil,
					nil,
					n.prefix[:lcpIndex],
					"",
					k,
					method,
					nn)
				nn.children = append(nn.children, nnn)
			}
		}

		fmt.Println("inserted")
		n.methods[method] = handler
		n.originalPath = originalPath
		return
	}
}

func (r *Router) Search(method, path string) (HandleFunc, []*param) {
	// search root
	if path == "/" {
		return r.tree.methods[method], nil
	}

	current := r.tree
	target := path

	min := len(target)

	params := make([]*param, 0)
	for {
		if min > len(r.tree.prefix) {
			min = len(r.tree.prefix)
		}

		next := current.findMaxLengthChild(target, static)
		target = target[len(next.prefix):]

		if target == "" {
			return next.methods[method], params
		}
	}

	return nil, nil
}

func newNode(
	middlewares []HandleFunc,
	handler HandleFunc,
	prefix string,
	path string,
	k kind,
	method string,
	parent *node,
) *node {
	return &node{
		middlewares: middlewares,
		handler:     handler,
		prefix:      prefix,
		path:        path,
		kind:        k,
		methods:     make(map[string]HandleFunc),
		children:    make([]*node, 0),
		parent:      parent,
	}
}

func (n *node) update(prefix string, parent *node) {
	n.kind = static
	n.prefix = prefix
	n.originalPath = ""
	n.methods = make(map[string]HandleFunc)
	n.children = nil
	n.handler = nil
	n.middlewares = nil
	n.parent = parent
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
