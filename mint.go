package mint

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	mutex          sync.RWMutex
	defaultHandler = []Handler{loggerMW, customHeadersMW}
)

//JSON basic json type
type JSON map[string]interface{}

//Mint is framework's instance, it contains default middleware, DB, handlers configuration
//Create Intance of Mint using New() method
type Mint struct {
	// defaultHandler is default middleware like logger, Custom Headers
	defaultHandler []Handler
	//handlers contains HandlersContext information
	handlers       []*HandlersContext
	groupHandlers  []*HandlersGroup
	store          map[string]interface{}
	staticPath     string
	staticHandler  http.Handler
	router         *mux.Router
	contextPool    *sync.Pool
	gzipWriterPool *sync.Pool
	DB             *sql.DB
	built          bool
}

//Path sets URL Path to handler
func (mt *Mint) Path(path string) *HandlersContext {
	handlerContext := newHandlerContext(mt)
	mt.handlers = append(mt.handlers, handlerContext)
	return handlerContext.Path(path)
}

//RegisterDB sets db connection
func (mt *Mint) RegisterDB(db Database) *Mint {
	mt.DB = db.Connection()
	return mt
}

//Get the value from store by key
func (mt *Mint) Get(key string) interface{} {
	mutex.RLock()
	value := mt.store[key]
	mutex.RUnlock()
	return value
}

//Set the value to store with key
func (mt *Mint) Set(key string, value interface{}) {
	mutex.Lock()
	mt.store[key] = value
	mutex.Unlock()
}

//Views registers more than one view to application
func (mt *Mint) Views(vcs Views) *Mint {
	for _, vc := range vcs {
		mt.View(vc)
	}
	return mt
}

//View registers a single view to application
func (mt *Mint) View(vc ViewContext) *Mint {
	handlerContext := newHandlerContext(mt)
	handlerContext.Path(vc.path).Handlers(vc.handlers...).Methods(vc.methods...).Compressed(vc.compressed)
	mt.handlers = append(mt.handlers, handlerContext)
	return mt
}

//HandleStatic registers a new handler to handle static content such as img, css, html, js.
func (mt *Mint) HandleStatic(path string, dir string) {
	mt.staticPath = path
	mt.staticHandler = http.FileServer(http.Dir(dir))
}

func (mt *Mint) buildViews() {

	for _, handler := range mt.handlers {
		handler.build(mt.router)
	}
	for _, handlerGroup := range mt.groupHandlers {
		handlerGroup.build(mt.router)
	}
	if len(mt.staticPath) != 0 {
		mt.router.PathPrefix(mt.staticPath).Handler(mt.staticHandler)
	}
}

//New creates new application
func New() *Mint {
	mintEngine := Simple()
	mintEngine.defaultHandler = defaultHandler
	return mintEngine
}

//Use register new middleware
func (mt *Mint) Use(handler ...Handler) {
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

//GET register get handler
func (mt *Mint) GET(path string, handler Handler) *HandlersContext {
	return mt.SimpleHandler(path, http.MethodGet, handler)
}

//POST registers post handler
func (mt *Mint) POST(path string, handler Handler) *HandlersContext {
	return mt.SimpleHandler(path, http.MethodPost, handler)
}

//SimpleHandler registers simple handler
func (mt *Mint) SimpleHandler(path string, method string, handler ...Handler) *HandlersContext {
	hc := newHandlerContext(mt)
	hc.Methods(method)
	hc.Handlers(handler...)
	hc.Path(path)
	mt.handlers = append(mt.handlers, hc)
	return hc
}

//PUT register simple PUT handler
func (mt *Mint) PUT(path string, handler Handler) *HandlersContext {
	return mt.SimpleHandler(path, http.MethodPut, handler)
}

//DELETE register simple delete handler
func (mt *Mint) DELETE(path string, handler Handler) *HandlersContext {
	return mt.SimpleHandler(path, http.MethodDelete, handler)
}

//Run runs application
func (mt *Mint) Run(port string) {
	serverAdd := ":" + port
	fmt.Println("🚀  Starting server....")

	protocal := "http"
	var err error
	localAddress := protocal + "://localhost" + serverAdd
	fmt.Println("🌠 Ready on " + localAddress)
	err = http.ListenAndServe(serverAdd, handlers.RecoveryHandler()(mt.Build()))
	if err != nil {
		fmt.Println("Stopping the server" + err.Error())
	}

}
