package mint

import (
	"github.com/gorilla/mux"
)

//HandlersGroup #
type HandlersGroup struct {
	mint          *Mint
	middleware    []Handler
	basePath      string
	prefixHandler *HandlerContext
	router        *mux.Router
	handlersGroup []*HandlersGroup
	handlers      []*HandlerContext
}

func (hg *HandlersGroup) build(parentRouter *mux.Router) {
	if hg == nil {
		return
	}
	route := parentRouter.PathPrefix(hg.basePath)
	if hg.prefixHandler != nil {
		hg.prefixHandler.mint = hg.mint
		hg.prefixHandler.middleware = append(hg.middleware, hg.prefixHandler.middleware...)
		hg.prefixHandler.buildWithRoute(route)
		return
	}
	subrouter := route.Subrouter()
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

//PrefixHandler registers handler for prefix request
func (hg *HandlersGroup) PrefixHandler(hc *HandlerContext) {
	if hg == nil {
		return
	}
	hg.prefixHandler = hc
}

//NewGroup creates new handlers Group
func NewGroup(pathPrefix string) *HandlersGroup {
	handlersGroup := new(HandlersGroup)
	handlersGroup.basePath = pathPrefix
	return handlersGroup
}

//AddGroup add new subgroup
func (hg *HandlersGroup) AddGroup(newhg *HandlersGroup) *HandlersGroup {
	if hg == nil {
		return hg
	}
	hg.handlersGroup = append(hg.handlersGroup, newhg)
	return hg
}

//AddGroups adds new subgroups
func (hg *HandlersGroup) AddGroups(hgs []*HandlersGroup) *HandlersGroup {
	if hg == nil {
		return hg
	}
	for _, nhg := range hgs {
		hg.AddGroup(nhg)
	}
	return hg
}

//ChainGroups chains groups in linear
func (hg *HandlersGroup) ChainGroups(groups []*HandlersGroup) *HandlersGroup {
	if hg == nil {
		return hg
	}
	count := len(groups)
	if count > 0 {
		parentGroup := groups[0]
		hg.AddGroup(parentGroup)
		for iter := 1; iter < count; iter++ {
			parentGroup.AddGroup(groups[iter])
			parentGroup = groups[iter]
		}
	}
	return hg
}

//Group creates new subgroup
func (hg *HandlersGroup) Group(pathPrefix string) *HandlersGroup {
	if hg == nil {
		return hg
	}
	handlersGroup := new(HandlersGroup)
	handlersGroup.basePath = pathPrefix
	hg.handlersGroup = append(hg.handlersGroup, handlersGroup)
	return handlersGroup
}

//Use register new middleware
func (hg *HandlersGroup) Use(handler ...Handler) *HandlersGroup {
	if hg == nil {
		return hg
	}
	hg.middleware = append(hg.middleware, handler...)
	return hg
}

//Schemes #
func (hg *HandlersGroup) Schemes(schemes ...string) *HandlersGroup {
	if hg == nil {
		return hg
	}
	hg.router.Schemes(schemes...)
	return hg
}

//Headers #
func (hg *HandlersGroup) Headers(headers ...string) *HandlersGroup {
	if hg == nil {
		return hg
	}
	hg.router.Headers(headers...)
	return hg
}

//Queries #
func (hg *HandlersGroup) Queries(queries ...string) *HandlersGroup {
	if hg == nil {
		return hg
	}
	hg.router.Queries(queries...)
	return hg
}

//SimpleHandler registers simple handler
func (hg *HandlersGroup) SimpleHandler(path string, method string, handler ...Handler) *HandlerContext {
	if hg == nil {
		return nil
	}
	hc := HandlerBuilder()
	hc.Methods(method)
	hc.Handle(handler...)
	hc.Path(path)
	hg.handlers = append(hg.handlers, hc)
	return hc
}

//Handler registers new Handler
func (hg *HandlersGroup) Handler(hc *HandlerContext) *HandlersGroup {
	if hg == nil {
		return nil
	}
	hg.handlers = append(hg.handlers, hc)
	return hg
}

//Handlers registers chain of handlers
func (hg *HandlersGroup) Handlers(hsc []*HandlerContext) *HandlersGroup {
	if hg == nil {
		return nil
	}
	for _, handler := range hsc {
		hg.Handler(handler)
	}
	return hg
}
