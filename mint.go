package mint

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	mutex sync.RWMutex
	//DefaultHandlerWithLogger middlewares including logger
	DefaultHandlerWithLogger = []HandlerFunc{loggerMW}
)

//JSON basic json type
type JSON map[string]interface{}

//Mint is framework's instance, it contains default middleware, DB, handlers configuration
//Create Intance of Mint using New() method
type Mint struct {
	// defaultHandler is default middleware like logger, Custom Headers
	defaultHandler []HandlerFunc
	//handlers contains HandlersContext information
	handlers         []*HandlerContext
	groupHandlers    []*HandlersGroup
	store            map[string]interface{}
	staticPath       string
	staticHandler    http.Handler
	router           *mux.Router
	contextPool      *sync.Pool
	gzipWriterPool   *sync.Pool
	built            bool
	strictSlash      bool
	notFoundHandler  *HandlerContext
	methodNotAllowed *HandlerContext
}

//Path sets URL Path to handler
func (mt *Mint) Path(path string) *HandlerContext {
	handlerContext := new(HandlerContext)
	mt.handlers = append(mt.handlers, handlerContext)
	return handlerContext.Path(path)
}

//StrictSlash enable strictslash in router
func (mt *Mint) StrictSlash(strictSlash bool) *Mint {
	mt.strictSlash = strictSlash
	return mt
}

//Get the value from store by key
func (mt *Mint) Get(key string) (interface{}, bool) {
	mutex.RLock()
	value, ok := mt.store[key]
	mutex.RUnlock()
	return value, ok
}

//Set the value to store with key
func (mt *Mint) Set(key string, value interface{}) {
	mutex.Lock()
	mt.store[key] = value
	mutex.Unlock()
}

//Handler registers single handlers context
func (mt *Mint) Handler(hc *HandlerContext) *Mint {
	mt.handlers = append(mt.handlers, hc)
	return mt
}

//Handlers registers multiple handlers context
func (mt *Mint) Handlers(hsc []*HandlerContext) *Mint {
	for _, handler := range hsc {
		mt.Handler(handler)
	}
	return mt
}

//NotFoundHandler registers not found handler context
func (mt *Mint) NotFoundHandler(hc *HandlerContext) {
	hc.Mint = mt
	mt.notFoundHandler = hc
}

//MethodNotAllowedHandler registers method not allowed handler
func (mt *Mint) MethodNotAllowedHandler(hc *HandlerContext) {
	hc.Mint = mt
	mt.methodNotAllowed = hc
}

//HandleStatic registers a new handler to handle static content such as img, css, html, js.
func (mt *Mint) HandleStatic(path string, dir string) {
	mt.staticPath = path
	mt.staticHandler = http.FileServer(http.Dir(dir))
}

//ChainGroups chains groups in linear
func (mt *Mint) ChainGroups(groups []*HandlersGroup) *Mint {
	count := len(groups)
	if count > 0 {
		parentGroup := groups[0]
		mt.AddGroup(parentGroup)
		for iter := 1; iter < count; iter++ {
			parentGroup.AddGroup(groups[iter])
			parentGroup = groups[iter]
		}
	}
	return mt
}

func (mt *Mint) buildViews() {
	mt.router.StrictSlash(mt.strictSlash)
	mt.buildOtherHandlers()
	for _, handler := range mt.handlers {
		handler.Mint = mt
		handler.middleware = append(mt.defaultHandler, handler.middleware...)
		handler.build(mt.router)
	}
	for _, handlerGroup := range mt.groupHandlers {
		handlerGroup.mint = mt
		handlerGroup.middleware = append(mt.defaultHandler, handlerGroup.middleware...)
		handlerGroup.build(mt.router)
	}
	if len(mt.staticPath) != 0 {
		mt.router.PathPrefix(mt.staticPath).Handler(mt.staticHandler)
	}
}

