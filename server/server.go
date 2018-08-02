package LDAPServer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/digininja/vuLnDAP/config"
	"github.com/nmcclain/ldap"
	"net"
	"strings"
	//	"reflect" //fmt.Printf ("Type: %s", reflect.TypeOf(x))
)

import log "github.com/sirupsen/logrus"

// interface for backend handler
type Backend interface {
	ldap.Binder
	ldap.Searcher
	ldap.Closer
}

type lDAPServer struct {
	cfg config.Config
}

func NewLDAPServer(cfg config.Config) lDAPServer {
	l := lDAPServer{}
	l.cfg = cfg

	return l
}

func (h lDAPServer) Bind(bindDN, bindSimplePw string, conn net.Conn) (resultCode ldap.LDAPResultCode, err error) {
	bindDN = strings.ToLower(bindDN)
	baseDN := strings.ToLower("," + h.cfg.LDAPServer.BaseDN)
	serverLogger.Debug(fmt.Sprintf("Bind request as %s from %s", bindDN, conn.RemoteAddr().String()))

	// parse the bindDN
	if !strings.HasSuffix(bindDN, baseDN) {
		serverLogger.Warn(fmt.Sprintf("Bind Error: BindDN %s not our BaseDN %s", bindDN, baseDN))
		return ldap.LDAPResultInvalidCredentials, nil
	}
	parts := strings.Split(strings.TrimSuffix(bindDN, baseDN), ",")
	groupName := ""
	userName := ""
	if len(parts) == 1 {
		userName = strings.TrimPrefix(parts[0], "cn=")
	} else if len(parts) == 2 {
		userName = strings.TrimPrefix(parts[0], "cn=")
		groupName = strings.TrimPrefix(parts[1], "ou=")
	} else {
		serverLogger.Warn(fmt.Sprintf("Bind Error: BindDN %s should have only one or two parts (has %d)", bindDN, len(parts)))
		return ldap.LDAPResultInvalidCredentials, nil
	}
	// find the user
	user := config.NewConfigUser()
	found := false
	for _, u := range h.cfg.Users {
		if u.Name == userName {
			found = true
			user = u
		}
	}
	if !found {
		serverLogger.Warn(fmt.Sprintf("Bind Error: User %s not found.", userName))
		return ldap.LDAPResultInvalidCredentials, nil
	}
	// find the group
	group := config.NewConfigGroup()
	found = false
	for _, g := range h.cfg.Groups {
		if g.Name == groupName {
			found = true
			group = g
		}
	}
	if !found {
		serverLogger.Warn(fmt.Sprintf("Bind Error: Group %s not found.", groupName))
		return ldap.LDAPResultInvalidCredentials, nil
	}
	// validate group membership
	if user.PrimaryGroup != group.UnixID {
		serverLogger.Warn(fmt.Sprintf("Bind Error: User %s primary group is not %s.", userName, groupName))
		return ldap.LDAPResultInvalidCredentials, nil
	}

	// finally, validate user's pw
	hash := sha256.New()
	hash.Write([]byte(bindSimplePw))
	if user.PassSHA256 != hex.EncodeToString(hash.Sum(nil)) {
		serverLogger.Warn(fmt.Sprintf("Bind Error: invalid credentials as %s from %s", bindDN, conn.RemoteAddr().String()))
		return ldap.LDAPResultInvalidCredentials, nil
	}
	serverLogger.Debug(fmt.Sprintf("Bind success as %s from %s", bindDN, conn.RemoteAddr().String()))
	return ldap.LDAPResultSuccess, nil
}

