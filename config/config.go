package config

import "fmt"
import "github.com/BurntSushi/toml"
import log "github.com/sirupsen/logrus"

type configWebServer struct {
	ListenIP   string
	ListenPort int
}
type configLDAPClient struct {
	BaseDN       string
	BindUser     string
	BindPassword string
	BindHost     string
	BindPort     int
}
type configLDAPServer struct {
	BaseDN     string
	ListenIP   string
	ListenPort int
}
type configUser struct {
	Name         string
	Fruit        string
	Vegetable    string
	OtherGroups  []int
	PassSHA256   string
	PrimaryGroup int
	SSHKeys      []string
	UnixID       int
	Description  string
	Gecos        string //https://en.wikipedia.org/wiki/Gecos_field
}
type configVegetable struct {
	Name        string
	Description string
	Stock       int
}
type configFruit struct {
	Name        string
	Description string
	Stock       int
}
type configGroup struct {
	Name   string
	UnixID int
}
type Config struct {
	Debug      bool
	Groups     []configGroup
	Users      []configUser
	Fruits     []configFruit
	Vegetables []configVegetable
	LDAPClient configLDAPClient
	LDAPServer configLDAPServer
	WebServer  configWebServer
}

func NewConfigGroup() configGroup {
	cfgGroup := configGroup{}
	return cfgGroup
}

func NewConfigUser() configUser {
	cfgUser := configUser{}
	return cfgUser
}

func NewConfig(configFile string) (cfg Config, err error) {
	err = cfg.parseFile(configFile)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (cfg *Config) parseFile(configFile string) error {
	if _, err := toml.DecodeFile(configFile, &cfg); err != nil {
		return err
	}
	return nil
}

func (cfg Config) Dump() {
	var configLogger = log.WithFields(log.Fields{"Owner": "Config"})
	log.SetLevel(log.DebugLevel)

	configLogger.Debug("Dumping configuration information")
	configLogger.Debug(fmt.Sprintf("Web server listening on: %s:%d", cfg.WebServer.ListenIP, cfg.WebServer.ListenPort))
	configLogger.Debug(fmt.Sprintf("Connecting to LDAP server on: %s:%d", cfg.LDAPClient.BindHost, cfg.LDAPClient.BindPort))
	configLogger.Debug(fmt.Sprintf("BaseDN: %s", cfg.LDAPClient.BaseDN))
	configLogger.Debug(fmt.Sprintf("Using credentials: %s / %s", cfg.LDAPClient.BindUser, cfg.LDAPClient.BindPassword))
	configLogger.Debug(fmt.Sprintf("Debug mode: %t", cfg.Debug))
}
