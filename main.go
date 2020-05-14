package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"

	ldap "github.com/go-ldap/ldap/v3"
	"github.com/kelseyhightower/envconfig"
)

type (
	Config struct {
		Url        string `envconfig:"URL" required:"true"`
		BindDN     string `envconfig:"BIND_DN" required:"true"`
		Password   string `envconfig:"PASSWORD" required:"true"`
		ServerName string `envconfig:"SERVER_NAME" required:"true"`

		BaseDN  string `envconfig:"BASE_DN" required:"true"`
		Filter  string `envconfig:"FILTER" required:"true"`
		Outfile string `envconfig:"OUTFILE" required:"true"`
	}
)

func main() {
	c := Config{}
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatalf("unable to parse environment: %s", err)
	}

	conn, err := ldap.DialURL(c.Url)
	if err != nil {
		log.Fatalf("unable to connect to ldap: %s", err)
	}
	err = conn.StartTLS(&tls.Config{
		ServerName: c.ServerName,
	})
	if err != nil {
		log.Fatalf("unable to upgrade to tls: %s", err)
	}
	err = conn.Bind(c.BindDN, c.Password)
	if err != nil {
		log.Fatalf("unable to bind as %s: %s", c.BindDN, err)
	}

	res, err := conn.Search(&ldap.SearchRequest{
		BaseDN:     c.BaseDN,
		Filter:     c.Filter,
		Scope:      ldap.ScopeWholeSubtree,
		Attributes: []string{"email", "userPassword"},
	})
	if err != nil {
		log.Fatalf("unable to search: %s", err)
	}

	tmpFile := fmt.Sprintf("%s.tmp", c.Outfile)
	f, err := os.OpenFile(tmpFile, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		log.Fatalf("unable to open outputfile: %s", err)
	}
	defer f.Close()
	defer f.Sync()

	for _, entry := range res.Entries {
		umail := entry.GetAttributeValue("email")
		upass := entry.GetAttributeValue("userPassword")
		user := strings.SplitN(umail, "@", 2)[0]
		_, err = fmt.Fprintf(f, "%s:%s\n", user, upass)
		if err != nil {
			log.Fatalf("unable to write temporary outfile: %s", err)
		}
	}
	err = os.Rename(tmpFile, c.Outfile)
	if err != nil {
		log.Fatalf("unable to replace outfile: %s", err)
	}
}
