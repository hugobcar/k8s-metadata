package router

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	DB "github.com/hugobcar/k8s-metadata/models"
	netAPI "github.com/hugobcar/k8s-metadata/net"
)

type argumentosAPPStruct struct {
	LocationPrefix    string
	IDArquivo         string
	Location          string
	ProxyPass         string
	Env               string
	User              string
	Pass              string
	AppConsome        string
	Owner             string
	SomenteDMZ        string
	Tipo              string
	DataHora          string
	IgnorarTesteProxy string
}

type argumentosTestProxy struct {
	ProxyPass string
	Env       string
}

type argumentosAplicarProxy struct {
	User string
	Pass string
	Env  string
}

type argumentosDeletaMapsFiles struct {
	IDArquivo string
}

type jsonCreateMapsFilesStruct struct {
	ServerOrigem string `json:"serverOrigem"`
	Msg          string `json:"msg"`
	Status       string `json:"status"`
	MD5Sum       string `json:"md5sum"`
}

type jsonDeleteMapsFilesStruct struct {
	ServerOrigem string `json:"serverOrigem"`
	Msg          string `json:"msg"`
	Status       string `json:"status"`
}

type jsonRetornoPadrao struct {
	Msg    string `json:"msg"`
	Status string `json:"status"`
}

var wg sync.WaitGroup

// Gera MD5Sum de determinado arquivo
func md5sum(arquivo string) (md5Retorno string) {
	log.Println("arquivo: " + arquivo)

	file, err := os.Open(arquivo)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)

	md5Retorno = hex.EncodeToString(hash.Sum(nil))

	if err != nil {
		panic(err)

		md5Retorno = "ERRO"
	}

	return md5Retorno
}

func chamaGoRoutinesAplicarMapsAPP(w http.ResponseWriter, env, idArquivo, locationPrefix, location, proxyPass, consome, owner, tipo string) (temErro bool, md5Retorno string) {
	var servers []string

	// IPs dos servidores de APP
	var serversStage = DB.ReturnServers("STAGE", "APP")
	var serversDev = DB.ReturnServers("DEV", "APP")
	var serversProd = DB.ReturnServers("PROD", "APP")
	var serversTest = DB.ReturnServers("TEST", "APP")

	// IPs dos servidores da DMZ
	var serversDMZStage = DB.ReturnServers("STAGE", "DMZ")
	var serversDMZDev = DB.ReturnServers("DEV", "DMZ")
	var serversDMZProd = DB.ReturnServers("PROD", "DMZ")
	var serversDMZTest = DB.ReturnServers("TEST", "DMZ")

	servers = []string{""}

	d := time.Now()
	dataHora := d.String()

	// tipo - APP ou DMZ
	if tipo == "APP" {
		switch env {
		case "STAGE":
			servers = serversStage
		case "DEV":
			servers = serversDev
		case "TEST":
			servers = serversTest
		case "PROD":
			servers = serversProd
		}
	} else {
		switch env {
		case "STAGE":
			servers = serversDMZStage
		case "DEV":
			servers = serversDMZDev
		case "TEST":
			servers = serversDMZTest
		case "PROD":
			servers = serversDMZProd
		}
	}

	// Aloca um processador logico para o escalonador usar
	runtime.GOMAXPROCS(1)

	// wg é usada para esperar o programa terminar
	// Adiciona um contador igual o numero de valores de array para as goroutines
	wg.Add(len(servers))

	log.Println("Iniciando GoRoutines (AplicarMaps)- " + tipo)

	//	var msg string
	var c = make(chan string, len(servers))
	var md5 = make(chan string, len(servers))

	// Efetua testes das URLs em todos os servidores
	for i := 0; i < len(servers); i++ {
		go aplicarMapsAPP(i, w, idArquivo, locationPrefix, location, proxyPass, consome, owner, servers[i], tipo, dataHora, c, md5)
	}

	// Verifica se retornou algum erro nas mensagens dos canais
	for x := 0; x < len(servers); x++ {
		if <-c == "ERRO" {
			temErro = true
			md5Retorno = "ERRO"
		} else {
			md5Retorno = <-md5
		}
	}

	wg.Wait()
	log.Println("GoRoutines (AplicarMaps) - " + tipo + " finalizadas")

	return temErro, md5Retorno
}

