package jazzy

import (
	"net/http"
)

type (
	Context struct {
		Response http.Response
		Request  *http.Request
		Writer   http.ResponseWriter
		Bind     (interface{})
		params   []*param
	}
)

func NewContext(r *http.Request, w http.ResponseWriter) *Context {
	return &Context{
		Writer:  w,
		Request: r,
	}
}

func (ctx *Context) Init(r *http.Request, w http.ResponseWriter) {
	ctx.Request = r
	ctx.Writer = w
}

func (ctx *Context) Param(key string) string {
	for i := 0; i < len(ctx.params); i++ {
		if ctx.params[i].key == key {
			return ctx.params[i].value
		}
	}
	return ""
}
