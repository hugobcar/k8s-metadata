package router

import (
	"net/http"

	"github.com/gorilla/mux"
)

type routeStruct struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type routesStruct []routeStruct

var env string
var routes routesStruct

// NewRouter - Used for create a routes
func NewRouter() *mux.Router {
	routes = routesStruct{
		routeStruct{
			"GetClusters",
			"GET",
			"/",
			GetIndex,
		},
		routeStruct{
			"GetClusters",
			"GET",
			"/v1/clusters",
			GetClusters,
		},
		routeStruct{
			"GetApps",
			"GET",
			"/v1/apps",
			GetApps,
		},
		routeStruct{
			"GetAppsByClusters",
			"GET",
			"/v1/appsbyclusters",
			GetAppsByClusters,
		},
	}

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	return router
}
