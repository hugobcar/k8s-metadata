package models

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// Variaveis de conexao com o Banco
var UserDB string
var PassDB string
var DatabaseDB string
var HostDB string
var PortDB string

// ValidaUser - Used for check user and password is OK
func ValidaUser(user, password, tipoUser string) (retorno bool) {
	var SQLSelect string

	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	// Se tipoUser = 0 quer dizer que nao eh necessario ser admin para efetuar validacao
	if tipoUser == "0" {
		SQLSelect = "SELECT COUNT(nome_usuario) FROM usuarios WHERE ativo=1 AND usuario='" + user + "' AND senha='" + password + "'"
	} else {
		SQLSelect = "SELECT COUNT(nome_usuario) FROM usuarios WHERE ativo=1 AND usuario='" + user + "' AND senha='" + password + "' AND admin='" + tipoUser + "'"
	}

	rows, err := db.Query(SQLSelect)

	checkErr(err)

	for rows.Next() {
		var count int
		err = rows.Scan(&count)
		checkErr(err)

		if count >= 1 {
			return true
		} else {
			return false
		}
	}

	return retorno
}

// LocationExist - Verifica se Location ja existe
func LocationExist(locationPrefix, location, env string) (retorno bool) {
	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	// query
	rows, err := db.Query("SELECT COUNT(id_mapeamento) FROM mapeamentos WHERE location_prefix='" + locationPrefix + "' AND location='" + location + "' AND env='" + env + "'")
	checkErr(err)

	for rows.Next() {
		var count int
		err = rows.Scan(&count)
		checkErr(err)

		if count >= 1 {
			return true
		} else {
			return false
		}
	}

	return retorno
}

// ReturnIDUserAndIDUser - Verifica se Location ja existe
func ReturnIDUserAndIDUser(usuario string) (idUser, idGroup int) {
	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	// query
	rows, err := db.Query("SELECT id_usuario, id_grupo FROM usuarios WHERE usuario='" + usuario + "'")
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&idUser, &idGroup)
		checkErr(err)
	}

	return idUser, idGroup
}

// ReturnMapeamentosCriarAndDeleter - Retorna todos os mapeamentos que estao aprovados para criar e deletar
func ReturnMapeamentosCriarAndDeleter(env string) (idM []int, locationPrefix, location, proxyPass, consome, owner []string, somenteDMZ, aplicado []int) {
	var SidM, SDMZ, Saplicado int
	var SlocationPrefix, Slocation, SproxyPass, Sconsome, Sowner string

	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	// aplicado=0 - Ainda nao foi criado mapeamento
	// aplicado=1 - Mapeamento foi criado COM SUCESSO
	// aplicado=2 - Erro ao criar mapeamento, ele nao tentara ser criado novamente
	// aplicado=3 - Mapeamentos marcados para serem deletados
	rows, err := db.Query("SELECT id_mapeamento, location_prefix, location, proxy_pass, consome, owner, somente_dmz, aplicado FROM mapeamentos WHERE env='" + env + "' AND aprovado=1 AND (aplicado=0 OR aplicado=3)")
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&SidM, &SlocationPrefix, &Slocation, &SproxyPass, &Sconsome, &Sowner, &SDMZ, &Saplicado)
		checkErr(err)

		idM = append(idM, SidM)
		locationPrefix = append(locationPrefix, SlocationPrefix)
		location = append(location, Slocation)
		proxyPass = append(proxyPass, SproxyPass)
		consome = append(consome, Sconsome)
		owner = append(owner, Sowner)
		somenteDMZ = append(somenteDMZ, SDMZ)
		aplicado = append(aplicado, Saplicado)
	}

	return idM, locationPrefix, location, proxyPass, consome, owner, somenteDMZ, aplicado
}

// CreateMapDB - Cria mapeamento no banco de dados
func CreateMapDB(location, proxyPass, env string, idUser, idGroup int, consome, owner, locationPrefix, somenteDMZ string) (retorno bool) {
	var SDMZ int

	if somenteDMZ == "sim" {
		SDMZ = 1
	} else {
		SDMZ = 0
	}

	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	// insert
	stmt, err := db.Prepare("INSERT mapeamentos SET location_prefix=?, location=?, proxy_pass=?, env=?, id_usuario=?, id_grupo=?, consome=?, owner=?, somente_dmz=?, aprovado=0")
	checkErr(err)

	res, err := stmt.Exec(locationPrefix, location, proxyPass, env, idUser, idGroup, consome, owner, SDMZ)
	checkErr(err)

	id, err := res.LastInsertId()
	checkErr(err)

	// Verifica se registro foi inserido com sucesso
	if id > 0 {
		return true
	} else {
		return false
	}

	return retorno
}

