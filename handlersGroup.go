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
		handler.build(subrouter)
	}
	for _, group := range hg.handlersGroup {
		group.build(subrouter)
	}
}

//Group creates new subgroup
func (hg *HandlersGroup) Group() *HandlersGroup {
	handlersGroup := &HandlersGroup{}
	handlersGroup.mint = hg.mint
	handlersGroup.middleware = hg.middleware
	hg.handlersGroup = append(hg.handlersGroup, handlersGroup)
	return handlersGroup
}

//Use register new middleware
func (hg *HandlersGroup) Use(handler ...Handler) {
	hg.middleware = append(hg.middleware, handler...)
}

//View registers a single view to application
func (hg *HandlersGroup) View(vc ViewContext) *HandlersGroup {
	handlerContext := newHandlerContext(hg.mint)
	handlerContext.Path(vc.path).Handlers(vc.handlers...).Methods(vc.methods...).Compressed(vc.compressed)
	hg.handlers = append(hg.handlers, handlerContext)
	return hg
}

//Views registers more than one view to application
func (hg *HandlersGroup) Views(vcs Views) *HandlersGroup {
	for _, vc := range vcs {
		hg.View(vc)
	}
	return hg
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
	hc.Handlers(handler...)
	hc.Path(path)
	hg.handlers = append(hg.handlers, hc)
	return hc
}
