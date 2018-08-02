package main

import "fmt"
import "os"
import "time"
import "flag"
import log "github.com/sirupsen/logrus"
import "github.com/digininja/vuLnDAP/server"
import "github.com/digininja/vuLnDAP/config"

var VulndapConnection = NewVulndap()
var clientLogger = log.WithFields(log.Fields{"Owner": "Client"})
var mainLogger = log.WithFields(log.Fields{"Owner": "Main"})

func main() {
	mainLogger.Info("Main App Started")
	clientLogger.Info("Client Started")

	configFilePtr := flag.String("configl", "vulndap.cfg", "Alternative configuration file")
	flag.Parse()

	var cfg, err = config.NewConfig(*configFilePtr)
	if err != nil {
		mainLogger.Fatal(fmt.Sprintf("Configuration file error: %s", err.Error()))
	}

	if cfg.Debug {
		cfg.Dump()
	}

	mainLogger.Info("Starting LDAP server")
	go LDAPServer.StartLDAPServer(cfg)

	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	mainLogger.Debug("Sleeping for 500ms to give the server time to start up")
	time.Sleep(500 * time.Millisecond)

	mainLogger.Info("Binding to LDAP server")
	VulndapConnection.connect(cfg)
	defer VulndapConnection.close()

	webserver := NewWebServer(cfg)
	webserver.startWebApp()

	os.Exit(1)
}
