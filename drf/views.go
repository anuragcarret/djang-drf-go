package drf

import (
	"fmt"
	"net/http"
)

// ViewSet is the interface for handling API actions.
type ViewSet interface {
	List(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Retrieve(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Destroy(w http.ResponseWriter, r *http.Request)
}

// ModelViewSet provides a default implementation for CRUD.
type ModelViewSet struct {
	BaseSerializer BaseSerializer
}

func (v *ModelViewSet) List(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "List of objects (Generic Implementation)")
}

func (v *ModelViewSet) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Create object (Generic Implementation)")
}

func (v *ModelViewSet) Retrieve(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Retrieve object (Generic Implementation)")
}

func (v *ModelViewSet) Update(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Update object (Generic Implementation)")
}

func (v *ModelViewSet) Destroy(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Destroy object (Generic Implementation)")
}

// Router helps register ViewSets to paths.
type Router struct {
	mux *http.ServeMux
}

func NewRouter() *Router {
	return &Router{mux: http.NewServeMux()}
}

func (router *Router) Register(prefix string, vset ViewSet) {
	router.mux.HandleFunc(prefix+"/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			vset.List(w, r)
		case http.MethodPost:
			vset.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}
