package jazzy

import (
	"bytes"
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
		r.tree.handler = handler
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

		n.methods[method] = handler
		n.originalPath = originalPath
		return
	}
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

func (r *Router) Search(method, path string) (HandleFunc, []*param) {
	// search root
	return nil, nil
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
func lcp(x, y string) int {
	for i := 0; i < min(len(x), len(y)); i++ {
		if x[i] != y[i] {
			return i
		}
	}
	return min(len(x), len(y))
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