func chamaGoRoutinesDeletarMapsAPP(w http.ResponseWriter, env, idArquivo, tipo string) (temErro bool) {
	var servers []string

	// IPs dos servidores de APP
	var serversStage = DB.ReturnServers("STAGE", "APP")
	var serversDev = DB.ReturnServers("DEV", "APP")
	var serversProd = DB.ReturnServers("PROD", "APP")
	var serversTest = DB.ReturnServers("TEST", "APP")

	// IPs dos servidores da DMZ
	var serversDMZStage = DB.ReturnServers("STAGE", "DMZ")
	var serversDMZDev = DB.ReturnServers("DEV", "DMZ")
	var serversDMZProd = DB.ReturnServers("PROD", "DMZ")
	var serversDMZTest = DB.ReturnServers("TEST", "DMZ")
	servers = []string{""}

	// tipo - APP ou DMZ
	if tipo == "APP" {
		switch env {
		case "STAGE":
			servers = serversStage
		case "DEV":
			servers = serversDev
		case "TEST":
			servers = serversTest
		case "PROD":
			servers = serversProd
		}
	} else {
		switch env {
		case "STAGE":
			servers = serversDMZStage
		case "DEV":
			servers = serversDMZDev
		case "TEST":
			servers = serversDMZTest
		case "PROD":
			servers = serversDMZProd
		}
	}

	// Aloca um processador logico para o escalonador usar
	runtime.GOMAXPROCS(1)

	// wg é usada para esperar o programa terminar
	// Adiciona um contador igual o numero de valores de array para as goroutines
	wg.Add(len(servers))

	log.Println("Iniciando GoRoutines (DeletarMapsAPP)")

	//	var msg string
	var c = make(chan string, len(servers))

	// Efetua testes das URLs em todos os servidores
	for i := 0; i < len(servers); i++ {
		go deletarMapsAPP(i, w, idArquivo, servers[i], c)
	}

	// Verifica se retornou algum erro nas mensagens dos canais
	for x := 0; x < len(servers); x++ {
		if <-c == "ERRO" {
			temErro = true
		}
	}

	wg.Wait()
	log.Println("GoRoutines (DeletarMapsAPP) finalizadas")

	return temErro
}

func chamaGoRoutinesTestProxyAPP(w http.ResponseWriter, proxypass string, env string) (temErro, msgRetorno string) {
	var servers []string

	// IPs dos servidores de APP
	var serversStage = DB.ReturnServers("STAGE", "APP")
	var serversDev = DB.ReturnServers("DEV", "APP")
	var serversProd = DB.ReturnServers("PROD", "APP")
	var serversTest = DB.ReturnServers("TEST", "APP")

	servers = []string{""}

	switch env {
	case "STAGE":
		servers = serversStage
	case "DEV":
		servers = serversDev
	case "TEST":
		servers = serversTest
	case "PROD":
		servers = serversProd
	}

	// Aloca um processador logico para o escalonador usar
	runtime.GOMAXPROCS(1)

	// wg é usada para esperar o programa terminar
	// Adiciona um contador igual o numero de valores de array para as goroutines
	wg.Add(len(servers))

	log.Println("Iniciando GoRoutines (testProxyAPP)")

	var msgCanal string
	var servidores string

	var c = make(chan string, len(servers))
	var cMsg = make(chan string, len(servers))

	// Efetua testes das URLs em todos os servidores
	for i := 0; i < len(servers); i++ {
		go testProxyAPP(i, w, proxypass, servers[i], c, cMsg)
	}

	temErro = "OK"

	// Verifica se retornou algum erro nas mensagens dos canais
	for x := 0; x < len(servers); x++ {
		msgCanal = <-cMsg
		servidores = servidores + " - " + servers[x]

		if <-c == "ERRO" {
			temErro = "ERRO"

			msgRetorno = msgCanal
		}
	}

	msgRetorno = msgRetorno + " -- Servidores de Origem: (" + servidores + ")"

	wg.Wait()
	log.Println("GoRoutines (testProxyAPP) finalizadas")

	return temErro, msgRetorno
}

func aplicarMapsAPP(i int, w http.ResponseWriter, idArquivo, locationPrefix, location, proxyPass, consome, owner, server, tipo, dataHora string, c, md5 chan string) {
	var resp1 jsonCreateMapsFilesStruct

	defer wg.Done()

	log.Println("Chamada POST: http://" + server + ":6885/CreateMapsFiles")

	respMakeHTTPPostReqCreateFilesAPP := netAPI.MakeHTTPPostReqCreateFilesAPP("http://"+server+":6885/CreateMapsFiles", idArquivo, locationPrefix, location, proxyPass, consome, owner, tipo, dataHora)

	err := json.NewDecoder(strings.NewReader(respMakeHTTPPostReqCreateFilesAPP)).Decode(&resp1)
	if err != nil {
		fmt.Println(err)
		c <- "ERRO"
		md5 <- "ERRO"
	}

	resp1.ServerOrigem = server

	if resp1.Status == "ERRO" {
		c <- "ERRO"
		md5 <- "ERRO"
	} else {
		c <- "OK"
		md5 <- resp1.MD5Sum
	}
}

