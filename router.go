package jazzy

import (
	"fmt"
	"strings"
	"time"
)

const (
	static = iota
	pathParam
  none

	colon = ":"
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
		params      []param
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

  var _n *node
	for {
    if _n != nil {
      n = _n
    }
		n, l := lcpMinChild(n, suffix)

    time.Sleep(time.Second * 1)

    // if next node doesn't exist
		if len(suffix) == l {
			nn := newNode(
				[]HandleFunc{},
				handler,
				suffix,
				handleType(path),
				method,
				n,
			)
			n.children = append(n.children, nn)
      fmt.Printf("inserted %v after %v\n", nn, n)
			break
		}

    // a node have children, suffix is left
    if len(suffix) != l && suffix[:l] != "/" {
      // create new node
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
      }

      pn.children = append(pn.children, nn)

      fmt.Println(nn)
      suffix = suffix[l:]
      _n = nn
      continue
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

func (r *Router) Search(method, path string) HandleFunc {
	// search root
	n := r.tree

	if len(path) == 0 {
		path = "/"
	}

	if path[0] != '/' {
		path = "/" + path
	}

	suffix := path

	prev := ""
	now := ""

	var _next *node
	var l int

	for {
		if _next != nil {
			n = _next
		}

		_next, l = lcpMinChild(n, suffix)

		now += n.prefix
		prefix := suffix[:l]
		suffix = suffix[l:]

		if now == path {
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

func lcpMinChildren(ns []*node, suffix string) ([]*node) {
  rns := make([]*node, 0, len(ns))
  l := len(suffix)
  for i := 0; i < len(ns); i++ {
    if ns[i].prefix[:l] == suffix {
      rns = append(rns, ns[i])
    }
  }
  return rns
}

func remove(ns []*node, n*node) []*node {
  for i := 0; i < len(ns); i++ {
    if ns[i] == n {
      return append(ns[:i], ns[i+1:]...)
    }
  }
  return ns
}

func handleType(path string) int {
	if strings.Contains(path, colon) {
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

func handleindex(n *node, method string) int {
	for i := 0; i < len(n.methods); i++ {
		if n.methods[i] == method {
			return i
		}
	}
	return -1
}
