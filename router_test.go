package jazzy

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestInsert(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		searchPath string
		handler    HandleFunc
	}{
		{
			name:       "insert root handler",
			method:     http.MethodGet,
			path:       "/",
			searchPath: "/",
			handler: func(ctx *Context) {

			},
		},
		{
			name:       "insert simple handler",
			method:     http.MethodGet,
			path:       "/hoge",
			searchPath: "/hoge",
			handler: func(ctx *Context) {

			},
		},
		{
			name:       "insert partial match handler",
			method:     http.MethodGet,
			path:       "/hog/fuga",
			searchPath: "/hog/fuga",
			handler: func(ctx *Context) {

			},
		},
		{
			name:       "insert 2nested simple handler",
			method:     http.MethodGet,
			path:       "/hoge/fuga",
			searchPath: "/hoge/fuga",
			handler: func(ctx *Context) {

			},
		},
		{
			name:       "insert path param handler",
			method:     http.MethodGet,
			path:       "/:hoge",
			searchPath: "/aaaa",
			handler: func(ctx *Context) {

			},
		},
    {

			name:       "insert 2nested path param handler",
			method:     http.MethodGet,
			path:       "/:hoge/hoge",
			searchPath: "/aaaa/hoge",
			handler: func(ctx *Context) {

			},
    }
	}

	r := NewRouter()

	for _, tt := range tests {
		r.Insert(tt.method, tt.path, tt.handler)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := r.Search(tt.method, tt.searchPath)

			if reflect.ValueOf(tt.handler).Pointer() != reflect.ValueOf(f).Pointer() {
				t.Errorf("fail: handler isn't same\n\texpect: %v\n\tactual:%v\n", tt.handler, f)
			}
		})
	}
}
