package jazzy

const (
	static = iota
	pathParam
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
			handlers: []HandleFunc{},
			methods:  []string{},
		},
	}
}

func (r *Router) Insert(method, path string, handler HandleFunc) {
	// root insert
	n := r.tree
	if len(path) == 0 {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	n.prefix = path
	n.handlers = append(n.handlers, handler)
	n.methods = append(n.methods, method)
}

func (r *Router) Search(method, path string) HandleFunc {
	// search root
	n := r.tree
	for i, m := range n.methods {
		if m == method {
			return n.handlers[i]
		}
	}
	return nil
}
