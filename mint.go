package mint

import (
	"compress/gzip"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

var (
	mutex sync.RWMutex
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
}

//Path #
func (mt *Mint) Path(path string) *HandlersContext {
	handlerContext := newHandlerContext(mt)
	mt.handlers = append(mt.handlers, handlerContext)
	return handlerContext.Path(path)
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
		handlerContext.Path(vc.path).Handlers(vc.handlers...).Methods(vc.methods...)
	}
	return mt
}

//View #
func (mt *Mint) View(vc ViewContext) *HandlersContext {
	handlerContext := newHandlerContext(mt)
	return handlerContext.Path(vc.path).Handlers(vc.handlers...).Methods(vc.methods...)
}

//HandleStatic registers a new handler to handle static content such as img, css, html, js.
func (mt *Mint) HandleStatic(path string, dir string) {
	mt.staticPath = path
	mt.staticHandler = http.FileServer(http.Dir(dir))
}

func (mt *Mint) buildViews() {

	for _, handler := range mt.handlers {
		mt.router.PathPrefix(handler.path).Handler(handler).Methods(handler.methods...)
	}
	if len(mt.staticPath) != 0 {
		mt.router.PathPrefix(mt.staticPath).Handler(mt.staticHandler)
	}
}

//New creates new application
func New() *Mint {
	app := &Mint{}
	app.contextPool = &sync.Pool{
		New: newContextPool(app),
	}
	app.gzipWriterPool = &sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(nil)
		},
	}
	app.store = make(map[string]interface{})
	app.router = NewRouter()
	return app
}

//Build the application
func (mt *Mint) Build() {
	mt.buildViews()
}
