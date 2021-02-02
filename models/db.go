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
type DescriptionMap map[string][]DescriptionStruct

type jsonListClusters struct {
	Aws         AWS       `json:"aws"`
	ClusterName string    `json:"clusterName"`
	K8SVersion  string    `json:"k8sVersion"`
	Instances   Instances `json:"instances"`
}

type Instances struct {
	TotalInstances int            `json:"totalInstances"`
	Description    DescriptionMap `json:"description"`
}

type DescriptionStruct struct {
	Description
}

type Description struct {
	Type               string `json:"type"`
	TotalTypeInstances int    `json:"totalTypeInstances"`
}

type AWS struct {
	Account     int64  `json:"account"`
	Region      string `json:"region"`
	Environment string `json:"environment"`
}

type jsonApps struct {
	ClusterName  string `json:"clusterName"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Type         string `json:"type"`
	HpaEnabled   bool   `json:"hpaEnabled"`
	VaultEnabled bool   `json:"vaultEnabled"`
	Helm         Helm   `json:"helm"`
}

type Helm struct {
	Version    string `json:"version"`
	Chart      string `json:"chart"`
	APPVersion string `json:"appVersion"`
}

type jsonAppsByClusters struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Type         string `json:"type"`
	HpaEnabled   bool   `json:"hpaEnabled"`
	VaultEnabled bool   `json:"vaultEnabled"`
	Helm         Helm   `json:"helm"`
}

// ListAllClusters - List all Clusters
func ListAllClusters(response *JsonListClustersMap) *JsonListClustersMap {
	var SIDCluster int
	var SName string
	var SAWSAccount int64
	var SAWSRegion string
	var SAWSEnvironment string
	var SK8sVersion string

	var SNodeType string
	var SNodeInstance string
	var STotalInstances int

	var totalInstances int

	description := make(DescriptionMap)

	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	rows, err := db.Query("SELECT id_cluster, nome, aws_account, aws_region, aws_env, k8s_version FROM clusters ORDER BY nome")
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&SIDCluster, &SName, &SAWSAccount, &SAWSRegion, &SAWSEnvironment, &SK8sVersion)
		checkErr(err)

		description = DescriptionMap{}
		totalInstances = 0

		rows1, err := db.Query("SELECT node_type, node_instance, total_instances FROM nodes WHERE id_cluster=?", SIDCluster)
		checkErr(err)

		for rows1.Next() {
			err = rows1.Scan(&SNodeType, &SNodeInstance, &STotalInstances)
			checkErr(err)

			description[SNodeType] = append(
				description[SNodeType],
				DescriptionStruct{
					Description{
						Type:               SNodeInstance,
						TotalTypeInstances: STotalInstances,
					},
				},
			)

			totalInstances = totalInstances + STotalInstances
		}

		*response = append(
			*response,
			jsonListClusters{
				ClusterName: SName,
				Aws: AWS{
					Account:     SAWSAccount,
					Region:      SAWSRegion,
					Environment: SAWSEnvironment,
				},
				K8SVersion: SK8sVersion,
				Instances: Instances{
					TotalInstances: totalInstances,
					Description:    description,
				},
			},
		)
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
	var SHpaEnabled bool
	var SVaultEnabled bool

	var response JsonApps

	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	rows, err := db.Query("SELECT clusters.nome, apps.namespace, apps.app, apps.type, IFNULL(helm.helm_version, \"\"), IFNULL(helm.chart, \"\"), IFNULL(helm.app_version, \"\"), apps.hpa_enabled, apps.vault_enabled FROM apps INNER JOIN clusters ON (apps.id_cluster=clusters.id_cluster) LEFT JOIN helm ON (apps.app=helm.app AND apps.namespace=helm.namespace AND apps.id_cluster=helm.id_cluster) ORDER BY clusters.nome,apps.namespace,apps.app")
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&SClusterName, &SNamespace, &SAppName, &SAppType, &SHelmVersion, &SHelmChart, &SHelmAPPVersion, &SHpaEnabled, &SVaultEnabled)
		checkErr(err)

		response = append(
			response,
			jsonApps{
				ClusterName:  SClusterName,
				Name:         SAppName,
				Namespace:    SNamespace,
				Type:         SAppType,
				HpaEnabled:   SHpaEnabled,
				VaultEnabled: SVaultEnabled,
				Helm: Helm{
					Version:    SHelmVersion,
					Chart:      SHelmChart,
					APPVersion: SHelmAPPVersion,
				},
			},
		)
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
	var SHpaEnabled bool
	var SVaultEnabled bool

	response := make(JsonAppsByClustersMap)

	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	rows, err := db.Query("SELECT clusters.nome, apps.namespace, apps.app, apps.type, IFNULL(helm.helm_version, \"\"), IFNULL(helm.chart, \"\"), IFNULL(helm.app_version, \"\"), apps.hpa_enabled, apps.vault_enabled FROM apps INNER JOIN clusters ON (apps.id_cluster=clusters.id_cluster) LEFT JOIN helm ON (apps.app=helm.app AND apps.namespace=helm.namespace AND apps.id_cluster=helm.id_cluster) ORDER BY apps.namespace,apps.app")
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&SClusterName, &SNamespace, &SAppName, &SAppType, &SHelmVersion, &SHelmChart, &SHelmAPPVersion, &SHpaEnabled, &SVaultEnabled)
		checkErr(err)

		response[SClusterName] = append(
			response[SClusterName],
			jsonAppsByClusters{
				Name:         SAppName,
				Namespace:    SNamespace,
				Type:         SAppType,
				HpaEnabled:   SHpaEnabled,
				VaultEnabled: SVaultEnabled,
				Helm: Helm{
					Version:    SHelmVersion,
					Chart:      SHelmChart,
					APPVersion: SHelmAPPVersion,
				},
			},
		)
	}

	return response
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
