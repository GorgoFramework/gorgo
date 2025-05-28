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

func (r *Router) FindHandler(method, path string) (HandlerFunc, map[string]string) {
	if methodRoutes, exists := r.routes[method]; exists {
		if handler, exists := methodRoutes[path]; exists {
			return handler, nil
		}

		for routePath, handler := range methodRoutes {
			if params := r.matchPath(routePath, path); params != nil {
				return handler, params
			}
		}
	}
	return nil, nil
}

func (r *Router) matchPath(routePath, requestPath string) map[string]string {
	routeParts := strings.Split(routePath, "/")
	requestParts := strings.Split(requestPath, "/")

	if len(routeParts) != len(requestParts) {
		return nil
	}

	params := make(map[string]string)

	for i, routePart := range routeParts {
		if strings.HasPrefix(routePart, ":") {
			paramName := routePart[1:] // Remove the ':' prefix
			params[paramName] = requestParts[i]
			continue
		}
		if routePart != requestParts[i] {
			return nil
		}
	}

	return params
}