func deletarMapsAPP(i int, w http.ResponseWriter, idArquivo, server string, c chan string) {
	var resp1 jsonCreateMapsFilesStruct

	defer wg.Done()

	log.Println("Chamada POST: http://" + server + ":6885/DeleteMapsFiles")

	respMakeHTTPPostReqDeleteFilesAPP := netAPI.MakeHTTPPostReqDeleteFilesAPP("http://"+server+":6885/DeleteMapsFiles", idArquivo)

	err := json.NewDecoder(strings.NewReader(respMakeHTTPPostReqDeleteFilesAPP)).Decode(&resp1)
	if err != nil {
		fmt.Println(err)
		c <- "ERRO"
	}

	resp1.ServerOrigem = server

	if resp1.Status == "ERRO" {
		c <- "ERRO"
	} else {
		c <- "OK"
	}
}

func testProxyAPP(i int, w http.ResponseWriter, proxypass, server string, c, cMsg chan string) {
	var resp1 jsonCreateMapsFilesStruct

	defer wg.Done()

	log.Println("Chamada POST: http://" + server + ":6885/TestProxy")

	respMakeHTTPPostReq := netAPI.MakeHttpPostReq("http://"+server+":6885/TestProxy", proxypass)

	err := json.NewDecoder(strings.NewReader(respMakeHTTPPostReq)).Decode(&resp1)
	if err != nil {
		fmt.Println(err)
		c <- "ERRO"
	}

	resp1.ServerOrigem = server

	if resp1.Status == "ERRO" {
		c <- "ERRO"
		cMsg <- resp1.Msg
	} else {
		c <- "OK"
		cMsg <- resp1.Msg
	}
}

// Funcao para checar se Location tem alguns caracteres especiais que nao sao permitidos
func checkFormatacaoLocation(location string) (retorno bool) {
	retorno = strings.ContainsAny(location, "| & & & ; & * & ( & ) & % & ? & ' ' & ç & ` & ' & ~ & ^ & ] & [ & { & } & \\ & : & > & < & . & , & ã")

	return retorno
}

// Funcao para checar se Location tem alguns caracteres especiais que nao sao permitidos
func checkFormatacaoProxyPass(proxyPass string) (retorno bool) {
	retorno = strings.ContainsAny(proxyPass, ";")

	return retorno
}

// Create Files
func createFile(path string) {
	// detect if file exists
	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		checkError(err)
		defer file.Close()
	}
}

func writeFile(path string, locationPrefix, location, proxyPass, owner, consumir, tipo, dataHora string) {
	// tipo - APP ou DMZ
	if tipo == "DMZUpStream" {
		proxyPass = "http://UPSTREAM/" + locationPrefix + "/" + location
	}

	// retira ponto e virgula caso tenha
	location = strings.Replace(location, ";", "", -1)

	// open file using READ & WRITE permission
	var file, err = os.OpenFile(path, os.O_RDWR, 0644)
	checkError(err)
	defer file.Close()

	// write some text to file
	_, err = file.WriteString("##################################################################\n")
	checkError(err)
	_, err = file.WriteString("# APP a consumir.......: " + consumir + "\n")
	checkError(err)
	_, err = file.WriteString("# Responsavel..........: " + owner + "\n")
	checkError(err)
	_, err = file.WriteString("# Proxy criado por.....: k8s-metadata\n")
	checkError(err)
	_, err = file.WriteString("# Data.................: " + dataHora + "\n")
	checkError(err)
	_, err = file.WriteString("##################################################################\n\n")
	checkError(err)
	_, err = file.WriteString("location /" + locationPrefix + "/" + location + " {\n")
	checkError(err)
	_, err = file.WriteString("\tproxy_pass " + proxyPass + ";\n")
	checkError(err)
	_, err = file.WriteString("\tproxy_connect_timeout 600;\n")
	checkError(err)
	_, err = file.WriteString("\tproxy_send_timeout 600;\n")
	checkError(err)
	_, err = file.WriteString("\tproxy_read_timeout 600; \n")
	checkError(err)
	_, err = file.WriteString("}\n\n")
	checkError(err)

	// save changes
	err = file.Sync()
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
}

///////////////////////////////////////////////////////////////////////////////////
/////////////////////////////// Functions - APP ///////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////

