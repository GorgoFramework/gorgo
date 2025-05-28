package gorgo

import "strings"

type Router struct {
	routes map[string]map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]map[string]HandlerFunc),
	}
}

func (r *Router) AddRoute(method, path string, handler HandlerFunc) {
	if r.routes[method] == nil {
		r.routes[method] = make(map[string]HandlerFunc)
	}
	r.routes[method][path] = handler
}

func (r *Router) FindHandler(method, path string) HandlerFunc {
	if methodRoutes, exists := r.routes[method]; exists {
		if handler, exists := methodRoutes[path]; exists {
			return handler
		}

		for routePath, handler := range methodRoutes {
			if r.matchPath(routePath, path) {
				return handler
			}
		}
	}
	return nil
}

func (r *Router) matchPath(routePath, requestPath string) bool {
	routeParts := strings.Split(routePath, "/")
	requestParts := strings.Split(requestPath, "/")

	if len(routeParts) != len(requestParts) {
		return false
	}

	for i, routePart := range routeParts {
		if strings.HasPrefix(routePart, ":") {
			continue
		}
		if routePart != requestParts[i] {
			return false
		}
	}

	return true
}
