# Good Enough Multiplexer (gemux)

[![pkg.go.dev](https://img.shields.io/badge/go.dev-pkg-blue)](https://pkg.go.dev/github.com/fharding1/gemux?tab=doc)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`gemux` is the good enough multiplexer. It aims to provide functionality that is good enough for the majority of HTTP services,
with a focus on a small and easy to test codebase, fair performance, and no dependencies outside the standard library.

## Usage

```go
package main

func main() {
    mux := new(gemux.ServeMux)

    mux.Handle("/", http.MethodGet, http.HandlerFunc(healthHandler))
    mux.Handle("/posts", http.MethodGet, http.HandlerFunc(getPostsHandler))
    mux.Handle("/posts", http.MethodPost, http.HandlerFunc(createPostHandler))
    mux.Handle("/posts/*", http.MethodDelete, http.HandlerFunc(deletePostHandler))
    mux.Handle("/posts/*/comments", http.MethodPost, http.HandlerFunc(createCommentHandler))
    mux.Handle("/posts/*/comments", http.MethodGet, http.HandlerFunc(getCommentsHandler))
    mux.Handle("/posts/*/comments/*", http.MethodDelete, http.HandlerFunc(deleteCommentHandler))

    http.ListenAndServe(":8080", mux)
}
```

## Features

### Strict Path Based Routing (with wildcards)

Route strictly based on paths, but allow wildcards for path parameters such as a resource ID.

```go
mux.Handle("/users", http.MethodPost, http.HandlerFunc(getUsersHandler))
mux.Handle("/posts", http.MethodPost, http.HandlerFunc(createPostHandler))
mux.Handle("/posts/*", http.MethodGet, http.HandlerFunc(getPostHandler))
```

### Strict Method Based Routing (with wildcards)

Route based on methods, and allow wildcard methods if you need to write your own method multiplexer, or want
to match on any method.

```go
mux.Handle("/users", http.MethodGet, http.HandlerFunc(createPostHandler)) // implement your own method muxer
mux.Handle("/posts", "*", http.HandlerFunc(createPostHandler)) // implement your own method muxer
```

### Context Path Parameters

Extract path wildcard values via the request context.

```go
func getPostHandler(w http.ResponseWriter, r *http.Request) {
    rawPostID := gemux.PathParameter(r.Context(), 0)
    postID, err := strconv.ParseInt(rawPostID, 10, 64)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // ...
}
```

### Custom Error Handlers

Create custom error handlers for when a route or method isn't found.

```go
mux := new(gemux.ServeMux)
mux.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotFound)
    json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
})
mux.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusMethodNotAllowed)
    json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
})
```

## Benchmarks

Performed on a Dell XPS 13 with an Intel(R) Core(TM) i7-8550U CPU @ 1.80GHz. `gemux` is fast enough
that it's performance impact is negligible for most HTTP services.

```
goos: linux
goarch: amd64
pkg: github.com/fharding1/gemux
BenchmarkServeHTTP/one_static_path-8         	 7071818	       159 ns/op
BenchmarkServeHTTP/one_wildcard_path-8       	 2432485	       516 ns/op
BenchmarkServeHTTP/one_wildcard_path_and_method-8         	 2541954	       453 ns/op
BenchmarkServeHTTP/short_path_with_many_routes-8          	 6360070	       187 ns/op
BenchmarkServeHTTP/very_deep_static_path-8                	 1856644	       644 ns/op
BenchmarkServeHTTP/very_deep_wildcard_path-8              	  499780	      2262 ns/op
BenchmarkHandle/one_static_path-8                         	 6987087	       170 ns/op
BenchmarkHandle/one_wildcard_path-8                       	 8136362	       151 ns/op
BenchmarkHandle/one_wildcard_path_and_method-8            	 8723088	       139 ns/op
BenchmarkHandle/short_path_with_many_routes-8             	  191238	      6234 ns/op
BenchmarkHandle/very_deep_static_path-8                   	 1626010	       694 ns/op
BenchmarkHandle/very_deep_wildcard_path-8                 	 2015136	       616 ns/op
PASS
ok  	github.com/fharding1/gemux	18.140s
```