package gemux

import (
	"context"
	"net/http"
	"path"
	"strings"
)

// ServeMux is an HTTP request multiplexer. It matches the URL and method of the incoming
// request against a list of registered routes, and calls the matching route.
type ServeMux struct {
	handlers        map[string]http.Handler // methods describe actions on a resource
	wildcardHandler http.Handler            // * method
	children        map[string]*ServeMux    // paths describe resources
	wildcardChild   *ServeMux               // * path

	// NotFoundHandler is called when there is no path corrosponding to
	// the request URL. If NotFoundHandler is nil, `http.NotFoundHandler`
	// will be used.
	NotFoundHandler http.Handler

	// MethodNotAllowedHandler is called when there is no method corrosponding
	// to the request URL. If MethodNotAllowedHandler is nil, `MethodNotAllowedHandler`
	// will be used.
	MethodNotAllowedHandler http.Handler
}

// ServeHTTP dispatches the request to the handler whose pattern and method
// matches the request URL and method.
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)

	if head == "" {
		mux.serveHandler(w, r)
		return
	}

	if mux.wildcardChild != nil {
		ctx := appendPathParameter(r.Context(), head)
		mux.wildcardChild.ServeHTTP(w, r.WithContext(ctx))
		return
	}

	child, ok := mux.children[head]
	if !ok {
		mux.notFoundHandler().ServeHTTP(w, r)
		return
	}

	child.ServeHTTP(w, r)
}

// notFoundHandler returns the mux NotFoundHandler if there is one, otherwise
// http.NotFoundHandler.
func (mux *ServeMux) notFoundHandler() http.Handler {
	if mux.NotFoundHandler != nil {
		return mux.NotFoundHandler
	}

	return http.NotFoundHandler()
}

func (mux *ServeMux) serveHandler(w http.ResponseWriter, r *http.Request) {
	if mux.handlers == nil {
		mux.notFoundHandler().ServeHTTP(w, r)
		return
	}

	if mux.wildcardHandler != nil {
		mux.wildcardHandler.ServeHTTP(w, r)
		return
	}

	handler, ok := mux.handlers[r.Method]
	if !ok {
		mux.methodNotAllowedHandler().ServeHTTP(w, r)
		return
	}

	handler.ServeHTTP(w, r)
}

// methodNotAllowedHandler returns the mux MethodNotAllowedHandler if there is one, otherwise
// MethodNotAllowedHandler.
func (mux *ServeMux) methodNotAllowedHandler() http.Handler {
	if mux.MethodNotAllowedHandler != nil {
		return mux.MethodNotAllowedHandler
	}

	return MethodNotAllowedHandler()
}

// Handle registers a handler for the given pattern and method. If a handler matching the
// pattern and method already exists, Handle panics.
func (mux *ServeMux) Handle(pattern string, method string, handler http.Handler) {
	if pattern == "/" {
		mux.handleRoot(pattern, method, handler)

		return
	}

	if mux.children == nil {
		mux.children = make(map[string]*ServeMux)
	}

	head, tail := shiftPath(pattern)

	if head == "*" {
		if mux.wildcardChild == nil {
			mux.wildcardChild = &ServeMux{
				MethodNotAllowedHandler: mux.MethodNotAllowedHandler,
				NotFoundHandler:         mux.NotFoundHandler,
			}
		}

		mux.wildcardChild.Handle(tail, method, handler)
		return
	}

	if mux.children[head] == nil {
		mux.children[head] = &ServeMux{
			MethodNotAllowedHandler: mux.MethodNotAllowedHandler,
			NotFoundHandler:         mux.NotFoundHandler,
		}
	}

	mux.children[head].Handle(tail, method, handler)
}

func (mux *ServeMux) handleRoot(pattern string, method string, handler http.Handler) {
	if mux.handlers == nil {
		mux.handlers = make(map[string]http.Handler)
	}

	if method == "*" {
		if mux.wildcardHandler != nil {
			panic("wildcard handler already exists")
		}

		mux.wildcardHandler = handler
	} else {
		if _, ok := mux.handlers[method]; ok {
			panic("handler for that method already exists")
		}

		mux.handlers[method] = handler
	}
}

// PathParameters returns all the path parameters that were set when
// ServeHTTP was called on a request.
func PathParameters(ctx context.Context) []string {
	pathParameters := ctx.Value(pathParametersKey)
	if pathParameters == nil {
		return nil
	}

	return pathParameters.([]string)
}

// shiftPath is (ironically) stolen from
// https://blog.merovius.de/2017/06/18/how-not-to-use-an-http-router.html
// and is the fundamental building block for this entire library.
func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

type contextKey int

const (
	pathParametersKey contextKey = iota
)

func appendPathParameter(ctx context.Context, pathParameter string) context.Context {
	ctxPathParameters := ctx.Value(pathParametersKey)
	if ctxPathParameters == nil {
		ctxPathParameters = []string{}
	}

	pathParameters := append(ctxPathParameters.([]string), pathParameter)

	return context.WithValue(ctx, pathParametersKey, pathParameters)
}