func (h lDAPServer) Search(bindDN string, searchReq ldap.SearchRequest, conn net.Conn) (result ldap.ServerSearchResult, err error) {
	bindDN = strings.ToLower(bindDN)
	baseDN := strings.ToLower(h.cfg.LDAPServer.BaseDN)
	searchBaseDN := strings.ToLower(searchReq.BaseDN)
	serverLogger.Debug(fmt.Sprintf("Search request as %s from %s for %s", bindDN, conn.RemoteAddr().String(), searchReq.Filter))
	serverLogger.Debug(fmt.Sprintf("Search attributes %s", strings.Join(searchReq.Attributes, ",")))
	serverLogger.Debug(fmt.Sprintf("Search scope %s", ldap.ScopeMap[searchReq.Scope]))

	// validate the user is authenticated and has appropriate access
	if len(bindDN) < 1 {
		return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultInsufficientAccessRights}, fmt.Errorf("Search Error: Anonymous BindDN not allowed %s", bindDN)
	}
	if !strings.HasSuffix(bindDN, baseDN) {
		return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultInsufficientAccessRights}, fmt.Errorf("Search Error: BindDN %s not in our BaseDN %s", bindDN, baseDN)
	}

	if !strings.HasSuffix(searchBaseDN, baseDN) {

		//			return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultInsufficientAccessRights}, fmt.Errorf("Search Error: Search BaseDN %s is not in our BaseDN %s", searchBaseDN, baseDN)
	}
	searchBaseDN = "dc=hack,dc=me"

	// return all users in the config file - the LDAP library will filter results for us
	entries := []*ldap.Entry{}
	/* No longer needed
	filterEntity, err := ldap.GetFilterObjectClass(searchReq.Filter)
	if err != nil {
		return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, fmt.Errorf("Search Error: error parsing filter: %s", searchReq.Filter)
	}
	*/

	// Load all the content into the entries
	serverLogger.Debug("Populating the entities list")

	/*
		Would be nice to implement this
		http://ldapwiki.com/wiki/RootDSE

		It is used by phpldapadmin and other things to help work out the structure to the database.

		This command gets the name of the sub schema entry

			ldapsearch -x -b '' -s base subschemaSubentry

		That gives this result
			subschemaSubentry: cn=Subschema

		Then take the cn=... and put it into here

			ldapsearch -x -b 'cn=Subschema' -s base '(objectClass=subschema)' attributetypes

		All this comes from here
			http://phpldapadmin.sourceforge.net/wiki/index.php/FAQ

		Would also need to support anonymous connections which I currently don't

		serverLogger.Debug(fmt.Sprintf("Loading: Structure"))
		attrs := []*ldap.EntryAttribute{}
		attrs = append(attrs, &ldap.EntryAttribute{"namingcontexts", []string{"dc=hack,dc=me"}})
		attrs = append(attrs, &ldap.EntryAttribute{"objectClass", []string{"top"}})
		attrs = append(attrs, &ldap.EntryAttribute{"subschemaSubentry", []string{"cn=Subschema"}})

		attrs = append(attrs, &ldap.EntryAttribute{"altServer", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"supportedExtension", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"supportedControl", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"supportedSASLMechanisms", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"supportedLDAPVersion", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"currentTime", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"dsServiceName", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"defaultNamingContext", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"schemaNamingContext", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"configurationNamingContext", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"rootDomainNamingContext", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"supportedLDAPPolicies", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"highestCommittedUSN", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"dnsHostName", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"ldapServiceName", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"serverName", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"supportedCapabilities", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"changeLog", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"tlsAvailableCipherSuites", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"tlsImplementationVersion", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"supportedSASLMechanisms", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"dsaVersion", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"myAccessPoint", []string{""}})
		attrs = append(attrs, &ldap.EntryAttribute{"dseType", []string{""}})

		dn := fmt.Sprintf("dc=hack,dc=me")
		entries = append(entries, &ldap.Entry{dn, attrs})

	*/
	serverLogger.Debug(("Vegetables"))
	for _, g := range h.cfg.Vegetables {
		serverLogger.Debug(fmt.Sprintf("Loading: %s", g.Name))
		attrs := []*ldap.EntryAttribute{}
		attrs = append(attrs, &ldap.EntryAttribute{"cn", []string{g.Name}})
		attrs = append(attrs, &ldap.EntryAttribute{"stock", []string{fmt.Sprintf("%d", g.Stock)}})
		attrs = append(attrs, &ldap.EntryAttribute{"description", []string{g.Description}})
		attrs = append(attrs, &ldap.EntryAttribute{"objectClass", []string{"vegetables"}})
		dn := fmt.Sprintf("cn=%s,ou=fruits,%s", g.Name, baseDN)
		entries = append(entries, &ldap.Entry{dn, attrs})
	}
	serverLogger.Debug(("Fruit"))
	for _, g := range h.cfg.Fruits {
		serverLogger.Debug(fmt.Sprintf("Loading: %s", g.Name))
		attrs := []*ldap.EntryAttribute{}
		attrs = append(attrs, &ldap.EntryAttribute{"cn", []string{g.Name}})
		attrs = append(attrs, &ldap.EntryAttribute{"stock", []string{fmt.Sprintf("%d", g.Stock)}})
		attrs = append(attrs, &ldap.EntryAttribute{"description", []string{g.Description}})
		attrs = append(attrs, &ldap.EntryAttribute{"objectClass", []string{"fruits"}})
		dn := fmt.Sprintf("cn=%s,ou=fruits,%s", g.Name, baseDN)
		entries = append(entries, &ldap.Entry{dn, attrs})
	}
	serverLogger.Debug(("Groups"))
	for _, g := range h.cfg.Groups {
		serverLogger.Debug(fmt.Sprintf("Loading: %s", g.Name))
		attrs := []*ldap.EntryAttribute{}
		attrs = append(attrs, &ldap.EntryAttribute{"cn", []string{g.Name}})
		attrs = append(attrs, &ldap.EntryAttribute{"description", []string{fmt.Sprintf("%s via LDAP", g.Name)}})
		attrs = append(attrs, &ldap.EntryAttribute{"gidNumber", []string{fmt.Sprintf("%d", g.UnixID)}})
		attrs = append(attrs, &ldap.EntryAttribute{"objectClass", []string{"posixGroup"}})
		attrs = append(attrs, &ldap.EntryAttribute{"uniqueMember", h.getGroupMembers(g.UnixID)})
		attrs = append(attrs, &ldap.EntryAttribute{"memberUid", h.getGroupMemberIDs(g.UnixID)})
		dn := fmt.Sprintf("cn=%s,ou=groups,%s", g.Name, baseDN)
		entries = append(entries, &ldap.Entry{dn, attrs})
	}
	// Defined here - http://ldapwiki.com/wiki/PosixAccount
	serverLogger.Debug(("Users"))
	for _, u := range h.cfg.Users {
		serverLogger.Debug(fmt.Sprintf("Loading: %s", u.Name))
		attrs := []*ldap.EntryAttribute{}
		attrs = append(attrs, &ldap.EntryAttribute{"cn", []string{u.Name}})
		attrs = append(attrs, &ldap.EntryAttribute{"uid", []string{u.Name}})
		attrs = append(attrs, &ldap.EntryAttribute{"ou", []string{h.getGroupName(u.PrimaryGroup)}})
		attrs = append(attrs, &ldap.EntryAttribute{"uidNumber", []string{fmt.Sprintf("%d", u.UnixID)}})
		attrs = append(attrs, &ldap.EntryAttribute{"accountStatus", []string{"active"}})
		attrs = append(attrs, &ldap.EntryAttribute{"objectClass", []string{"posixAccount"}})
		attrs = append(attrs, &ldap.EntryAttribute{"homeDirectory", []string{"/home/" + u.Name}})
		attrs = append(attrs, &ldap.EntryAttribute{"loginShell", []string{"/bin/bash"}})
		attrs = append(attrs, &ldap.EntryAttribute{"description", []string{fmt.Sprintf("%s", u.Description)}})
		attrs = append(attrs, &ldap.EntryAttribute{"gecos", []string{u.Gecos}})
		attrs = append(attrs, &ldap.EntryAttribute{"gidNumber", []string{fmt.Sprintf("%d", u.PrimaryGroup)}})
		attrs = append(attrs, &ldap.EntryAttribute{"memberOf", h.getGroupDNs(u.OtherGroups)})
		if len(u.SSHKeys) > 0 {
			attrs = append(attrs, &ldap.EntryAttribute{"sshPublicKey", u.SSHKeys})
		}
		dn := fmt.Sprintf("cn=%s,ou=%s,%s", u.Name, h.getGroupName(u.PrimaryGroup), baseDN)
		entries = append(entries, &ldap.Entry{dn, attrs})
	}
	serverLogger.Debug("Content loaded, about to search")
	res := ldap.ServerSearchResult{entries, []string{}, []ldap.Control{}, ldap.LDAPResultSuccess}
	serverLogger.Info("Search finished")
	return res, nil
}

