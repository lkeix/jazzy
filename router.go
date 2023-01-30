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

func (n *node) addChild(child *node) {
	n.children = append(n.children, child)
}

func (n *node) findMaxLengthChild(path string, k kind) *node {
	var maxLengthNode *node
	maxLength := 0

	pathl := len(path)
	for _, child := range n.children {
		max := len(child.prefix)
		if max > pathl {
			max = pathl
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

type route struct {
	originalPath string
	paramNames   []string
	handler      HandleFunc
}

// reference: https://github.com/labstack/echo/blob/master/router.go
func (r *Router) Insert(method, path string, handler HandleFunc) {
	if path == "" {
		path += "/"
	}

	if path[0] != '/' {
		path = "/" + path
	}

	if path == "/" {
		r.tree.methods[method] = handler
		r.tree.prefix = "/"
		// r.tree.handler = handler
		return
	}

	originalPath := path

	lcpIndex := len(path)
	for i := 0; i < lcpIndex; i++ {
		if path[i] == ':' {
			if i > 0 && path[i-1] == '\\' {
				path = path[:i-1] + path[i:]
				i--
				lcpIndex--
			}
			j := i + 1
			r.insert(method, path[:i], static, &route{})

			for ; i < lcpIndex && path[i] != '/'; i++ {
			}

			path = path[:j] + path[i:]
			i = j
			lcpIndex = len(path)

			if i == lcpIndex {
				r.insert(method, path[:i], pathParam, &route{
					handler:      handler,
					originalPath: originalPath,
				})
			} else {
				r.insert(method, path[:i], pathParam, &route{})
			}
		}
	}

	r.insert(method, path, static, &route{
		originalPath: originalPath,
		paramNames:   []string{},
		handler:      handler,
	})
}

func (r *Router) insert(method, path string, k kind, route *route) {
	n := r.tree
	for {
		pathLength := len(path)
		prefixLen := len(n.prefix)

		commonLengthMin := prefixLen
		if commonLengthMin > pathLength {
			commonLengthMin = pathLength
		}

		lcpIndex := 0
		for ; lcpIndex < commonLengthMin; lcpIndex++ {
			if path[lcpIndex] != n.prefix[lcpIndex] {
				break
			}
		}

		if lcpIndex == 0 {
			n.prefix = path
			n.kind = k
			n.originalPath = route.originalPath
			n.methods[method] = route.handler
		}

		if lcpIndex < prefixLen {
			nn := newNode(
				nil,
				nil,
				n.prefix[lcpIndex:],
				n.originalPath,
				k,
				method,
				n)

			nn.methods = n.methods
			for _, child := range n.children {
				child.parent = nn
			}

			n.kind = static
			n.prefix = n.prefix[:lcpIndex]
			n.originalPath = ""
			// n.methods[method] = nil

			n.addChild(nn)
			if lcpIndex == pathLength {
				if route.handler != nil {
					n.methods[method] = route.handler
					n.originalPath = route.originalPath
				}
			} else {
				nn := newNode(nil, nil, path[lcpIndex:], route.originalPath, static, method, n)
				if route.handler != nil {
					nn.methods[method] = route.handler
					nn.originalPath = route.originalPath
				}
				n.addChild(nn)
			}
		} else if lcpIndex < pathLength {
			path = path[lcpIndex:]
			next := n.findMaxLengthChild(path, static)
			if next != nil {
				n = next
				continue
			}

			nn := newNode(nil, nil, path, route.originalPath, static, method, n)
			if route.handler != nil {
				nn.methods[method] = route.handler
			}
			n.addChild(nn)
		} else {
			if route.handler != nil {
				n.methods[method] = route.handler
				n.originalPath = route.originalPath
			}
		}
		return
	}
}

func (r *Router) Search(method, path string) (HandleFunc, []*param) {
	// search root
	if path == "/" {
		return r.tree.methods[method], nil
	}

	current := r.tree
	searchIndex := 0

	for {
		prefixLength := 0

		pathLength := len(path)
		prefixLength = len(current.prefix)

		minCommonLength := prefixLength
		if pathLength < minCommonLength {
			minCommonLength = pathLength
		}

		lcpIndex := 0
		for ; lcpIndex < minCommonLength; lcpIndex++ {
			if path[lcpIndex] != current.prefix[lcpIndex] {
				break
			}
		}

		if lcpIndex != prefixLength {
			// backtrack
		}

		path = path[lcpIndex:]
		searchIndex = searchIndex + lcpIndex

		if path == "" {
			if h, ok := current.methods[method]; ok {
				return h, nil
			}

			return nil, nil
		}

		if path != "" {
			if child := current.findMaxLengthChild(path, static); child != nil {
				current = child
				continue
			}
		}

		if child := current.findMaxLengthChild(path, pathParam); child != nil {
			i := 0
			for ; i < len(path) && path[i] != '/'; i++ {
			}
			param := path[:i]
			path = path[i:]
			fmt.Println(param)
			continue
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
