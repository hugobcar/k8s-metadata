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

type JsonListClustersMap []jsonListClusters
type JsonApps []jsonApps
type JsonAppsByClustersMap map[string][]jsonAppsByClusters

type jsonListClusters struct {
	ID          int    `json:"id"`
	ClusterName string `json:"clusterName"`
	K8sVersion  string `json:"k8sVersion"`
}

type jsonApps struct {
	ClusterName    string `json:"clusterName"`
	AppName        string `json:"appName"`
	Namespace      string `json:"namespace"`
	AppType        string `json:"appType"`
	HelmVersion    string `json:"helmVersion"`
	HelmChart      string `json:"helmChart"`
	HelmAPPVersion string `json:"helmAPPVersion"`
}

type jsonAppsByClusters struct {
	AppName        string `json:"appName"`
	Namespace      string `json:"namespace"`
	AppType        string `json:"appType"`
	HelmVersion    string `json:"helmVersion"`
	HelmChart      string `json:"helmChart"`
	HelmAPPVersion string `json:"helmAPPVersion"`
}

// ListAllClusters - List all Clusters
func ListAllClusters(response *JsonListClustersMap) *JsonListClustersMap {
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

		*response = append(*response, jsonListClusters{ID: SIDCluster, ClusterName: SName})
	}

	return response
}

// ListApps - List details apps in clusters
func ListApps() JsonApps {
	var SClusterName string
	var SNamespace string
	var SAppName string
	var SAppType string
	var SHelmVersion string
	var SHelmChart string
	var SHelmAPPVersion string

	var response JsonApps

	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	rows, err := db.Query("SELECT clusters.nome, apps.namespace, apps.app, apps.type, IFNULL(helm.helm_version, \"\"), IFNULL(helm.chart, \"\"), IFNULL(helm.app_version, \"\") FROM apps INNER JOIN clusters ON (apps.id_cluster=clusters.id_cluster) LEFT JOIN helm ON (apps.app=helm.app AND apps.namespace=helm.namespace AND apps.id_cluster=helm.id_cluster) ORDER BY apps.namespace,apps.app")
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&SClusterName, &SNamespace, &SAppName, &SAppType, &SHelmVersion, &SHelmChart, &SHelmAPPVersion)
		checkErr(err)

		response = append(response, jsonApps{ClusterName: SClusterName, AppName: SAppName, Namespace: SNamespace, AppType: SAppType, HelmVersion: SHelmVersion, HelmChart: SHelmChart, HelmAPPVersion: SHelmAPPVersion})
	}

	return response
}

// ListAppsByClusters - List details apps by clusters
func ListAppsByClusters() JsonAppsByClustersMap {
	var SClusterName string
	var SNamespace string
	var SAppName string
	var SAppType string
	var SHelmVersion string
	var SHelmChart string
	var SHelmAPPVersion string

	response := make(JsonAppsByClustersMap)

	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	rows, err := db.Query("SELECT clusters.nome, apps.namespace, apps.app, apps.type, IFNULL(helm.helm_version, \"\"), IFNULL(helm.chart, \"\"), IFNULL(helm.app_version, \"\") FROM apps INNER JOIN clusters ON (apps.id_cluster=clusters.id_cluster) LEFT JOIN helm ON (apps.app=helm.app AND apps.namespace=helm.namespace AND apps.id_cluster=helm.id_cluster) ORDER BY apps.namespace,apps.app")
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&SClusterName, &SNamespace, &SAppName, &SAppType, &SHelmVersion, &SHelmChart, &SHelmAPPVersion)
		checkErr(err)

		response[SClusterName] = append(response[SClusterName], jsonAppsByClusters{AppName: SAppName, Namespace: SNamespace, AppType: SAppType, HelmVersion: SHelmVersion, HelmChart: SHelmChart, HelmAPPVersion: SHelmAPPVersion})
	}

	return response
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