// GetClusters - Return Clusters List
func GetClusters(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// DeleteMapsFiles - Deleta arquivos de mapeamento
func DeleteMapsFiles(w http.ResponseWriter, r *http.Request) {
	var argsDeletaMapsFiles argumentosDeletaMapsFiles

	decoder := json.NewDecoder(r.Body)

	var resp jsonDeleteMapsFilesStruct

	err := decoder.Decode(&argsDeletaMapsFiles)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	// Valida para verificar se todos os argumentos foram passados via POST
	if argsDeletaMapsFiles.IDArquivo == "" {
		resp = jsonDeleteMapsFilesStruct{Msg: "Faltando argumentos POST", Status: "ERRO"}
	} else {
		nomeArquivoDelete := "/etc/nginx/conf.d/maps/k8s-metadata_" + argsDeletaMapsFiles.IDArquivo + ".conf"

		// Apaga arquivo
		cmd := exec.Command("rm", "-f", nomeArquivoDelete)
		err := cmd.Run()
		if err != nil {
			log.Println("Erro ao apagar o arquivo: ", err)

			resp = jsonDeleteMapsFilesStruct{Msg: "Erro ao apagar o arquivo: " + nomeArquivoDelete, Status: "ERRO"}
		} else {
			log.Println("Arquivo (" + nomeArquivoDelete + ") apagado COM SUCESSO")
		}

		// Restart Nginx
		cmd1 := exec.Command("/etc/init.d/nginx", "restart")
		err1 := cmd1.Run()
		if err1 != nil {
			log.Println("Erro ao dar reload no Nginx: ", err)

			resp = jsonDeleteMapsFilesStruct{Msg: "ERRO ao dar RESTART no Nginx", Status: "ERRO"}
		} else {
			log.Println("/etc/init.d/nginx restart")
		}

	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		panic(err)
	}
}

// TestProxy - Use for test URL of Proxy Pass
func TestProxy(w http.ResponseWriter, r *http.Request) {
	var argsTestProxy argumentosTestProxy

	decoder := json.NewDecoder(r.Body)

	var resp jsonCreateMapsFilesStruct

	err := decoder.Decode(&argsTestProxy)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	// Valida para verificar se todos os argumentos foram passados via POST
	if argsTestProxy.ProxyPass == "" {
		resp = jsonCreateMapsFilesStruct{Msg: "Faltando argumentos POST", Status: "ERRO"}
	} else {
		if netAPI.ValidaURLProxyPass(argsTestProxy.ProxyPass) == false {
			resp = jsonCreateMapsFilesStruct{Msg: "Erro ao validar endereço de Proxy. Favor usar (http://) ou (https://) no começo.", Status: "ERRO"}
		} else {
			host, port := netAPI.ReturnHostandPortURL(argsTestProxy.ProxyPass)
			var Message, Status string

			if netAPI.TestConnectionPort(host, port) {
				Message = "Test Connection (" + host + ":" + port + ")"
				Status = "OK"
			} else {
				Message = "Test Connection (" + host + ":" + port + ")"
				Status = "ERRO"
			}

			resp = jsonCreateMapsFilesStruct{Msg: Message, Status: Status}

		}
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		panic(err)
	}
}

///////////////////////////////////////////////////////////////////////////////////
/////////////////////////////// Functions - DMZ ///////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////

// CreateMapsFilesDMZ - Create archives maps Nginx na DMZ
func CreateMapsFilesDMZ(w http.ResponseWriter, r *http.Request) {
	var argsCreateMapsFiles argumentosAPPStruct

	decoder := json.NewDecoder(r.Body)

	var resp jsonCreateMapsFilesStruct

	err := decoder.Decode(&argsCreateMapsFiles)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	// Valida para verificar se todos os argumentos foram passados via POST
	if argsCreateMapsFiles.IDArquivo == "" || argsCreateMapsFiles.LocationPrefix == "" || argsCreateMapsFiles.Location == "" || argsCreateMapsFiles.ProxyPass == "" || argsCreateMapsFiles.AppConsome == "" || argsCreateMapsFiles.Owner == "" || argsCreateMapsFiles.Tipo == "" || argsCreateMapsFiles.DataHora == "" {
		resp = jsonCreateMapsFilesStruct{Msg: "Faltando argumentos POST", Status: "ERRO"}
	} else {
		nomeArquivo := "/etc/nginx/conf.d/maps/k8s-metadata_" + argsCreateMapsFiles.IDArquivo + ".conf"
		createFile(nomeArquivo)
		writeFile(nomeArquivo, argsCreateMapsFiles.LocationPrefix, argsCreateMapsFiles.Location, argsCreateMapsFiles.ProxyPass, argsCreateMapsFiles.Owner, argsCreateMapsFiles.AppConsome, argsCreateMapsFiles.Tipo, argsCreateMapsFiles.DataHora)

		md5Arquivo := md5sum(nomeArquivo)

		log.Println("Arquivo criado: " + nomeArquivo + " -- MD5Sum: " + md5Arquivo)

		resp = jsonCreateMapsFilesStruct{Msg: "Mapeamentos criados COM SUCESSO!", Status: "OK", MD5Sum: md5Arquivo}

		// Reload Nginx
		cmd := exec.Command("/etc/init.d/nginx", "reload")
		err := cmd.Run()
		if err != nil {
			log.Println("Erro ao dar reload no Nginx: ", err)

			resp = jsonCreateMapsFilesStruct{Msg: "ERRO ao dar RELOAD no Nginx", Status: "ERRO", MD5Sum: "ERRO"}
		} else {
			log.Println("/etc/init.d/nginx reload")
		}
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		panic(err)
	}
}

// DeleteMapsFilesDMZ - Deleta arquivos de mapeamento DMZ
func DeleteMapsFilesDMZ(w http.ResponseWriter, r *http.Request) {
	var argsDeletaMapsFiles argumentosDeletaMapsFiles

	decoder := json.NewDecoder(r.Body)

	var resp jsonDeleteMapsFilesStruct

	err := decoder.Decode(&argsDeletaMapsFiles)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	// Valida para verificar se todos os argumentos foram passados via POST
	if argsDeletaMapsFiles.IDArquivo == "" {
		resp = jsonDeleteMapsFilesStruct{Msg: "Faltando argumentos POST", Status: "ERRO"}
	} else {
		nomeArquivoDelete := "/etc/nginx/conf.d/maps/k8s-metadata_" + argsDeletaMapsFiles.IDArquivo + ".conf"

		// Apaga arquivo
		cmd := exec.Command("rm", "-f", nomeArquivoDelete)
		err := cmd.Run()
		if err != nil {
			log.Println("Erro ao apagar o arquivo: ", err)

			resp = jsonDeleteMapsFilesStruct{Msg: "Erro ao apagar o arquivo: " + nomeArquivoDelete, Status: "ERRO"}
		} else {
			log.Println("Arquivo (" + nomeArquivoDelete + ") apagado COM SUCESSO")
		}

		// Restart Nginx
		cmd1 := exec.Command("/etc/init.d/nginx", "restart")
		err1 := cmd1.Run()
		if err1 != nil {
			log.Println("Erro ao dar reload no Nginx: ", err)

			resp = jsonDeleteMapsFilesStruct{Msg: "ERRO ao dar RESTART no Nginx", Status: "ERRO"}
		} else {
			log.Println("/etc/init.d/nginx restart")
		}

	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		panic(err)
	}
}

///////////////////////////////////////////////////////////////////////////////////
//////////////////////////// Functions - SERVER ///////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////

// CreateMaps - Insert maps to DB for be create APP / DMZ
func CreateMaps(w http.ResponseWriter, r *http.Request) {
	var args argumentosAPPStruct

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&args)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	// Converte o Env e LocationPrefix tudo em MAIUSCULO
	args.Env = strings.ToUpper(args.Env)
	args.LocationPrefix = strings.ToUpper(args.LocationPrefix)

	if args.LocationPrefix == "" || args.Location == "" || args.ProxyPass == "" || args.Env == "" || args.User == "" || args.Pass == "" || args.AppConsome == "" || args.Owner == "" || args.SomenteDMZ == "" || args.IgnorarTesteProxy == "" {
		resp := jsonRetornoPadrao{Msg: "Faltando argumentos POST", Status: "ERRO"}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			panic(err)
		}
	} else {
		if args.SomenteDMZ != "sim" && args.SomenteDMZ != "nao" {
			resp := jsonRetornoPadrao{Msg: "SomenteDMZ=sim (Cria mapeamento somente na DMZ, utilizado para APIs internas) OU SomenteDMZ=nao (Cria mapeamento na DMZ [UPSTREAM] apontando para servers APP)", Status: "ERRO"}

			if err := json.NewEncoder(w).Encode(resp); err != nil {
				panic(err)
			}
		} else {
			if DB.ValidaUser(args.User, args.Pass, "0") {
				var testProxy string
				var msgProxy string

				// Antes de fazer os procedimentos abaixo, checa para verificar se tem algum servidor cadastrado e habilitado
				if DB.TemServerDMZAPP(args.Env) {
					// Se mapeamento for somente na DMZ, nao testar Proxy
					if args.SomenteDMZ == "sim" {
						testProxy = "OK"
					} else {
						if args.IgnorarTesteProxy == "sim" {
							testProxy = "OK"
						} else {
							testProxy, msgProxy = chamaGoRoutinesTestProxyAPP(w, args.ProxyPass, args.Env)
						}
					}

					if testProxy == "ERRO" {
						resp := jsonRetornoPadrao{Msg: msgProxy, Status: "ERRO"}

						log.Println("ERRO: " + msgProxy)

						if err := json.NewEncoder(w).Encode(resp); err != nil {
							panic(err)
						}
					} else {
						log.Println("OK: " + msgProxy)

						// Verificando formatacao do Location - Nao deve conter espacos ou qualquer tipo de caracter especial
						if checkFormatacaoLocation(args.Location) {
							resp := jsonRetornoPadrao{Msg: "Location contem caracteres nao permitidos. Favor retira-los.", Status: "ERRO"}

							if err := json.NewEncoder(w).Encode(resp); err != nil {
								panic(err)
							}
						} else {
							if checkFormatacaoProxyPass(args.ProxyPass) {
								resp := jsonRetornoPadrao{Msg: "Proxy Pass contem caracteres nao permitidos. Favor retira-los.", Status: "ERRO"}

								if err := json.NewEncoder(w).Encode(resp); err != nil {
									panic(err)
								}
							} else {
								// Transforma Location tudo em minusculo
								//args.Location = strings.ToLower(args.Location)

								// Verificando se Location ja esta criado
								if DB.LocationExist(args.LocationPrefix, args.Location, args.Env) {
									resp := jsonRetornoPadrao{Msg: "Location ja esta criada na base (/" + args.LocationPrefix + "/" + args.Location + " -- " + args.Env + ")", Status: "ERRO"}

									if err := json.NewEncoder(w).Encode(resp); err != nil {
										panic(err)
									}
								} else {
									idUser, idGroup := DB.ReturnIDUserAndIDUser(args.User)
									DB.CreateMapDB(args.Location, args.ProxyPass, args.Env, idUser, idGroup, args.AppConsome, args.Owner, args.LocationPrefix, args.SomenteDMZ)

									resp := jsonRetornoPadrao{Msg: "Mapeamento (/" + args.LocationPrefix + "/" + args.Location + " -- " + args.Env + ") criado COM SUCESSO! Favor pedir para que administradores aplique esse mapeamento nos servidores.", Status: "OK"}

									if err := json.NewEncoder(w).Encode(resp); err != nil {
										panic(err)
									}
								}
							}
						}
					}
				} else {
					resp := jsonRetornoPadrao{Msg: "Nenhum servidor cadastrado ou habilitado para esse ambiente!", Status: "ERRO"}

					if err := json.NewEncoder(w).Encode(resp); err != nil {
						panic(err)
					}
				}
			} else {
				resp := jsonRetornoPadrao{Msg: "ERRO ao tentar autenticar usuario!", Status: "ERRO"}

				if err := json.NewEncoder(w).Encode(resp); err != nil {
					panic(err)
				}
			}
		}
	}
}