//
func (h lDAPServer) Close(boundDn string, conn net.Conn) error {
	return nil
}

//
func (h lDAPServer) getGroupMembers(gid int) []string {
	members := make(map[string]bool)
	for _, u := range h.cfg.Users {
		if u.PrimaryGroup == gid {
			dn := fmt.Sprintf("cn=%s,ou=%s,%s", u.Name, h.getGroupName(u.PrimaryGroup), h.cfg.LDAPServer.BaseDN)
			members[dn] = true
		} else {
			for _, othergid := range u.OtherGroups {
				if othergid == gid {
					dn := fmt.Sprintf("cn=%s,ou=%s,%s", u.Name, h.getGroupName(u.PrimaryGroup), h.cfg.LDAPServer.BaseDN)
					members[dn] = true
				}
			}
		}
	}
	m := []string{}
	for k, _ := range members {
		m = append(m, k)
	}
	return m
}

//
func (h lDAPServer) getGroupMemberIDs(gid int) []string {
	members := make(map[string]bool)
	for _, u := range h.cfg.Users {
		if u.PrimaryGroup == gid {
			members[u.Name] = true
		} else {
			for _, othergid := range u.OtherGroups {
				if othergid == gid {
					members[u.Name] = true
				}
			}
		}
	}
	m := []string{}
	for k, _ := range members {
		m = append(m, k)
	}
	return m
}

