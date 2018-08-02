# Resources
These are various resources and commands that helped build this app.

## LDAP

### Import stuff
This will import an ldif file into the LDAP server:

`ldapadd -x -D cn=admin,dc=hack,dc=int -W -f bits.ldif`

### Query stuff
`ldapsearch -x -LLL -b dc=hack,dc=int 'uid=leonardo' cn gidNumber`

### Modify the schema
`ldapmodify -Q -Y EXTERNAL -H ldapi:/// -f uid_index.ldif`

### Check the change - must be ran as root
`ldapsearch -Q -LLL -Y EXTERNAL -H ldapi:/// -b cn=config '(olcDatabase={1}mdb)' olcDbIndex`

### Show installed schemas
`ldapsearch -Q -LLL -Y EXTERNAL -H ldapi:/// -b cn=schema,cn=config dn`

### OpenLDAP server howto

<https://help.ubuntu.com/lts/serverguide/openldap-server.html>

### phpldap

<https://www.digitalocean.com/community/tutorials/how-to-install-and-configure-openldap-and-phpldapadmin-on-ubuntu-16-04>

### Sample ldap client

<https://github.com/jtblin/go-ldap-client>

### Go ldap server

<https://github.com/vjeantet/ldapserver>

## Dump the schema

Will create a schema with secret stuff in it so they have to dump the schema to find it then show it's fields

<http://www.openldap.org/faq/data/cache/1366.html>

## Golang

### Set where the local packages are installed
`export GOPATH=/usr/local/go/`

### The Go LDAP package
<https://github.com/go-ldap/ldap>

### Useful Go stuff

<https://gobyexample.com/>

### fmt formats

<https://golang.org/pkg/fmt/>

### LDAP library docs

<https://godoc.org/gopkg.in/ldap.v2>

### Go Code Layout

<https://golang.org/doc/code.html>

### Vim plugin

Needs an up-to-date version of Vim to work

<https://github.com/fatih/vim-go>

### Logging
<https://github.com/sirupsen/logrus>

## LDAP Exploitation

### Injection cheatsheets and docs

<http://www.blackhat.com/presentations/bh-europe-08/Alonso-Parada/Whitepaper/bh-eu-08-alonso-parada-WP.pdf>

[OWASP](https://www.owasp.org/index.php/Testing_for_LDAP_Injection_\(OTG-INPVAL-006\))