// TestProxyServer - Used for connect APPs for test Proxy Pass
func TestProxyServer(w http.ResponseWriter, r *http.Request) {
	var resp jsonRetornoPadrao
	var argsTestProxy argumentosTestProxy

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&argsTestProxy)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	// Valida para verificar se todos os argumentos foram passados via POST
	if argsTestProxy.Env == "" || argsTestProxy.ProxyPass == "" {
		resp = jsonRetornoPadrao{Msg: "Faltando argumentos POST", Status: "ERRO"}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			panic(err)
		}
	} else {
		// Antes de fazer os procedimentos abaixo, checa para verificar se tem algum servidor cadastrado e habilitado
		if DB.TemServerDMZAPP(argsTestProxy.Env) {
			// Converte o Env tudo em MAIUSCULO
			argsTestProxy.Env = strings.ToUpper(argsTestProxy.Env)

			var testProxy string
			var msgProxy string

			testProxy, msgProxy = chamaGoRoutinesTestProxyAPP(w, argsTestProxy.ProxyPass, argsTestProxy.Env)

			if testProxy == "ERRO" {
				resp = jsonRetornoPadrao{Msg: msgProxy, Status: "ERRO"}

				if err := json.NewEncoder(w).Encode(resp); err != nil {
					panic(err)
				}
			} else {
				resp = jsonRetornoPadrao{Msg: "Teste de conexão realizado COM SUCESSO!", Status: "OK"}

				if err := json.NewEncoder(w).Encode(resp); err != nil {
					panic(err)
				}
			}
		} else {
			resp := jsonRetornoPadrao{Msg: "Nenhum servidor cadastrado ou habilitado para esse ambiente!", Status: "ERRO"}

			if err := json.NewEncoder(w).Encode(resp); err != nil {
				panic(err)
			}
		}
	}
}

