package gee

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := filterNonEmpty(strings.Split(pattern, "/"))
	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

func (r *router) handle(w http.ResponseWriter, req *http.Request, middlewares []HandlerFunc) {
	n, params := r.getRoute(req.Method, req.URL.Path)
	handlers := middlewares

	if n != nil {
		key := req.Method + "-" + n.pattern
		handlers = append(middlewares, r.handlers[key])
	} else {
		handlers = append(middlewares, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c := newContext(w, req, params, handlers)
	c.Next()
}

func (r *router) getRoute(method string, pattern string) (*node, map[string]string) {
	searchParts := filterNonEmpty(strings.Split(pattern, "/"))
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)

	if n != nil {
		parts := filterNonEmpty(strings.Split(n.pattern, "/"))
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
		}
		return n, params
	}

	return nil, nil
}

func (r *router) getRoutes(method string) []*node {
	root, ok := r.roots[method]
	if !ok {
		return nil
	}
	nodes := make([]*node, 0)
	root.travel(&nodes)
	return nodes
}

func filterNonEmpty(vs []string) []string {
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
		}
	}
	return parts
}
