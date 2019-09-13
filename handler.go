package mint

import (
	"net/http"

	"github.com/gorilla/mux"
)

//Handler handles requests for an URL
type Handler func(*Context)

//Handlers chain of Handler
type Handlers []*HandlerBuilder

//HandlerBuilder #
type HandlerBuilder struct {
	mint       *Mint
	middleware []Handler
	handlers   []Handler
	count      int
	methods    []string
	schemes    []string
	headers    []string
	queries    []string
	path       string
	name       string
	compressed bool
}

func (hc *HandlerBuilder) build(router *mux.Router) {
	hc.handlers = append(hc.middleware, hc.handlers...)
	hc.count = len(hc.handlers)
	route := router.Handle(hc.path, hc)
	addFilters(hc, route)
}

func addFilters(hc *HandlerBuilder, route *mux.Route) {
	if len(hc.methods) > 0 {
		route.Methods(hc.methods...)
	}
	if len(hc.schemes) > 0 {
		route.Schemes(hc.schemes...)
	}
	if len(hc.headers) > 0 {
		route.Headers(hc.headers...)
	}
	if len(hc.queries) > 0 {
		route.Queries(hc.queries...)
	}
	if len(hc.name) > 0 {
		route.Name(hc.name)
	}
}
func (hc *HandlerBuilder) buildWithRoute(route *mux.Route) {
	hc.handlers = append(hc.middleware, hc.handlers...)
	hc.count = len(hc.handlers)
	route1 := route.Handler(hc)
	addFilters(hc, route1)
}

//Methods #
func (hc *HandlerBuilder) Methods(methods ...string) *HandlerBuilder {
	hc.methods = append(hc.methods, methods...)
	return hc
}

//Handle #
func (hc *HandlerBuilder) Handle(handlers ...Handler) *HandlerBuilder {
	hc.handlers = append(hc.handlers, handlers...)
	return hc
}

//Schemes #
func (hc *HandlerBuilder) Schemes(schemes ...string) *HandlerBuilder {
	hc.schemes = append(hc.schemes, schemes...)
	return hc
}

//Headers #
func (hc *HandlerBuilder) Headers(headers ...string) *HandlerBuilder {
	hc.headers = append(hc.headers, headers...)
	return hc
}

//Queries #
func (hc *HandlerBuilder) Queries(queries ...string) *HandlerBuilder {
	hc.queries = append(hc.queries, queries...)
	return hc
}

//Path #
func (hc *HandlerBuilder) Path(path string) *HandlerBuilder {
	hc.path = path
	return hc
}

//Name #
func (hc *HandlerBuilder) Name(name string) *HandlerBuilder {
	hc.name = name
	return hc
}

//Compressed #
func (hc *HandlerBuilder) Compressed(isCompressed bool) *HandlerBuilder {
	hc.compressed = isCompressed
	return hc
}

//ServeHTTP #
func (hc *HandlerBuilder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := hc.mint.contextPool.Get().(*Context)
	c.Reset()
	c.HandlerBuilder = hc
	c.URLParams = mux.Vars(req)
	c.Request = req
	c.Method = req.Method
	c.Response = w
	c.Next()
	hc.mint.contextPool.Put(c)
}

//Use registers middleware
func (hc *HandlerBuilder) Use(handler ...Handler) *HandlerBuilder {
	hc.middleware = append(hc.middleware, handler...)
	return hc
}