//
func (h lDAPServer) getGroupDNs(gids []int) []string {
	groups := make(map[string]bool)
	for _, gid := range gids {
		for _, g := range h.cfg.Groups {
			if g.UnixID == gid {
				dn := fmt.Sprintf("cn=%s,ou=groups,%s", g.Name, h.cfg.LDAPServer.BaseDN)
				groups[dn] = true
			}
		}
	}
	g := []string{}
	for k, _ := range groups {
		g = append(g, k)
	}
	return g
}

//
func (h lDAPServer) getGroupName(gid int) string {
	for _, g := range h.cfg.Groups {
		if g.UnixID == gid {
			return g.Name
		}
	}
	return ""
}

var serverLogger = log.WithFields(log.Fields{"Owner": "Server"})

func StartLDAPServer(cfg config.Config) {
	serverLogger.Info("Server started")

	// configure the backend
	s := ldap.NewServer()
	s.EnforceLDAP = true
	handler := NewLDAPServer(cfg)
	s.BindFunc("", handler)
	s.SearchFunc("", handler)
	s.CloseFunc("", handler)

	listen := fmt.Sprintf("%s:%d", cfg.LDAPServer.ListenIP, cfg.LDAPServer.ListenPort)
	serverLogger.Info(fmt.Sprintf("LDAP server listening on %s", listen))
	if err := s.ListenAndServe(listen); err != nil {
		serverLogger.Fatalf(fmt.Sprintf("LDAP Server Failed: %s", err.Error()))
	}

	serverLogger.Info("AP exit")
}
