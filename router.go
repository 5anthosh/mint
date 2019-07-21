package mint

import (
	"github.com/gorilla/mux"
)

//Router route
type Router *mux.Router

//NewRouter creates new router for application
func NewRouter() Router {
	return mux.NewRouter()
}
