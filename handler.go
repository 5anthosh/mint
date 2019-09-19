package mint

import (
	"net/http"

	"github.com/gorilla/mux"
)

//HandlerFunc handles requests for an URL
type HandlerFunc func(*Context)

//Handlers chain of Handler
type Handlers []*HandlerContext

//HandlerContext #
type HandlerContext struct {
	Mint       *Mint
	middleware []HandlerFunc
	handlers   []HandlerFunc
	validator  HandlerFunc
	count      int
	methods    []string
	schemes    []string
	headers    []string
	queries    []string
	path       string
	name       string
	compressed bool
}

//HandlerBuilder new handerContext
func HandlerBuilder() *HandlerContext {
	return new(HandlerContext)
}

func (hc *HandlerContext) build(router *mux.Router) {
	if hc == nil {
		return
	}
	hc.handlers = append(hc.middleware, hc.handlers...)
	hc.count = len(hc.handlers)
	route := router.Handle(hc.path, hc)
	addFilters(hc, route)
}

func addFilters(hc *HandlerContext, route *mux.Route) {
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
func (hc *HandlerContext) buildWithRoute(route *mux.Route) {
	if hc == nil {
		return
	}
	hc.handlers = append(hc.middleware, hc.handlers...)
	hc.count = len(hc.handlers)
	route1 := route.Handler(hc)
	addFilters(hc, route1)
}

//Methods #
func (hc *HandlerContext) Methods(methods ...string) *HandlerContext {
	if hc == nil {
		return hc
	}
	hc.methods = append(hc.methods, methods...)
	return hc
}

//Handle #
func (hc *HandlerContext) Handle(handlers ...HandlerFunc) *HandlerContext {
	if hc == nil {
		return hc
	}
	hc.handlers = append(hc.handlers, handlers...)
	return hc
}

//Schemes #
func (hc *HandlerContext) Schemes(schemes ...string) *HandlerContext {
	if hc == nil {
		return hc
	}
	hc.schemes = append(hc.schemes, schemes...)
	return hc
}

//Headers #
func (hc *HandlerContext) Headers(headers ...string) *HandlerContext {
	if hc == nil {
		return hc
	}
	hc.headers = append(hc.headers, headers...)
	return hc
}

//Queries #
func (hc *HandlerContext) Queries(queries ...string) *HandlerContext {
	if hc == nil {
		return hc
	}
	hc.queries = append(hc.queries, queries...)
	return hc
}

//Path #
func (hc *HandlerContext) Path(path string) *HandlerContext {
	if hc == nil {
		return hc
	}
	hc.path = path
	return hc
}

//Name #
func (hc *HandlerContext) Name(name string) *HandlerContext {
	if hc == nil {
		return hc
	}
	hc.name = name
	return hc
}

//Compressed #
func (hc *HandlerContext) Compressed(isCompressed bool) *HandlerContext {
	if hc == nil {
		return hc
	}
	hc.compressed = isCompressed
	return hc
}

//ServeHTTP #
func (hc *HandlerContext) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := hc.Mint.contextPool.Get().(*Context)
	c.Reset()
	c.HandlerContext = hc
	c.params = mux.Vars(req)
	c.Req = req
	c.Res = w
	c.Next()
	hc.Mint.contextPool.Put(c)
}

//Use registers middleware
func (hc *HandlerContext) Use(handler ...HandlerFunc) *HandlerContext {
	if hc == nil {
		return hc
	}
	hc.middleware = append(hc.middleware, handler...)
	return hc
}
