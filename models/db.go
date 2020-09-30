package models

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// DB Variables
var UserDB string
var PassDB string
var DatabaseDB string
var HostDB string
var PortDB string

type JsonListClustersResponse []jsonListClusters
type JsonDetailsResponse map[string][]jsonDetails

type jsonListClusters struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	K8sVersion string `json:"k8sVersion"`
}

type jsonDetails struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	Type           string `json:"type"`
	HelmVersion    string `json:"helmVersion"`
	HelmChart      string `json:"helmChart"`
	HelmAPPVersion string `json:"helmAPPVersion"`
}

// ListAllClusters - List all Clusters
func ListAllClusters(response *JsonListClustersResponse) *JsonListClustersResponse {
	var SIDCluster int
	var SName string

	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	rows, err := db.Query("SELECT id_cluster, nome FROM clusters")
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&SIDCluster, &SName)
		checkErr(err)

		*response = append(*response, jsonListClusters{ID: SIDCluster, Name: SName})
	}

	return response
}

// ListDetails - List details apps in clusters
func ListDetails() JsonDetailsResponse {
	var SClusterName string
	var SNamespace string
	var SAppName string
	var SAppType string
	var SHelmVersion string
	var SHelmChart string
	var SHelmAPPVersion string

	response := make(JsonDetailsResponse)

	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()
	//helm.helm_version,

	rows, err := db.Query("SELECT clusters.nome, apps.namespace, apps.app, apps.type, IFNULL(helm.helm_version, \"\"), IFNULL(helm.chart, \"\"), IFNULL(helm.app_version, \"\") FROM apps INNER JOIN clusters ON (apps.id_cluster=clusters.id_cluster) LEFT JOIN helm ON (apps.app=helm.app AND apps.namespace=helm.namespace AND apps.id_cluster=helm.id_cluster) ORDER BY apps.namespace,apps.app")
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&SClusterName, &SNamespace, &SAppName, &SAppType, &SHelmVersion, &SHelmChart, &SHelmAPPVersion)
		checkErr(err)

		response[SClusterName] = append(response[SClusterName], jsonDetails{Name: SAppName, Namespace: SNamespace, Type: SAppType, HelmVersion: SHelmVersion, HelmChart: SHelmChart, HelmAPPVersion: SHelmAPPVersion})
	}

	return response
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
