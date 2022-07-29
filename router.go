package jazzy

import (
	"fmt"
	"strings"
	"time"
)

const (
	static = iota
	pathParam

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

  for {
    n, l := lcpMinChildren(n, suffix)
    time.Sleep(time.Second * 1)

    // if next node doesn't exist
    if len(n.children) == 0 || len(suffix) == l {
      fmt.Println(suf(suffix, l))
      nn := newNode(
        []HandleFunc{},
        handler,
        suf(suffix, l),
        handleType(path),
        method,
        n,
      )
      n.children = append(n.children, nn)
      return
    }
    suffix = suffix[l:]
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
  return &node {
    middlewares: middlewares,
    handlers: []HandleFunc{handler},
    prefix: prefix,
    handleType: handleType,
    methods: []string{method},
    parent: parent,
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
  l := 0
  for {
    if _next != nil {
      n = _next
    }
    _next, l = lcpMinChildren(n, suffix) 
    
    now += n.prefix
    prefix := suffix[:l]
    suffix = suffix[l:]

    if now == path {
      i := handleindex(n, method)
      return n.handlers[i]
    }
    if prev == prefix {
      fmt.Println("search end")
      return nil
    }
    prev = prefix
  }
	return nil
}

func lcpMinChildren(n *node, suffix string) (*node, int) {
  mn := len(suffix)
  next := n

  for _, child := range n.children {
    l := lcp(child.prefix, suffix)
    if l <= mn {
      mn = l
      next = child
    }
  }
  return next, mn
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
