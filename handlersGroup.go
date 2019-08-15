package mint

import (
	"github.com/gorilla/mux"
)

//HandlersGroup #
type HandlersGroup struct {
	mint          *Mint
	middleware    []Handler
	basePath      string
	router        *mux.Router
	handlersGroup []*HandlersGroup
	handlers      []*HandlersContext
}

func (hg *HandlersGroup) build(parentRouter *mux.Router) {
	subrouter := parentRouter.PathPrefix(hg.basePath).Subrouter()
	for _, handler := range hg.handlers {
		handler.mint = hg.mint
		handler.middleware = append(hg.middleware, handler.middleware...)
		handler.build(subrouter)
	}
	for _, group := range hg.handlersGroup {
		group.mint = hg.mint
		group.middleware = append(hg.middleware, group.middleware...)
		group.build(subrouter)
	}
}

//NewGroup creates new handlers Group
func NewGroup(pathPrefix string) *HandlersGroup {
	handlersGroup := new(HandlersGroup)
	handlersGroup.basePath = pathPrefix
	return handlersGroup
}

//AddGroup add new subgroup
func (hg *HandlersGroup) AddGroup(newhg *HandlersGroup) *HandlersGroup {
	hg.handlersGroup = append(hg.handlersGroup, newhg)
	return hg
}

//AddGroups adds new subgroups
func (hg *HandlersGroup) AddGroups(hgs []*HandlersGroup) *HandlersGroup {
	for _, nhg := range hgs {
		hg.AddGroup(nhg)
	}
	return hg
}

//Group creates new subgroup
func (hg *HandlersGroup) Group(pathPrefix string) *HandlersGroup {
	handlersGroup := new(HandlersGroup)
	handlersGroup.basePath = pathPrefix
	hg.handlersGroup = append(hg.handlersGroup, handlersGroup)
	return handlersGroup
}

//Use register new middleware
func (hg *HandlersGroup) Use(handler ...Handler) *HandlersGroup {
	hg.middleware = append(hg.middleware, handler...)
	return hg
}

//Schemes #
func (hg *HandlersGroup) Schemes(schemes ...string) *HandlersGroup {
	hg.router.Schemes(schemes...)
	return hg
}

//Headers #
func (hg *HandlersGroup) Headers(headers ...string) *HandlersGroup {
	hg.router.Headers(headers...)
	return hg
}

//Queries #
func (hg *HandlersGroup) Queries(queries ...string) *HandlersGroup {
	hg.router.Queries(queries...)
	return hg
}

//SimpleHandler registers simple handler
func (hg *HandlersGroup) SimpleHandler(path string, method string, handler ...Handler) *HandlersContext {
	hc := new(HandlersContext)
	hc.Methods(method)
	hc.Handle(handler...)
	hc.Path(path)
	hg.handlers = append(hg.handlers, hc)
	return hc
}

//Handler registers new Handler
func (hg *HandlersGroup) Handler(hc *HandlersContext) *HandlersGroup {
	hg.handlers = append(hg.handlers, hc)
	return hg
}

//Handlers registers chain of handlers
func (hg *HandlersGroup) Handlers(hsc []*HandlersContext) *HandlersGroup {
	for _, handler := range hsc {
		hg.Handler(handler)
	}
	return hg
}
