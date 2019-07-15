# Good Enough Multiplexer (gemux)

[![Godoc](https://godoc.org/github.com/fharding1/gemux?status.svg)](http://godoc.org/github.com/fharding1/gemux)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CircleCI](https://circleci.com/gh/fharding1/gemux.svg?style=svg)](https://circleci.com/gh/fharding1/gemux)

`gemux` is the good enough multiplexer. It aims to provide functionality that is good enough for the majority of HTTP services.

## Disclaimer

This project was mostly just written for fun. While it's intended to be production-ready, you're probably better
off using an older and more vetted muxer, or [not using one at all](https://blog.merovius.de/2017/06/18/how-not-to-use-an-http-router.html).

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

Route based on methods, and allow wildcard methods if you need to write your own method multiplexer (or don't care about methods for some reason).

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

### No External Dependencies

`gemux` does not depend upon any libraries outside of the standard library.

### It's Fast... Enough.

`gemux` isn't slow, and it's not fast either. If you're writing an HTTP service you're probably doing a bunch of I/O, so your multiplexer speed is almost totally insignificant. Since `gemux` doesn't try to do anything clever to be fast ([the root of all evil](http://wiki.c2.com/?PrematureOptimization)) it's more maintainable than other multiplexers.

## Won't Fix

There many features common in HTTP multiplexers that `gemux` does not and will not support.

### Subrouters

Subrouters are often a source of complexity without any real advantages. If you need subrouters, you can implement your own solution using this library, or just use another library like `gorilla/mux`.

### HandleFunc

There's no reason to have two methods that do the same thing. If you have an `http.HandlerFunc`, wrap it with `http.HandlerFunc` to make it an `http.Handler`.

### Middlewares

A middleware should just be a function that takes an `http.Handler` and returns one. Since you can already use middlewares this way, there's no reason for `gemux` to handle them.

### Method Helpers

There's no reason to have multiple methods for doing the same thing. If you need to handle a `GET`, just pass `http.MethodGet` to `gemux.ServeMux.Handle`.