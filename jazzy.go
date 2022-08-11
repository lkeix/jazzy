package jazzy

import (
	"net/http"
	"sync"
)

type (
	HandleFunc func(*Context)

	JazzyInterface interface {
		GET(string, HandleFunc)
		POST(string, HandleFunc)
		PUT(string, HandleFunc)
		DELETE(string, HandleFunc)
		PATCH(string, HandleFunc)
		OPTIONS(string, HandleFunc)
		Group(string) *Jazzy
		Serve(string)
	}

	Jazzy struct {
		prefix string
		pool   sync.Pool
		Router *Router
	}
)

const (
	notfound = "{ \"message\": \"not found\" }"
)

func New() JazzyInterface {
	pool := sync.Pool{
		New: func() interface{} {
			var w http.ResponseWriter
			r := &http.Request{}
			return NewContext(r, w)
		},
	}
	return &Jazzy{
		pool:   pool,
		Router: NewRouter(),
	}
}

func (jazz *Jazzy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := jazz.pool.Get().(*Context)
	ctx.Init(r, w)
	h, params := jazz.Router.Search(ctx.Request.Method, ctx.Request.URL.Path)
	if h != nil {
		ctx.params = params
		h(ctx)
	}

	if h == nil {
		noRoute(ctx)
	}

	jazz.pool.Put(ctx)
}

func (jazz *Jazzy) Serve(port string) {
	if port[0] != ':' {
		port = ":" + port
	}
	server := http.Server{
		Handler: jazz,
		Addr:    port,
	}
	server.ListenAndServe()
}

func (jazz *Jazzy) GET(path string, handler HandleFunc) {
	jazz.Router.Insert(http.MethodGet, jazz.prefix+path, handler)
}

func (jazz *Jazzy) POST(path string, handler HandleFunc) {
	jazz.Router.Insert(http.MethodPost, jazz.prefix+path, handler)
}

func (jazz *Jazzy) PUT(path string, handler HandleFunc) {
	jazz.Router.Insert(http.MethodPut, jazz.prefix+path, handler)
}

func (jazz *Jazzy) DELETE(path string, handler HandleFunc) {
	jazz.Router.Insert(http.MethodDelete, jazz.prefix+path, handler)
}

func (jazz *Jazzy) PATCH(path string, handler HandleFunc) {
	jazz.Router.Insert(http.MethodPatch, jazz.prefix+path, handler)
}

func (jazz *Jazzy) OPTIONS(path string, handler HandleFunc) {
	jazz.Router.Insert(http.MethodOptions, jazz.prefix+path, handler)
}

func (jazz *Jazzy) Group(path string) *Jazzy {
	jaz := new(Jazzy)
	*jaz = *jazz
	jaz.Router.Insert("GROUP", path, nil)
	return jaz
}

func noRoute(ctx *Context) {
	ctx.Writer.Write(([]byte)(notfound))
}