// UpdateAplicadoMapeamento - Altera o status de aplicacao do mapeamento
func UpdateAplicadoMapeamento(idMapeamento, Aplicado int) (retorno bool) {
	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	// aplicado=0 - Ainda nao foi criado mapeamento
	// aplicado=1 - Mapeamento foi criado COM SUCESSO
	// aplicado=2 - Erro ao criar mapeamento, ele nao tentara ser criado novamente
	stmt, err := db.Prepare("UPDATE mapeamentos SET aplicado=? WHERE id_mapeamento=?")
	checkErr(err)

	res, err := stmt.Exec(Aplicado, idMapeamento)
	checkErr(err)

	affect, err := res.RowsAffected()
	checkErr(err)

	// Verifica se o registro foi alterado com sucesso
	if affect > 0 {
		return true
	} else {
		return false
	}

	return retorno
}

// ReturnServers - Retorna IP dos servidores
func ReturnServers(env, tipo string) (IP []string) {
	var SIP string

	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	rows, err := db.Query("SELECT IP FROM servidores WHERE env='" + env + "' AND type_server='" + tipo + "' AND ativo=1")
	checkErr(err)

	for rows.Next() {
		err = rows.Scan(&SIP)
		checkErr(err)

		IP = append(IP, SIP)
	}

	return IP
}

// DeleteMapDB - Deleta mapeamento no Banco de Dados
func DeleteMapDB(ID int, env string) (retorno bool) {
	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	// So deleta registro caso o mesmo esteja marcado para exclusao (aplicado=3) e aprovado (aprovado=1)
	stmt, err := db.Prepare("DELETE FROM mapeamentos WHERE id_mapeamento=? AND env=? AND aprovado=1 AND aplicado=3")
	checkErr(err)

	res, err := stmt.Exec(ID, env)
	checkErr(err)

	affect, err := res.RowsAffected()
	checkErr(err)

	if affect >= 1 {
		retorno = true
	} else {
		retorno = false
	}

	return retorno
}

// UpdateMD5Mapeamento - Atualiza o MD5 gerado de acordo com o arquivo
func UpdateMD5Mapeamento(idMapeamento int, MD5, tipo string) (retorno bool) {
	if tipo == "DMZ" {
		db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
		checkErr(err)

		defer db.Close()

		stmt, err := db.Prepare("UPDATE mapeamentos SET md5_dmz=? WHERE id_mapeamento=?")
		checkErr(err)

		res, err := stmt.Exec(MD5, idMapeamento)
		checkErr(err)

		affect, err := res.RowsAffected()
		checkErr(err)

		// Verifica se o registro foi alterado com sucesso
		if affect > 0 {
			return true
		} else {
			return false
		}
	} else {
		db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
		checkErr(err)

		defer db.Close()

		stmt, err := db.Prepare("UPDATE mapeamentos SET md5_app=? WHERE id_mapeamento=?")
		checkErr(err)

		res, err := stmt.Exec(MD5, idMapeamento)
		checkErr(err)

		affect, err := res.RowsAffected()
		checkErr(err)

		// Verifica se o registro foi alterado com sucesso
		if affect > 0 {
			return true
		} else {
			return false
		}

	}

	return retorno
}

// TemServerDMZAPP - Verifica se tem algum servidor cadastrado e habilitado de APP e DMZ
func TemServerDMZAPP(env string) (temServer bool) {
	db, err := sql.Open("mysql", UserDB+":"+PassDB+"@tcp("+HostDB+":"+PortDB+")/"+DatabaseDB+"?charset=utf8")
	checkErr(err)

	defer db.Close()

	rows, err := db.Query("SELECT count(IP) FROM servidores WHERE env='" + env + "' AND (type_server='APP' OR type_server='DMZ') AND ativo=1")
	checkErr(err)

	for rows.Next() {
		var count int
		err = rows.Scan(&count)
		checkErr(err)

		if count >= 1 {
			return true
		} else {
			return false
		}
	}

	return
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
