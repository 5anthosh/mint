package mint

import (
	"github.com/gorilla/mux"
)

//HandlersGroup #
type HandlersGroup struct {
	mint          *Mint
	middleware    []Handler
	basePath      string
	host          string
	schemes       []string
	headers       []string
	queries       []string
	methods       []string
	handlersGroup []*HandlersGroup
	handlers      []*HandlersContext
}

func (hg *HandlersGroup) build(parentRouter *mux.Router) {
	subrouter := parentRouter.PathPrefix(hg.basePath).Subrouter()
	subrouter.
		Host(hg.host).
		Schemes(hg.schemes...).
		Headers(hg.headers...).
		Queries(hg.queries...).
		Methods(hg.methods...)
	for _, handler := range hg.handlers {
		handler.middleware = append(hg.middleware, handler.middleware...)
		handler.build(subrouter)
	}
	for _, group := range hg.handlersGroup {
		group.middleware = append(hg.middleware, group.middleware...)
		group.build(subrouter)
	}
}

//Group creates new subgroup
func (hg *HandlersGroup) Group(pathPrefix string) *HandlersGroup {
	handlersGroup := &HandlersGroup{}
	handlersGroup.basePath = pathPrefix
	handlersGroup.mint = hg.mint
	hg.handlersGroup = append(hg.handlersGroup, handlersGroup)
	return handlersGroup
}

//Use register new middleware
func (hg *HandlersGroup) Use(handler ...Handler) {
	hg.middleware = append(hg.middleware, handler...)
}

//Schemes #
func (hg *HandlersGroup) Schemes(schemes ...string) *HandlersGroup {
	hg.schemes = append(hg.schemes, schemes...)
	return hg
}

//Headers #
func (hg *HandlersGroup) Headers(headers ...string) *HandlersGroup {
	hg.headers = append(hg.headers, headers...)
	return hg
}

//Queries #
func (hg *HandlersGroup) Queries(queries ...string) *HandlersGroup {
	hg.queries = append(hg.queries, queries...)
	return hg
}

//SimpleHandler registers simple handler
func (hg *HandlersGroup) SimpleHandler(path string, method string, handler ...Handler) *HandlersContext {
	hc := newHandlerContext(hg.mint)
	hc.Methods(method)
	hc.Handle(handler...)
	hc.Path(path)
	hg.handlers = append(hg.handlers, hc)
	return hc
}
