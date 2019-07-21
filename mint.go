package mint

import (
	"compress/gzip"
	"database/sql"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

var (
	mutex          sync.RWMutex
	defaultHandler = []Handler{loggerMW, customHeadersMW}
)

//Mint #
type Mint struct {
	defaultHandler []Handler
	handlers       []*HandlersContext
	store          map[string]interface{}
	staticPath     string
	staticHandler  http.Handler
	router         *mux.Router
	contextPool    *sync.Pool
	gzipWriterPool *sync.Pool
	DB             *sql.DB
}

//Path #
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

//Get #
func (mt *Mint) Get(key string) interface{} {
	mutex.RLock()
	value := mt.store[key]
	mutex.RUnlock()
	return value
}

//Set #
func (mt *Mint) Set(key string, value interface{}) {
	mutex.Lock()
	mt.store[key] = value
	mutex.Unlock()
}

//Views #
func (mt *Mint) Views(vcs Views) *Mint {
	for _, vc := range vcs {
		handlerContext := newHandlerContext(mt)
		handlerContext.Path(vc.path).Handlers(vc.handlers...).Methods(vc.methods...).Compressed(vc.compressed)
		mt.handlers = append(mt.handlers, handlerContext)
	}
	return mt
}

//View #
func (mt *Mint) View(vc ViewContext) *HandlersContext {
	handlerContext := newHandlerContext(mt)
	return handlerContext.Path(vc.path).Handlers(vc.handlers...).Methods(vc.methods...).Compressed(vc.compressed)
}

//HandleStatic registers a new handler to handle static content such as img, css, html, js.
func (mt *Mint) HandleStatic(path string, dir string) {
	mt.staticPath = path
	mt.staticHandler = http.FileServer(http.Dir(dir))
}

func (mt *Mint) buildViews() {

	for _, handler := range mt.handlers {
		mt.router.Handle(handler.path, handler).Methods(handler.methods...)
	}
	if len(mt.staticPath) != 0 {
		mt.router.PathPrefix(mt.staticPath).Handler(mt.staticHandler)
	}
}

//New creates new application
func New() *Mint {
	mintEngine := &Mint{}
	mintEngine.contextPool = &sync.Pool{
		New: newContextPool(mintEngine),
	}
	mintEngine.gzipWriterPool = &sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(nil)
		},
	}
	mintEngine.defaultHandler = defaultHandler
	mintEngine.store = make(map[string]interface{})
	mintEngine.router = NewRouter()
	return mintEngine
}

//Build the application
func (mt *Mint) Build() *mux.Router {
	mt.buildViews()
	return mt.router
}
