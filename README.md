# Good Enough Multiplexer (gemux)

`gemux` to be "good enough" for most Go web APIs (unlike `http.ServeMux`), and doesn't attempt to cover all use cases (like `gorilla/mux`).

## Usage

```go
package main

func main() {
    r := new(gemux.ServeMux)

    r.Handle("/", "*", http.HandlerFunc(healthHandler)) // write your own method mux within the handler
    r.Handle("/posts", http.MethodGet, http.HandlerFunc(getPostsHandler))
    r.Handle("/posts", http.MethodPost, http.HandlerFunc(createPostHandler))
    r.Handle("/posts/*", http.MethodGet, http.HandlerFunc(getPostHandler)) // use gemux.PathParameters to extract wildcard values
    r.Handle("/posts/*", http.MethodDelete, http.HandlerFunc(deletePostHandler))
    r.Handle("/posts/*/comments", http.MethodGet, http.HandlerFunc(getCommentsHandler))
    r.Handle("/posts/*/comments", http.MethodPost, http.HandlerFunc(createCommentHandler))
    r.Handle("/posts/*/comments/*", http.MethodGet, http.HandlerFunc(getCommentHandler))
    r.Handle("/posts/*/comments/*", http.MethodDelete, http.HandlerFunc(deleteCommentHandler))

    http.ListenAndServe(":8080", r)
}
```

## Rationale

There are already a billion HTTP Go multiplexers, so why was `gemux` created?

* While it's [possible](https://blog.merovius.de/2017/06/18/how-not-to-use-an-http-router.html) to use the standard library `net/http.ServeMux`, it's insufficient for most projects without writing lots of redundant code that's hard to follow (no quick list of routes)
* Libraries like `gorilla/mux` are great, but they intend to cover many use cases and therefore have too much cleverness/magic, as well as a very large API, which means there's a high learning curve and it's sometimes difficult to understand what exactly is going on
* Other libraries weigh speed over practicality and readability, causing weird APIs with hard to read codebases
* Some libraries disallow things the HTTP specification allows (e.g. treat methods as something other than just a plain string)

## Features

### Strict Path Based Routing (with wildcards)

Route based on paths strictly, but allow wildcards for path parameters such as a resource ID.

### Strict Method Based Routing (with wildcards)

Route based on methods, and allow wildcard methods if you need to write your own method multiplexer.

### Context Path Parameters

Extract path wildcard values via the request context. Some request multiplexers allow named path parameters. `gemux` does not because there is only a slight benifit advantage of doing this (reusable path parameter parsers), and it would add complexity.

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

### Very Small API

`gemux` has a very small exported API, and not much "magic" within the `Handle` function (such as named path parameters or regular expressions) which means it's easier to learn than other multiplexer libraries.

### Just Standard Library Dependencies

`gemux` does not depend upon any libraries outside of the standard library.

### It's Fast

Gotcha. Not really (but it is faster than some popular ones). However it doesn't really matter how fast your HTTP multiplexer is. If you're writing an HTTP service you're probably doing a bunch of I/O, so your multiplexer speed is almost totally insignificant. Since `gemux` doesn't try to do anything clever to be fast ([the root of all evil](http://wiki.c2.com/?PrematureOptimization)) it's more maintainable than other multiplexers.

### Sticks to the Standard

`gemux` handlers are just `http.Handler`s so you can easily plug in other middlewares without writing wrapper functions. `gemux` also recognizes that HTTP methods are ultimately just strings, so it doesn't do anything weird like provide helper functions named after methods (e.g. `router.Get()`). You can use `http.MethodGet` if you need a GET method.