func (mt *Mint) buildOtherHandlers() {
	handlers := append(mt.defaultHandler, mt.notFoundHandler.middleware...)
	handlers = append(handlers, mt.notFoundHandler.handlers...)
	mt.notFoundHandler.count = len(handlers)
	mt.notFoundHandler.handlers = handlers
	mt.router.NotFoundHandler = mt.notFoundHandler

	handlers = append(mt.defaultHandler, mt.methodNotAllowed.middleware...)
	handlers = append(handlers, mt.methodNotAllowed.handlers...)
	mt.methodNotAllowed.count = len(handlers)
	mt.methodNotAllowed.handlers = handlers
	mt.router.MethodNotAllowedHandler = mt.methodNotAllowed
}

//Group creates new group handlers W
func (mt *Mint) Group(pathPrefix string) *HandlersGroup {
	handlersGroup := &HandlersGroup{}
	handlersGroup.basePath = pathPrefix
	mt.groupHandlers = append(mt.groupHandlers, handlersGroup)
	return handlersGroup
}

//AddGroup adds a group to router
func (mt *Mint) AddGroup(hg *HandlersGroup) *Mint {
	mt.groupHandlers = append(mt.groupHandlers, hg)
	return mt
}

//AddGroups adds a groups of handlers
func (mt *Mint) AddGroups(hgs []*HandlersGroup) *Mint {
	for _, hg := range hgs {
		mt.AddGroup(hg)
	}
	return mt
}

//New creates new application
func New() *Mint {
	mintEngine := Simple()
	mintEngine.defaultHandler = DefaultHandlerWithLogger
	mintEngine.NotFoundHandler(new(HandlerContext).Handle(notFoundHandler))
	mintEngine.MethodNotAllowedHandler(new(HandlerContext).Handle(methodNotAllowedHandler))
	return mintEngine
}

//Use register new middleware
func (mt *Mint) Use(handler ...HandlerFunc) {
	mt.defaultHandler = append(mt.defaultHandler, handler...)
}

//Build the application
func (mt *Mint) Build() *mux.Router {
	if !mt.built {
		mt.buildViews()
		mt.built = true
	}
	return mt.router
}

//Simple creates new application without any defualt handlers
func Simple() *Mint {
	mintEngine := &Mint{}
	mintEngine.contextPool = &sync.Pool{
		New: newContextPool(mintEngine),
	}
	mintEngine.gzipWriterPool = &sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(nil)
		},
	}
	mintEngine.store = make(map[string]interface{})
	mintEngine.router = NewRouter()
	mintEngine.built = false
	return mintEngine
}

//From registers router to mt
func From(r Router, mt *Mint) *Mint {
	mt.router = r
	return mt
}

//GET register get handler
func (mt *Mint) GET(path string, handler HandlerFunc) *HandlerContext {
	return mt.SimpleHandler(path, http.MethodGet, handler)
}

//POST registers post handler
func (mt *Mint) POST(path string, handler HandlerFunc) *HandlerContext {
	return mt.SimpleHandler(path, http.MethodPost, handler)
}

//SimpleHandler registers simple handler
func (mt *Mint) SimpleHandler(path string, method string, handler ...HandlerFunc) *HandlerContext {
	hc := new(HandlerContext)
	hc.Methods(method)
	hc.Handle(handler...)
	hc.Path(path)
	mt.handlers = append(mt.handlers, hc)
	return hc
}

//PUT register simple PUT handler
func (mt *Mint) PUT(path string, handler HandlerFunc) *HandlerContext {
	return mt.SimpleHandler(path, http.MethodPut, handler)
}

//DELETE register simple delete handler
func (mt *Mint) DELETE(path string, handler HandlerFunc) *HandlerContext {
	return mt.SimpleHandler(path, http.MethodDelete, handler)
}

//Run runs application
func (mt *Mint) Run(serverAdd string) {
	fmt.Println("🚀  Starting server....")
	protocal := "http"
	localAddress := protocal + "://localhost" + serverAdd
	fmt.Println("🌠 Ready on " + localAddress)
	err := http.ListenAndServe(serverAdd, handlers.RecoveryHandler()(mt.Build()))
	if err != nil {
		fmt.Println("Stopping the server" + err.Error())
	}

}

//URLVar formats url var
func URLVar(urlvar string) string {
	return "{" + urlvar + "}"
}
