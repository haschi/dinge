package webx

import "net/http"

type Module interface {
	Mount(mux *http.ServeMux, prefix string, middleware ...Middleware)
}
