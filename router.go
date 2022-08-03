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
		params      []*param
		parent      *node
		children    []*node
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

	// add static route
	suffix := path
	htype := handleType(path)

	var _n *node
	for {
		if _n != nil {
			n = _n
		}
		n, l := lcpMinChild(n, suffix)

		if len(suffix) == 0 {
			break
		}
		// path param
		if suffix[0] == colon || suffix[min(1, len(suffix)-1)] == colon {
			for i := 0; i < len(n.children); i++ {
				if strings.Contains(n.children[i].prefix, ":") {
					panic("path param handler is conflicted")
				}
			}

			i := 1

			for ; i < len(suffix); i++ {
				if suffix[i] == '/' {
					break
				}
			}

			nn := newNode(
				[]HandleFunc{},
				nil,
				":/",
				pathParam,
				method,
				n,
			)

			nn.params = append(nn.params, &param{
				key:   suffix[1:i],
				value: "",
			})

			n.children = append(n.children, nn)

			fmt.Printf("inserted %v after %v\n", nn, n)
			if i == len(suffix) {
				fmt.Println(suffix)
				break
			}

			suffix = suffix[i:]
			_n = nn

			continue
		}

		/*
		   fmt.Printf("now node: %v\n", n)
		   fmt.Printf("len(suffix): %d\n", len(suffix))
		   fmt.Printf("l: %d\n", l)
		   fmt.Printf("suffix[:l]: %v\n", suffix[:l])
		*/
		// a node have children, suffix is left, and don't exist intermediate node
		if len(suffix) != l && suffix[:l] != "/" && n.prefix != suffix[:l] {
			// create intermediate node
			nn := newNode(
				[]HandleFunc{},
				nil,
				suffix[:l],
				none,
				method,
				n.parent,
			)

			pn := n.parent
			ns := lcpMinChildren(pn.children, suffix[:l])

			for i := 0; i < len(ns); i++ {
				ns[i].prefix = ns[i].prefix[l:]
				nn.children = append(nn.children, ns[i])
				pn.children = remove(pn.children, ns[i])
				fmt.Printf("switched %v after %v\n", n, nn)
			}

			pn.children = append(pn.children, nn)

			suffix = suffix[l:]
			_n = nn
			continue
		}

		// create a new node
		if len(suffix) == l {
			nn := newNode(
				[]HandleFunc{},
				handler,
				suffixSlash(suffix[:l]),
				htype,
				method,
				n,
			)

			if n.parent != nil {
				pn := n.parent
				ns := lcpMinChildren(pn.children, suffix[:l])
				l := lcp(n.prefix, suffix)

				if len(ns) != 0 && l != 0 {
					for i := 0; i < len(ns); i++ {
						ns[i].prefix = ns[i].prefix[l:]
						nn.parent = pn
						nn.children = append(nn.children, ns[i])
						pn.children = remove(pn.children, ns[i])
						fmt.Printf("switched %v after %v\n", n, nn)
					}

					n = pn
				}
			}

			fmt.Printf("now node: %v\n", n)
			n.children = append(n.children, nn)
			fmt.Printf("inserted %v after %v\n", nn, n)
			break
		}

		suffix = suffix[l:]
		_n = n
	}

	// if path is perfect match
	// n.handlers = append(n.handlers, handler)
	// n.methods = append(n.methods, method)
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

func newNode(middlewares []HandleFunc, handler HandleFunc, prefix string, handleType int, method string, parent *node) *node {
	return &node{
		middlewares: middlewares,
		handlers:    []HandleFunc{handler},
		prefix:      prefix,
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

	handler := staticRouting(n, path, method)

	if handler != nil {
		return handler, nil
	}
	return nil, nil
}

func staticRouting(n *node, path, method string) HandleFunc {
	suffix := path

	prev := ""
	now := ""

	var _n *node
	var l int

	for {
		if _n != nil {
			n = _n
		}

		_n, l = lcpMinChild(n, suffix)

		fmt.Println(n)
		now += n.prefix
		prefix := suffix[:l]
		suffix = suffix[l:]

		if now == path || now == path+"/" {
			i := handleindex(n, method)
			return n.handlers[i]
		}

		if prev == prefix {
			return nil
		}

		prev = prefix
	}
	return nil
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

func min[T int | int32 | int64 | float32 | float64](x, y T) T {
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
