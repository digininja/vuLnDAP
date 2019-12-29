package main

import (
	"fmt"
	"gopkg.in/ldap.v2"
	"net/http"
)
import "github.com/digininja/vuLnDAP/config"

type Vulndap struct {
	bindusername string
	bindpassword string
	host         string
	port         int
	baseDN       string
	ldap         *ldap.Conn
	verbose      bool
}

func NewVulndap() Vulndap {
	vulndap := Vulndap{}
	return vulndap
}

func (v *Vulndap) close() {
	v.ldap.Close()

	if v.verbose {
		clientLogger.Info("LDAP connection closed")
	}
}

func (v *Vulndap) connect(config config.Config) {
	v.bindusername = config.LDAPClient.BindUser
	v.bindpassword = config.LDAPClient.BindPassword
	v.host = config.LDAPClient.BindHost
	v.port = config.LDAPClient.BindPort
	v.baseDN = config.LDAPClient.BaseDN

	host_port := fmt.Sprintf("%s:%d", v.host, v.port)

	clientLogger.Info(fmt.Sprintf("Binding to %s", host_port))

	l, err := ldap.Dial("tcp", host_port)

	// Enable debugging, dumps load of useful info
	// l.Debug = true
	if err != nil {
		clientLogger.Error("Failed to connect to LDAP server")
		clientLogger.Fatal(err)
	}

	// Bind with user
	err = l.Bind(v.bindusername, v.bindpassword)
	if err != nil {
		clientLogger.Error("Failed to bind using login credentials")
		clientLogger.Fatal(err)
	}

	v.ldap = l

	if v.verbose {
		clientLogger.Fatal("Connected to LDAP server")
	}
}

func (v *Vulndap) search(w http.ResponseWriter, filter string, attributes []string, numberOfResults int) (ldap.SearchResult, error) {

	/*
		   From

		   https://godoc.org/gopkg.in/ldap.v2#NewSearchRequest

		   SizeLimit is the number that should be returned, if more come back
		   the search fails with an error, it doesn't truncate

		   func NewSearchRequest(
					BaseDN string,
					Scope int,
					DerefAliases int,
					SizeLimit int,
					TimeLimit int,
					TypesOnly bool,
					Filter string,
					Attributes []string,
					Controls []Control,
		   ) *SearchRequest
	*/

	clientLogger.Info(fmt.Sprintf("Search filter: %s", filter))

	// Do the search
	searchRequest := ldap.NewSearchRequest(
		v.baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		numberOfResults,
		0,
		false,
		filter,
		attributes,
		nil,
	)

	sr, err := v.ldap.Search(searchRequest)

	return *sr, err
}
