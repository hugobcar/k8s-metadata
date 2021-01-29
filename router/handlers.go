package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	DB "github.com/hugobcar/k8s-metadata/models"
)

// GetIndex - Default route
func GetIndex(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "/v1/clusters \n")
	fmt.Fprintf(w, "/v1/apps \n")
	fmt.Fprintf(w, "/v1/appsbyclusters \n")
}

// GetClusters - Return Clusters List
func GetClusters(w http.ResponseWriter, r *http.Request) {
	var response *DB.JsonListClustersMap

	response = DB.ListAllClusters(&DB.JsonListClustersMap{})

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

// GetApps - Return detail apps in clusters
func GetApps(w http.ResponseWriter, r *http.Request) {
	var response DB.JsonApps

	response = DB.ListApps()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

// GetAppsByClusters - Return detail apps by Clusters in clusters
func GetAppsByClusters(w http.ResponseWriter, r *http.Request) {
	var response DB.JsonAppsByClustersMap

	response = DB.ListAppsByClusters()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}
