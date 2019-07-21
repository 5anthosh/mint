package mint

import (
	"net/http"

	"github.com/gorilla/mux"
)

//Handler handles requests for an URL
type Handler func(*Context)

//HandlersContext #
type HandlersContext struct {
	mint     *Mint
	handlers []Handler
	methods  []string
	path     string
}

//newHandlerContext creates new app handler
func newHandlerContext(mint *Mint) *HandlersContext {
	handlerContext := &HandlersContext{
		mint: mint,
	}
	handlerContext.Handlers(mint.defaultHandler...)
	return handlerContext
}

//Methods #
func (hc *HandlersContext) Methods(methods ...string) *HandlersContext {
	hc.methods = append(hc.methods, methods...)
	return hc
}

//Handlers #
func (hc *HandlersContext) Handlers(handlers ...Handler) *HandlersContext {
	hc.handlers = append(hc.handlers, handlers...)
	return hc
}

//Path #
func (hc *HandlersContext) Path(path string) *HandlersContext {
	hc.path = path
	return hc
}

//ServeHTTP #
func (hc *HandlersContext) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := hc.mint.contextPool.Get().(*Context)
	c.Reset()
	c.HandlersContext = hc
	c.URLParams = mux.Vars(req)
	c.Request = req
	c.Response = w
	c.Next()
	hc.mint.contextPool.Put(c)
}

func (hc *HandlersContext) use(handler ...Handler) {
	hc.handlers = append(hc.handlers, handler...)
}
