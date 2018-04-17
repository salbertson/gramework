// Copyright 2017 Kirill Danshin and Gramework contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//

package gramework

import (
	"fmt"
)

// JSON register internal handler that sets json content type
// and serves given handler with GET method
func (r *Router) JSON(route string, handler interface{}) *Router {
	h := r.determineHandler(handler)
	r.GET(route, jsonHandler(h))

	return r
}

// GET registers a handler for a GET request to the given route
func (r *Router) GET(route string, handler interface{}) *Router {
	r.Handle(MethodGET, route, handler)
	return r
}

// Forbidden serves 403 on route it registered on
func (r *Router) Forbidden(ctx *Context) {
	ctx.Forbidden()
}

// DELETE registers a handler for a DELETE request to the given route
func (r *Router) DELETE(route string, handler interface{}) *Router {
	r.Handle(MethodDELETE, route, handler)
	return r
}

// HEAD registers a handler for a HEAD request to the given route
func (r *Router) HEAD(route string, handler interface{}) *Router {
	r.Handle(MethodHEAD, route, handler)
	return r
}

// OPTIONS registers a handler for a OPTIONS request to the given route
func (r *Router) OPTIONS(route string, handler interface{}) *Router {
	r.Handle(MethodOPTIONS, route, handler)
	return r
}

// PUT registers a handler for a PUT request to the given route
func (r *Router) PUT(route string, handler interface{}) *Router {
	r.Handle(MethodPUT, route, handler)
	return r
}

// POST registers a handler for a POST request to the given route
func (r *Router) POST(route string, handler interface{}) *Router {
	r.Handle(MethodPOST, route, handler)
	return r
}

// PATCH registers a handler for a PATCH request to the given route
func (r *Router) PATCH(route string, handler interface{}) *Router {
	r.Handle(MethodPATCH, route, handler)
	return r
}

// ServeFile serves a file on a given route
func (r *Router) ServeFile(route, file string) *Router {
	r.Handle(MethodGET, route, func(ctx *Context) {
		ctx.SendFile(file)
	})
	return r
}

// SPAIndex serves an index file on any unregistered route
func (r *Router) SPAIndex(path string) *Router {
	r.NotFound(func(ctx *Context) {
		ctx.HTML()
		ctx.SendFile(path)
	})
	return r
}

// Sub let you quickly register subroutes with given prefix
// like app.Sub("v1").GET("route", "hi"), that give you /v1/route
// registered
func (r *Router) Sub(path string) *SubRouter {
	return &SubRouter{
		prefix: path,
		parent: r,
	}
}

func (r *Router) handleReg(method, route string, handler interface{}) {
	r.initRouter()
	r.app.Logger.Debugf("registering %s %s", method, route)
	r.router.Handle(method, route, r.determineHandler(handler))
}

func (r *Router) getEFuncStrHandler(h func() string) func(*Context) {
	return func(ctx *Context) {
		ctx.WriteString(h())
	}
}

// Handle registers a new request handle with the given path and method.
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut functions can be used.
// This function is intended for bulk loading and to allow the usage of less frequently used,
// non-standardized or custom methods (e.g. for internal communication with a proxy).
func (r *Router) Handle(method, route string, handler interface{}) *Router {
	r.handleReg(method, route, handler)
	return r
}

func (r *Router) getFmtVHandler(v interface{}) func(*Context) {
	cache := []byte(fmt.Sprintf("%v", v))
	return func(ctx *Context) {
		ctx.Write(cache)
	}
}

func (r *Router) getStringServer(str string) func(*Context) {
	b := []byte(str)
	return func(ctx *Context) {
		ctx.Write(b)
	}
}

func (r *Router) getBytesServer(b []byte) func(*Context) {
	return func(ctx *Context) {
		ctx.Write(b)
	}
}

