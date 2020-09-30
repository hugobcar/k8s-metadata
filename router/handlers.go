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
	fmt.Fprintf(w, "/v1/details \n")
	fmt.Fprintf(w, "/v1/newapps")
	fmt.Fprintf(w, "/v1/removedapps")
}

// GetClusters - Return Clusters List
func GetClusters(w http.ResponseWriter, r *http.Request) {
	var response *DB.JsonListClustersResponse

	response = DB.ListAllClusters(&DB.JsonListClustersResponse{})

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

// GetDetails - Return details apps in clusters
func GetDetails(w http.ResponseWriter, r *http.Request) {
	var response DB.JsonDetailsResponse

	response = DB.ListDetails()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}
