package jazzy

import (
	"bytes"
	"fmt"
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
		handlers    []HandleFunc
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
			handlers: []HandleFunc{},
			methods:  []string{},
		},
	}
}

func (r *Router) Insert(method, path string, handler HandleFunc) {
	n := r.tree
	if len(path) == 0 {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}

	var _n *node
	var l int
	suffix := path
	now := ""

	for {

		if _n != nil {
			n = _n
		}

		n, l = lcpMinChild(n, suffix)

		// root
		if n.prefix == "" {
			nn := newNode(
				[]HandleFunc{},
				handler,
				suffix,
				now,
				static,
				method,
				n,
			)

			n.children = append(n.children, nn)
			fmt.Printf("inserted: %v, after: %v\n", nn, n)
			break
		}

		// update node
		if l < len(suffix) {
			children := lcpMinChildren(n.children, suffix[l:])

			// create intermediate node
			in := newNode(
				[]HandleFunc{},
				nil,
				suffix[l:],
				now+n.prefix,
				static,
				method,
				n,
			)

			for _, child := range children {
				child.prefix = child.prefix[l:]
				child.parent = in
				n.children = remove(n.children, child)
			}

			if len(children) != 0 {
				n.children = append(n.children, in)
				n = in
			}
		}

		// create a new static node
		if l == len(suffix) {
			nn := newNode(
				[]HandleFunc{},
				handler,
				suffix,
				now,
				static,
				method,
				n,
			)
			n.children = append(n.children, nn)
			fmt.Printf("inserted: %v, after: %v\n", nn, n)
			break
		}

		suffix = suffix[l:]
		now += n.prefix
		_n = n
	}
}

func switching(n, nn *node, now, suffix string, l int) {
	pn := n.parent
	ns := lcpMinChildren(pn.children, suffix[:l])
	l = lcp(n.prefix, suffix)

	if len(ns) != 0 {
		for i := 0; i < len(ns); i++ {
			ns[i].prefix = ns[i].prefix[l:]
			ns[i].path = ns[i].path[l:]
			ns[i].parent = nn
			nn.parent = pn
			nn.children = append(nn.children, ns[i])
			pn.children = remove(pn.children, ns[i])
			fmt.Printf("switched %v after %v\n", n, nn)
		}

		n = pn

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
	handleType int,
	method string,
	parent *node,
) *node {
	return &node{
		middlewares: middlewares,
		handlers:    []HandleFunc{handler},
		prefix:      prefix,
		path:        path,
		handleType:  handleType,
		methods:     []string{method},
		parent:      parent,
	}
}

func (r *Router) Search(method, path string) (HandleFunc, []*param) {
	// search root
	n := r.tree

	if len(path) == 0 {
		path = "/"
	}

	if path[0] != '/' {
		path = "/" + path
	}

	handler, _n := staticRouting(n, path, method)

	if handler != nil {
		return handler, nil
	}

	fmt.Printf("returned: %v\n", _n)
	_n = backtrack(_n)
	fmt.Printf("backtracked: %v\n", _n)

	handler, params := paramRouting(_n, path, method)

	/*
		if handler != nil {
			return handler, nil
		}
	*/

	return handler, params
}

func backtrack(n *node) *node {
	var _n *node

	for {
		if _n != nil {
			n = _n
		}

		if n.parent == nil {
			return n
		}

		child := paramChild(n.children)

		if child != nil {
			return n
		}

		_n = n.parent
	}

	return _n
}

func staticRouting(n *node, path, method string) (HandleFunc, *node) {
	suffix := path

	now := ""

	var _n *node
	var l int

	for {
		if _n != nil {
			n = _n
		}

		if now == path || now == path+"/" {
			i := handleindex(n, method)
			return n.handlers[i], nil
		}

		l = len(path)

		for i := 0; i < len(n.children); i++ {
			if n.children[i].handleType != pathParam {
				mn := lcp(n.children[i].prefix, suffix)
				fmt.Printf("%s, %s: %d\n", n.children[i].prefix, suffix, mn)
				if l >= mn && mn != 0 {
					l = mn
					_n = n.children[i]
				}
			}
		}

		if _n == n {
			return nil, _n
		}

		now += _n.prefix
		suffix = suffix[l:]
	}
	return nil, _n
}

func paramRouting(n *node, path, method string) (HandleFunc, []*param) {
	l := lcp(n.path, path)
	suffix := path[l:]
	prev := ""
	now := ""

	if now != "/" {
		now += "/"
	}

	now = n.path

	var _n *node

	params := make([]*param, 0)
	for {
		if _n != nil {
			n = _n
		}

		child := paramChild(n.children)

		if now == path || now == path+"/" {
			i := handleindex(n, method)
			return n.handlers[i], params
		}

		if len(suffix) == 0 {
			return nil, nil
		}

		if n.handleType == pathParam {
			i := 0
			pp := ""

			if len(suffix) == 0 {
				continue
			}

			if suffix[0] == '/' {
				i += 1
			}

			for ; i < len(suffix); i++ {
				if suffix[i] == '/' {
					break
				}
				pp += string(suffix[i])
			}

			_n = child
			if child.param != nil {
				child.param.value = pp
				params = append(params, child.param)
			}

			if now != "/" {
				now += "/"
			}

			now += pp
			suffix = suffix[i:]

			if len(child.children) == 0 {
				_n = child
			}
			continue
		}

		n, l = lcpMinChild(n, suffix)
		now += n.path
		prefix := suffix[:l]
		suffix = suffix[l:]

		if prev == prefix {
			return nil, nil
		}

		prev = prefix
		_n = n
	}
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

func handleindex(n *node, method string) int {
	for i := 0; i < len(n.methods); i++ {
		if n.methods[i] == method {
			return i
		}
	}
	return -1
}
