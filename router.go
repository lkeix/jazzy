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
		handler     HandleFunc
		prefix      string
		handleType  int
		methods     []string
		params      []param
		parent      *node
		children    []*node
	}

	RouterRepo interface {
		Insert(method string, path string, handler HandleFunc)
		Search(method string, path string) HandleFunc
	}

	Router struct {
		tree *node
	}
)

func NewRouter() RouterRepo {
	return &Router{
		tree: &node{},
	}
}

func (r *Router) Insert(method, path string, handler HandleFunc) {
	// root insert
	node := r.tree
	if len(path) == 0 {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	node.prefix = path
	node.handler = handler
	node.methods = append(node.methods, method)
}

func (r *Router) Search(method, path string) HandleFunc {
	// search root
	node := r.tree
	if path == node.prefix {
		return node.handler
	}
	return nil
}
