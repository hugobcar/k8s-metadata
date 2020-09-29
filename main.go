package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/hugobcar/k8s-metadata/router"

	DB "github.com/hugobcar/k8s-metadata/models"
)

var portListen string
var userDB string
var passDB string
var database string
var hostDB string
var portDB string

func init() {
	flag.StringVar(&portListen, "port", "6885", "Port application to listenner")
	flag.StringVar(&userDB, "userDB", "k8s", "User MySQL")
	flag.StringVar(&passDB, "passDB", "", "Password MySQL. Use default case your not specific.")
	flag.StringVar(&database, "database", "k8s", "Database MySQL")
	flag.StringVar(&hostDB, "hostDB", "localhost", "Hostname MySQL")
	flag.StringVar(&portDB, "portDB", "3306", "Port MySQL")

	flag.Parse()

	DB.UserDB = userDB
	DB.PassDB = passDB
	DB.DatabaseDB = database
	DB.HostDB = hostDB
	DB.PortDB = portDB
}

func main() {
	router := router.NewRouter()

	log.Fatal(http.ListenAndServe(":"+portListen, router))
}