func (r *Router) getFmtDHandler(v interface{}) func(*Context) {
	const fmtD = "%d"
	return func(ctx *Context) {
		fmt.Fprintf(ctx, fmtD, v)
	}
}

func (r *Router) getFmtFHandler(v interface{}) func(*Context) {
	const fmtF = "%f"
	return func(ctx *Context) {
		fmt.Fprintf(ctx, fmtF, v)
	}
}

// PanicHandler set a handler for unhandled panics
func (r *Router) PanicHandler(panicHandler func(*Context, interface{})) {
	r.initRouter()
	r.router.PanicHandler(panicHandler)
}

// NotFound set a handler which is called when no matching route is found
func (r *Router) NotFound(notFoundHandler func(*Context)) {
	r.initRouter()
	r.router.SetNotFound(notFoundHandler)
}

// HandleMethodNotAllowed changes HandleMethodNotAllowed mode in the router
func (r *Router) HandleMethodNotAllowed(newValue bool) (oldValue bool) {
	r.initRouter()
	return r.router.HandleMethodNotAllowed(newValue)
}

// HandleOPTIONS changes HandleOPTIONS mode in the router
func (r *Router) HandleOPTIONS(newValue bool) (oldValue bool) {
	r.initRouter()
	return r.router.HandleOPTIONS(newValue)
}

// HTTP router returns a router instance that work only on HTTP requests
func (r *Router) HTTP() *Router {
	if r.root != nil {
		return r.root.HTTP()
	}
	r.mu.Lock()
	if r.httprouter == nil {
		r.httprouter = &Router{
			router: r.app.newRouter(),
			app:    r.app,
			root:   r,
		}
	}
	r.mu.Unlock()

	return r.httprouter
}

// HTTPS router returns a router instance that work only on HTTPS requests
func (r *Router) HTTPS() *Router {
	if r.root != nil {
		return r.root.HTTPS()
	}
	r.mu.Lock()
	if r.httpsrouter == nil {
		r.httpsrouter = &Router{
			router: r.app.newRouter(),
			app:    r.app,
			root:   r,
		}
	}
	r.mu.Unlock()

	return r.httpsrouter
}

// ServeFiles serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
//     router.ServeFiles("/src/*filepath", "/var/www")
func (r *Router) ServeFiles(path string, rootPath string) {
	r.router.ServeFiles(path, rootPath)
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (r *Router) Lookup(method, path string, ctx *Context) (RequestHandler, bool) {
	return r.router.Lookup(method, path, ctx)
}

// MethodNotAllowed sets MethodNotAllowed handler
func (r *Router) MethodNotAllowed(c func(ctx *Context)) {
	r.router.MethodNotAllowed(c)
}

// Allowed returns Allow header's value used in OPTIONS responses
func (r *Router) Allowed(path, reqMethod string) (allow string) {
	return r.router.Allowed(path, reqMethod)
}

// Handler makes the router implement the fasthttp.ListenAndServe interface.
func (r *Router) Handler() func(*Context) {
	return func(ctx *Context) {
		path := string(ctx.Path())
		method := string(ctx.Method())

		switch ctx.IsTLS() {
		case true:
			if r.httpsrouter != nil {
				if !r.httpsrouter.router.Process(method, path, ctx) {
					r.httpsrouter.router.NotFound(ctx)
				}
				return
			}
		case false:
			if r.httprouter != nil {
				if !r.httpsrouter.router.Process(method, path, ctx) {
					r.httpsrouter.router.NotFound(ctx)
				}
				return
			}
		}
		if !r.httpsrouter.router.Process(method, path, ctx) {
			r.httpsrouter.router.NotFound(ctx)
		}
		return
	}
}

// Redir sends 301 redirect to the given url
//
// it's equivalent to
//
//     ctx.Redirect(url, 301)
func (r *Router) Redir(route, url string) {
	r.GET(route, func(ctx *Context) {
		ctx.Redirect(route, redirectCode)
	})
}
