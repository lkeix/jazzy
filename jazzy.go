package jazzy

import (
	"net/http"
	"sync"
)

type (
	HandleFunc func(*Context)

	JazzyRepo interface {
		GET(string, HandleFunc)
		POST(string, HandleFunc)
		PUT(string, HandleFunc)
		DELETE(string, HandleFunc)
		PATCH(string, HandleFunc)
		OPTIONS(string, HandleFunc)
		Serve(string)
	}

	Jazzy struct {
		pool   sync.Pool
		Router Router
	}
)

func New() JazzyRepo {
	pool := sync.Pool{
		New: func() interface{} {
			var w http.ResponseWriter
			r := &http.Request{}
			return NewContext(r, w)
		},
	}
	return &Jazzy{
		pool: pool,
	}
}

func (jazz *Jazzy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := jazz.pool.Get().(*Context)
	ctx.Init(r, w)
	h := jazz.Router.Search(ctx.Request.Method, ctx.Request.URL.Path)
	h(ctx)
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
	jazz.Router.Insert(http.MethodGet, path, handler)
}

func (jazz *Jazzy) POST(path string, handler HandleFunc) {
	jazz.Router.Insert(http.MethodPost, path, handler)
}

func (jazz *Jazzy) PUT(path string, handler HandleFunc) {
	jazz.Router.Insert(http.MethodPut, path, handler)
}

func (jazz *Jazzy) DELETE(path string, handler HandleFunc) {
	jazz.Router.Insert(http.MethodDelete, path, handler)
}

func (jazz *Jazzy) PATCH(path string, handler HandleFunc) {
	jazz.Router.Insert(http.MethodPatch, path, handler)
}

func (jazz *Jazzy) OPTIONS(path string, handler HandleFunc) {
	jazz.Router.Insert(http.MethodOptions, path, handler)
}
