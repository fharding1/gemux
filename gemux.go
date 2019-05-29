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
	// the request URL. If NotFoundHandler is nil, http.NotFoundHandler
	// will be used.
	NotFoundHandler http.Handler

	// MethodNotAllowedHandler is called when there is no method corrosponding
	// to the request URL. If MethodNotAllowedHandler is nil, MethodNotAllowedHandler
	// will be used.
	MethodNotAllowedHandler http.Handler
}

// ServeHTTP dispatches the request to the handler whose pattern and method
// matches the request URL and method.
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	current := mux

	for head, tail := shiftPath(r.URL.Path); head != ""; head, tail = shiftPath(tail) {
		if current.wildcardChild != nil {
			r = r.WithContext(appendPathParameter(r.Context(), head))
			current = current.wildcardChild
			continue
		}

		child, ok := current.children[head]
		if !ok {
			current.notFoundHandler().ServeHTTP(w, r)
			return
		}

		current = child
	}

	current.serveHandler(w, r)
}

// notFoundHandler returns the mux NotFoundHandler if there is one, otherwise
// http.NotFoundHandler.
func (mux *ServeMux) notFoundHandler() http.Handler {
	if mux.NotFoundHandler != nil {
		return mux.NotFoundHandler
	}

	return http.NotFoundHandler()
}

// serveHandler serves the request to the proper method handler, or calls the
// 404 or 405 handler.
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

// Handle registers a handler for the given pattern and method on the muxer.
// The pattern should be the exact URL to match, with the exception of wildcards
// ("*"), which can be used for a single segment of a path (split on "/") to match
// anything. A wildcard method of "*" can also be used to match any method.
func (mux *ServeMux) Handle(pattern string, method string, handler http.Handler) {
	if pattern == "/" {
		if mux.handlers == nil {
			mux.handlers = make(map[string]http.Handler)
		}

		if method == "*" {
			mux.wildcardHandler = handler
		} else {
			mux.handlers[method] = handler
		}

		return
	}

	head, tail := shiftPath(pattern)

	if head == "*" {
		if mux.wildcardChild == nil {
			mux.wildcardChild = mux.newChild()
		}

		mux.wildcardChild.Handle(tail, method, handler)
		return
	}

	if mux.children == nil {
		mux.children = make(map[string]*ServeMux)
	}

	if mux.children[head] == nil {
		mux.children[head] = mux.newChild()
	}

	mux.children[head].Handle(tail, method, handler)
}

// newChild returns a pointer to a new ServeMux with NotFoundHandler
// and MethodNotAllowedHandler set to the parent mux values.
func (mux *ServeMux) newChild() *ServeMux {
	return &ServeMux{
		MethodNotAllowedHandler: mux.MethodNotAllowedHandler,
		NotFoundHandler:         mux.NotFoundHandler,
	}
}

// PathParameter returns the nth path parameter from the request
// context. It returns an empty string if no value exists at the
// given index.
func PathParameter(ctx context.Context, n int) string {
	contextValue := ctx.Value(pathParametersKey)
	if contextValue == nil {
		return ""
	}

	pathParameters, ok := contextValue.([]string)
	if !ok {
		return ""
	}

	if n < 0 || n >= len(pathParameters) {
		return ""
	}

	return pathParameters[n]
}

// MethodNotAllowedHandler returns a simple request handler that replies to
// each request with a "405 method not allowed" reply and writes the 405 status
// code.
func MethodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
	})
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

// appendPathParameter pushes a path parameter to the given context.
func appendPathParameter(ctx context.Context, pathParameter string) context.Context {
	var pathParameters []string

	if contextValue := ctx.Value(pathParametersKey); contextValue != nil {
		value, ok := contextValue.([]string)
		if ok {
			pathParameters = value
		}
	}

	return context.WithValue(ctx, pathParametersKey, append(pathParameters, pathParameter))
}
