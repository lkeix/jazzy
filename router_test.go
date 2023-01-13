package jazzy

import (
	"net/http"
	"reflect"
	"testing"
)

func TestInsert(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		path    string
		handler HandleFunc
	}{
		{
			name:   "insert simple handler",
			method: http.MethodGet,
			path:   "/",
			handler: func(ctx *Context) {

			},
		},
		{
			name:   "insert simple handler",
			method: http.MethodGet,
			path:   "/hoge",
			handler: func(ctx *Context) {

			},
		},
		{
			name:   "insert partial match handler",
			method: http.MethodGet,
			path:   "/hog/fuga",
			handler: func(ctx *Context) {

			},
		},
		{
			name:   "insert 2nested simple handler",
			method: http.MethodGet,
			path:   "/hoge/fuga",
			handler: func(ctx *Context) {

			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			r.Insert(tt.method, tt.path, tt.handler)

			f, _ := r.Search(tt.method, tt.path)

			if reflect.ValueOf(tt.handler).Pointer() != reflect.ValueOf(f).Pointer() {
				t.Errorf("fail: handler isn't same\n\texpect: %v\n\tactual:%v\n", tt.handler, f)
			}
		})
	}
}
