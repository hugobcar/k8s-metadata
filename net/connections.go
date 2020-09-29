package net

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// ValidaURLProxyPass - Metodo usado para validar se a URL tem o http:// e https://
func ValidaURLProxyPass(proxy string) bool {
	proxy = proxy[:8]

	// Verifica se tem ponto e virgula na URL
	if strings.Contains(proxy, ";") {
		return false
	} else {
		if strings.Contains(proxy, "http://") {
			return true
		} else {
			if strings.Contains(proxy, "https://") {
				return true
			} else {
				return false
			}
		}
	}
}

// ReturnHostandPortURL - Metodo usado para retornar apenas o HOST e PORTA de uma determinada URL
func ReturnHostandPortURL(proxy string) (host, port string) {
	if strings.Contains(proxy, "http://") {
		port = "80"
	} else {
		if strings.Contains(proxy, "https://") {
			port = "443"
		}
	}

	// Retira http:// ou https://
	proxy = strings.Replace(proxy, "http://", "", -1)
	proxy = strings.Replace(proxy, "https://", "", -1)

	// Verifica se tem porta especifica, se tiver pega ela e atribui a port
	if strings.Contains(proxy, ":") {
		proxy := strings.SplitN(proxy, ":", -1)
		portRest := strings.SplitN(proxy[1], "/", -1)
		portRest = strings.SplitN(portRest[0], "?", -1)
		portRest = strings.SplitN(portRest[0], "&", -1)
		portRest = strings.SplitN(portRest[0], "%", -1)

		host = proxy[0]
		port = portRest[0]
	} else {
		url := strings.SplitN(proxy, ":", -1)
		url = strings.SplitN(url[0], "/", -1)
		url = strings.SplitN(url[0], "?", -1)
		url = strings.SplitN(url[0], "&", -1)
		url = strings.SplitN(url[0], "%", -1)

		host = url[0]
	}

	return host, port
}

// TestConnectionPort - Used for test connection IP + PORT is OK.
func TestConnectionPort(host string, port string) bool {
	timeOut := time.Duration(10) * time.Second

	conn, err := net.DialTimeout("tcp", host+":"+port, timeOut)

	if err != nil {
		log.Printf("ERRO: Test Connection (%s:%s) - %s", host, port, err)

		return false
	}

	log.Printf("OK: Test Connection (%s:%s)", host, port)

	defer conn.Close()

	return true
}

// MakeHttpPostReq - Used for comunnication servers with POST
func MakeHttpPostReq(url string, proxypass string) string {

	client := http.Client{}

	var jsonprep string = "{\"proxypass\":\"" + proxypass + "\"}"

	var jsonStr = []byte(jsonprep)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)

		return "{\"msg\":\"Erro ao conectar na URL (" + url + ") para efetuar o POST\",\"status\":\"ERRO\"}"
	} else {
		body, _ := ioutil.ReadAll(resp.Body)

		return string(body)
	}
}

// MakeHTTPPostReqCreateFilesAPP - Usado para criar mapeamento no APP
func MakeHTTPPostReqCreateFilesAPP(url string, idArquivo, locationPrefix, location, proxyPass, consome, owner, tipo, dataHora string) string {

	client := http.Client{}

	var jsonprep = "{\"IDArquivo\":\"" + idArquivo + "\", \"LocationPrefix\":\"" + locationPrefix + "\" , \"Location\":\"" + location + "\", \"proxypass\":\"" + proxyPass + "\", \"AppConsome\":\"" + consome + "\", \"Owner\":\"" + owner + "\", \"Tipo\":\"" + tipo + "\", \"dataHora\":\"" + dataHora + "\"}"

	var jsonStr = []byte(jsonprep)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)

		return "{\"msg\":\"Erro ao conectar na URL (" + url + ") para efetuar o POST\",\"status\":\"ERRO\"}"
	} else {
		body, _ := ioutil.ReadAll(resp.Body)

		return string(body)
	}
}

// MakeHTTPPostReqDeleteFilesAPP - Delelta arquivo de mapeamento
func MakeHTTPPostReqDeleteFilesAPP(url string, idArquivo string) string {

	client := http.Client{}

	var jsonprep = "{\"IDArquivo\":\"" + idArquivo + "\"}"

	var jsonStr = []byte(jsonprep)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)

		return "{\"msg\":\"Erro ao conectar na URL (" + url + ") para efetuar o POST\",\"status\":\"ERRO\"}"
	} else {
		body, _ := ioutil.ReadAll(resp.Body)

		return string(body)
	}
}
