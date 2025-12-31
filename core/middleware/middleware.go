package middleware

import (
	"net/http"
)

// Middleware wraps an http.Handler with additional logic
type Middleware func(http.Handler) http.Handler

// Chain composes multiple middleware into a single handler
type Chain struct {
	middlewares []Middleware
}

// New creates a middleware chain
func New(middlewares ...Middleware) *Chain {
	return &Chain{
		middlewares: append(([]Middleware)(nil), middlewares...),
	}
}

// Then wraps the final handler with all middleware
func (c *Chain) Then(handler http.Handler) http.Handler {
	if handler == nil {
		handler = http.DefaultServeMux
	}

	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i](handler)
	}

	return handler
}

// ThenFunc wraps an http.HandlerFunc
func (c *Chain) ThenFunc(fn http.HandlerFunc) http.Handler {
	if fn == nil {
		return c.Then(nil)
	}
	return c.Then(http.HandlerFunc(fn))
}

// Append adds middleware to the chain
func (c *Chain) Append(middlewares ...Middleware) *Chain {
	newMiddlewares := make([]Middleware, 0, len(c.middlewares)+len(middlewares))
	newMiddlewares = append(newMiddlewares, c.middlewares...)
	newMiddlewares = append(newMiddlewares, middlewares...)
	return &Chain{middlewares: newMiddlewares}
}
