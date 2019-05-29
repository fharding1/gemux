# Good Enough Multiplexer (gemux)

[![Godoc](https://godoc.org/github.com/fharding1/gemux?status.svg)](http://godoc.org/github.com/fharding1/gemux)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CircleCI](https://circleci.com/gh/fharding1/gemux.svg?style=svg)](https://circleci.com/gh/fharding1/gemux)

`gemux` is the good enough multiplexer. It aims to provide functionality that is good enough for the majority of HTTP services.

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

## Rationale

In every single HTTP API I write:

* I want to multiplex requests based on the path and the method of the request
* I want to be able to capture path parameters (e.g. capture `42` from `/posts/42`)

So this is all the functionality `gemux` provides. This is good enough for almost all HTTP APIs. If it isn't good enough for yours, you should use another router, such as [gorilla/mux](https://github.com/gorilla/mux).

## Features

### Strict Path Based Routing (with wildcards)

Route strictly based on paths, but allow wildcards for path parameters such as a resource ID.

```go
mux.Handle("/users", http.MethodPost, http.HandlerFunc(getUsersHandler))
mux.Handle("/posts", http.MethodPost, http.HandlerFunc(createPostHandler))
mux.Handle("/posts/*", http.MethodGet, http.HandlerFunc(getPostHandler))
```

### Strict Method Based Routing (with wildcards)

Route based on methods, and allow wildcard methods if you need to write your own method multiplexer.

```go
mux.Handle("/users", http.MethodGet, http.HandlerFunc(createPostHandler)) // implement your own method muxer
mux.Handle("/posts", "*", http.HandlerFunc(createPostHandler)) // implement your own method muxer
```

### Context Path Parameters

Extract path wildcard values via the request context.

```go
func getPostHandler(w http.ResponseWriter, r *http.Request) {
    pathParameters := gemux.PathParameters(r.Context())
    if len(pathParameters) != 1 {
        http.Error(w, "got an unexpected number of path parameters, the muxer is broken", http.StatusInternalServerError)
        return
    }

    postID, err := strconv.ParseInt(pathParameters[0], 10, 64)
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

### Very Small API

`gemux` has a very small exported API, and not much "magic" within the `Handle` function (such as named path parameters or regular expressions) which means it's easier to learn than other multiplexer libraries.

### Just Standard Library Dependencies

`gemux` does not depend upon any libraries outside of the standard library.

### It's Fast... Enough.

`gemux` isn't slow, and it's not fast either. If you're writing an HTTP service you're probably doing a bunch of I/O, so your multiplexer speed is almost totally insignificant. Since `gemux` doesn't try to do anything clever to be fast ([the root of all evil](http://wiki.c2.com/?PrematureOptimization)) it's more maintainable than other multiplexers.

### Sticks to the Standard

`gemux` handlers are just `http.Handler`s so you can easily plug in other middlewares without writing wrapper functions. `gemux` also recognizes that HTTP methods are ultimately just strings, so it doesn't do anything weird like provide helper functions named after methods (e.g. `router.GET()`). You can use `http.MethodGet` if you need a GET method.