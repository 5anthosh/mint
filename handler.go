package mint

import (
	"net/http"

	"github.com/gorilla/mux"
)

//Handler handles requests for an URL
type Handler func(*Context)

//Handlers chain of Handler
type Handlers []*HandlersContext

//HandlersContext #
type HandlersContext struct {
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

func (hc *HandlersContext) build(router *mux.Router) {
	hc.handlers = append(hc.middleware, hc.handlers...)
	hc.count = len(hc.handlers)
	route := router.Handle(hc.path, hc)
	addFilters(hc, route)
}

func addFilters(hc *HandlersContext, route *mux.Route) {
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
func (hc *HandlersContext) buildWithRoute(route *mux.Route) {
	hc.handlers = append(hc.middleware, hc.handlers...)
	hc.count = len(hc.handlers)
	route1 := route.Handler(hc)
	addFilters(hc, route1)
}

//Methods #
func (hc *HandlersContext) Methods(methods ...string) *HandlersContext {
	hc.methods = append(hc.methods, methods...)
	return hc
}

//Handle #
func (hc *HandlersContext) Handle(handlers ...Handler) *HandlersContext {
	hc.handlers = append(hc.handlers, handlers...)
	return hc
}

//Schemes #
func (hc *HandlersContext) Schemes(schemes ...string) *HandlersContext {
	hc.schemes = append(hc.schemes, schemes...)
	return hc
}

//Headers #
func (hc *HandlersContext) Headers(headers ...string) *HandlersContext {
	hc.headers = append(hc.headers, headers...)
	return hc
}

//Queries #
func (hc *HandlersContext) Queries(queries ...string) *HandlersContext {
	hc.queries = append(hc.queries, queries...)
	return hc
}

//Path #
func (hc *HandlersContext) Path(path string) *HandlersContext {
	hc.path = path
	return hc
}

//Name #
func (hc *HandlersContext) Name(name string) *HandlersContext {
	hc.name = name
	return hc
}

//Compressed #
func (hc *HandlersContext) Compressed(isCompressed bool) *HandlersContext {
	hc.compressed = isCompressed
	return hc
}

//ServeHTTP #
func (hc *HandlersContext) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := hc.mint.contextPool.Get().(*Context)
	c.Reset()
	c.HandlersContext = hc
	c.URLParams = mux.Vars(req)
	c.Request = req
	c.Method = req.Method
	c.Response = w
	c.DB = hc.mint.DB
	c.Next()
	hc.mint.contextPool.Put(c)
}

//Use registers middleware
func (hc *HandlersContext) Use(handler ...Handler) *HandlersContext {
	hc.middleware = append(hc.middleware, handler...)
	return hc
}