// AplicarMapsAPP - Criar mapeamento em todos os servidores
func AplicarMapsAPP(w http.ResponseWriter, r *http.Request) {
	var resp1 jsonCreateMapsFilesStruct
	var argsAplicarProxy argumentosAplicarProxy

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&argsAplicarProxy)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	// Valida para verificar se todos os argumentos foram passados via POST

	if DB.ValidaUser(argsAplicarProxy.User, argsAplicarProxy.Pass, "1") {
		if argsAplicarProxy.User == "" || argsAplicarProxy.Pass == "" || argsAplicarProxy.Env == "" {
			resp1 = jsonCreateMapsFilesStruct{Msg: "Faltando argumentos POST", Status: "ERRO"}

			if err := json.NewEncoder(w).Encode(resp1); err != nil {
				panic(err)
			}
		} else {
			// Antes de fazer os procedimentos abaixo, checa para verificar se tem algum servidor cadastrado e habilitado
			if DB.TemServerDMZAPP(argsAplicarProxy.Env) {
				var IDsProxy []int
				var locationPrefix []string
				var location []string
				var proxyPass []string
				var consome []string
				var owner []string
				var somenteDMZ []int
				var aplicado []int

				var IDArq string

				var temErro bool
				var temMapeamentosAplicar bool

				IDsProxy, locationPrefix, location, proxyPass, consome, owner, somenteDMZ, aplicado = DB.ReturnMapeamentosCriarAndDeleter(argsAplicarProxy.Env)

				for i := 0; i < len(IDsProxy); i++ {
					temMapeamentosAplicar = true

					IDArq = strconv.Itoa(IDsProxy[i])
					var apagarRegistroDB = "nao"

					// --------- DELETE OS MAPEAMENTOS ---------
					if aplicado[i] == 3 {
						log.Println("Apagando APP e DMZ (" + IDArq + ")")
						// Delete mapeamento APP
						if chamaGoRoutinesDeletarMapsAPP(w, argsAplicarProxy.Env, IDArq, "APP") {
							log.Println("ERRO ao deletar mapeamento APP (ID: " + IDArq + " -- Location: /" + locationPrefix[i] + "/" + location[i] + " -- ProxyPass: " + proxyPass[i] + " -- Env: " + argsAplicarProxy.Env)

							apagarRegistroDB = "nao"
						}

						// Delete mapeamento DMZ
						if chamaGoRoutinesDeletarMapsAPP(w, argsAplicarProxy.Env, IDArq, "DMZ") {
							log.Println("ERRO ao deletar mapeamento APP (ID: " + IDArq + " -- Location: /" + locationPrefix[i] + "/" + location[i] + " -- ProxyPass: " + proxyPass[i] + " -- Env: " + argsAplicarProxy.Env)

							apagarRegistroDB = "nao"
						} else {
							apagarRegistroDB = "sim"
						}

						// Se tudo OK, apaga registro no Banco de Dados
						if apagarRegistroDB == "sim" {
							temErro = false

							if DB.DeleteMapDB(IDsProxy[i], argsAplicarProxy.Env) {
								log.Println("Registro (ID: " + IDArq + " -- Location: " + locationPrefix[i] + "/" + location[i] + " -- ProxyPass: " + proxyPass[i] + " -- Env: " + argsAplicarProxy.Env + ") foi apagado do banco COM SUCESSO!")
							} else {
								log.Println("ERRO ao apagar registro (ID: " + IDArq + " -- Location: /" + locationPrefix[i] + "/" + location[i] + " -- ProxyPass: " + proxyPass[i] + " -- Env: " + argsAplicarProxy.Env + ") do banco!")

								temErro = true
							}
						} else {
							temErro = true
						}
					} else {
						// --------- CRIA OS MAPEAMENTOS ---------
						if somenteDMZ[i] == 1 {
							/////////////////////////////////////////////////////////////////////////////////////
							//////// Cria os mapeamentos somente nos servidores da DMZ - utilizado para quando usamos nossos servers de APP
							/////////////////////////////////////////////////////////////////////////////////////
							respAplicarMapsDMZ, md5FileDMZ := chamaGoRoutinesAplicarMapsAPP(w, argsAplicarProxy.Env, IDArq, locationPrefix[i], location[i], proxyPass[i], consome[i], owner[i], "DMZ")

							if respAplicarMapsDMZ {
								log.Println("ERRO ao criar mapeamento DMZ (ID: " + IDArq + " -- Location: /" + locationPrefix[i] + "/" + location[i] + " -- ProxyPass: " + proxyPass[i] + " -- Env: " + argsAplicarProxy.Env)

								// Alterado o status "aplicado" do mapeamento para ERRO
								DB.UpdateAplicadoMapeamento(IDsProxy[i], 2)

								// Delete mapeamento de arquivos que deram problemas
								if chamaGoRoutinesDeletarMapsAPP(w, argsAplicarProxy.Env, IDArq, "DMZ") {
									log.Println("ERRO ao deletar mapeamento DMZ (ID: " + IDArq + " -- Location: /" + locationPrefix[i] + "/" + location[i] + " -- ProxyPass: " + proxyPass[i] + " -- Env: " + argsAplicarProxy.Env)
								}

								temErro = true
							} else {
								if md5FileDMZ == "" {
									temMapeamentosAplicar = false
								} else {
									log.Println("Mapeamento DMZ (ID: " + IDArq + "[" + md5FileDMZ + "]" + " -- Location: /" + locationPrefix[i] + "/" + location[i] + " -- ProxyPass: " + proxyPass[i] + " -- Env: " + argsAplicarProxy.Env + " criados COM SUCESSO!")

									// Alterado o status "aplicado" do mapeamento para COM SUCESSO
									DB.UpdateAplicadoMapeamento(IDsProxy[i], 1)

									DB.UpdateMD5Mapeamento(IDsProxy[i], md5FileDMZ, "DMZ")

									temErro = false
								}
							}
						} else {
							/////////////////////////////////////////////////////////////////////////////////////
							//////// Cria os mapeamentos nos servidores da DMZ (apontando para UPSTREAM) e cria os mapeamentos nos servidores de APP - utilizado para quando queremos direcionar da DMZ para APP e do APP para server externo
							/////////////////////////////////////////////////////////////////////////////////////
							respAplicarMaps, md5FileApp := chamaGoRoutinesAplicarMapsAPP(w, argsAplicarProxy.Env, IDArq, locationPrefix[i], location[i], proxyPass[i], consome[i], owner[i], "APP")

							if respAplicarMaps {
								log.Println("ERRO ao criar mapeamento APP (ID: " + IDArq + " -- Location: /" + locationPrefix[i] + "/" + location[i] + " -- ProxyPass: " + proxyPass[i] + " -- Env: " + argsAplicarProxy.Env)

								// Alterado o status "aplicado" do mapeamento para ERRO
								DB.UpdateAplicadoMapeamento(IDsProxy[i], 2)

								// Delete mapeamento de arquivos que deram problemas
								if chamaGoRoutinesDeletarMapsAPP(w, argsAplicarProxy.Env, IDArq, "APP") {
									log.Println("ERRO ao deletar mapeamento APP (ID: " + IDArq + " -- Location: /" + locationPrefix[i] + "/" + location[i] + " -- ProxyPass: " + proxyPass[i] + " -- Env: " + argsAplicarProxy.Env)
								}

								temErro = true
							} else {
								respAplicarMapsDMZUP, md5FileAppDMZUP := chamaGoRoutinesAplicarMapsAPP(w, argsAplicarProxy.Env, IDArq, locationPrefix[i], location[i], proxyPass[i], consome[i], owner[i], "DMZUpStream")

								if respAplicarMapsDMZUP {
									log.Println("ERRO ao criar mapeamento DMZ (" + IDArq + ") http://UPSTREAM/" + locationPrefix[i] + "/" + location[i])

									// Alterado o status "aplicado" do mapeamento para ERRO
									DB.UpdateAplicadoMapeamento(IDsProxy[i], 2)

									// Delete mapeamento de arquivos que deram problemas
									if chamaGoRoutinesDeletarMapsAPP(w, argsAplicarProxy.Env, IDArq, "DMZ") {
										log.Println("ERRO ao deletar mapeamento DMZ (" + IDArq + ") http://UPSTREAM/" + locationPrefix[i] + "/" + location[i])
									}

									// Delete mapeamento de arquivos que deram problemas
									if chamaGoRoutinesDeletarMapsAPP(w, argsAplicarProxy.Env, IDArq, "APP") {
										log.Println("ERRO ao deletar mapeamento APP (ID: " + IDArq + " -- Location: /" + locationPrefix[i] + "/" + location[i] + " -- ProxyPass: " + proxyPass[i] + " -- Env: " + argsAplicarProxy.Env)
									}

									temErro = true
								} else {
									if md5FileApp == "" || md5FileAppDMZUP == "" {
										temMapeamentosAplicar = false
									} else {
										log.Println("Mapeamento APP e DMZ (ID: " + IDArq + "[APP: " + md5FileApp + " -- DMZ: " + md5FileAppDMZUP + "]" + " -- Location: /" + locationPrefix[i] + "/" + location[i] + " -- ProxyPass: " + proxyPass[i] + " -- Env: " + argsAplicarProxy.Env + " criados COM SUCESSO!")

										// Alterado o status "aplicado" do mapeamento para COM SUCESSO
										DB.UpdateAplicadoMapeamento(IDsProxy[i], 1)

										DB.UpdateMD5Mapeamento(IDsProxy[i], md5FileApp, "APP")
										DB.UpdateMD5Mapeamento(IDsProxy[i], md5FileAppDMZUP, "DMZ")

										temErro = false
									}
								}
							}
						}
					}
				}

				// So mostra msg de ERRO ou SUCESSO caso tenha algum mapeamento a ser aplicado (Criado ou Deletado)
				if temMapeamentosAplicar {
					// Reporta mensagem com ERRO ou SUCESSO
					if temErro {
						resp := jsonRetornoPadrao{Msg: "ERRO ao aplicar mapeamentos. Verificar logs!", Status: "ERRO"}

						if err := json.NewEncoder(w).Encode(resp); err != nil {
							panic(err)
						}
					} else {
						resp := jsonRetornoPadrao{Msg: "Mapeamentos aplicados COM SUCESSO!", Status: "OK"}

						if err := json.NewEncoder(w).Encode(resp); err != nil {
							panic(err)
						}
					}
				}
			} else {
				resp := jsonRetornoPadrao{Msg: "Nenhum servidor cadastrado ou habilitado para esse ambiente!", Status: "ERRO"}

				if err := json.NewEncoder(w).Encode(resp); err != nil {
					panic(err)
				}
			}
		}
	} else {
		resp := jsonRetornoPadrao{Msg: "ERRO ao tentar autenticar usuário! Senha errada ou usuário não é Administrador.", Status: "ERRO"}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			panic(err)
		}
	}
}
