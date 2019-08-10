package mint

import (
	"net/http"

	"github.com/gorilla/mux"
)

//Handler handles requests for an URL
type Handler func(*Context)

//HandlersContext #
type HandlersContext struct {
	mint       *Mint
	handlers   []Handler
	methods    []string
	schemes    []string
	headers    []string
	queries    []string
	path       string
	compressed bool
}

//newHandlerContext creates new app handler
func newHandlerContext(mint *Mint) *HandlersContext {
	handlerContext := &HandlersContext{
		mint: mint,
	}
	handlerContext.Handlers(mint.defaultHandler...)
	return handlerContext
}

func (hc *HandlersContext) build(router *mux.Router) {
	router.Handle(hc.path, hc).
		Methods(hc.methods...).
		Schemes(hc.schemes...).
		Headers(hc.headers...).
		Queries(hc.queries...)
}

//Methods #
func (hc *HandlersContext) Methods(methods ...string) *HandlersContext {
	hc.methods = append(hc.methods, methods...)
	return hc
}

//Handlers #
func (hc *HandlersContext) Handlers(handlers ...Handler) *HandlersContext {
	hc.use(handlers...)
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

func (hc *HandlersContext) use(handler ...Handler) {
	hc.handlers = append(hc.handlers, handler...)
}
