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

		BaseDN    string `envconfig:"BASE_DN" required:"true"`
		Filter    string `envconfig:"FILTER" required:"true"`
		PassFile  string `envconfig:"PASS_FILE" required:"true"`
		AliasFile string `envconfig:"ALIAS_FILE" required:"true"`
	}
)

func main() {
	// parse config
	c := Config{}
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatalf("unable to parse environment: %s", err)
	}

	// connect, secure and login to ldap
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

	// search ldap
	res, err := conn.Search(&ldap.SearchRequest{
		BaseDN:     c.BaseDN,
		Filter:     c.Filter,
		Scope:      ldap.ScopeWholeSubtree,
		Attributes: []string{"email", "userPassword", "emailAlias"},
	})
	if err != nil {
		log.Fatalf("unable to search: %s", err)
	}

	// create temporary passfile
	tmpPassfile := fmt.Sprintf("%s.tmp", c.PassFile)
	f, err := os.OpenFile(tmpPassfile, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		log.Fatalf("unable to open outputfile: %s", err)
	}
	for _, entry := range res.Entries {
		umail := entry.GetAttributeValue("email")
		upass := entry.GetAttributeValue("userPassword")
		user := strings.SplitN(umail, "@", 2)[0]
		_, err = fmt.Fprintf(f, "%s:%s\n", user, upass)
		if err != nil {
			log.Fatalf("unable to write temporary outfile: %s", err)
		}
	}
	_ = f.Close()

	// create temporary aliasfile
	tmpAliasfile := fmt.Sprintf("%s.tmp", c.AliasFile)
	f, err = os.OpenFile(tmpAliasfile, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		log.Fatalf("unable to open aliasfile: %s", err)
	}
	aliases := map[string][]string{}
	for _, entry := range res.Entries {
		umail := entry.GetAttributeValue("email")
		ualiases := entry.GetAttributeValues("emailAlias")
		for _, ualias := range ualiases {
			_, ok := aliases[ualias]
			if !ok {
				aliases[ualias] = []string{}
			}
			aliases[ualias] = append(aliases[ualias], umail)
		}
	}
	for alias, targets := range aliases {
		_, err := fmt.Fprintf(f, "%s: %s", alias, targets[0])
		if err != nil {
			log.Fatalf("unable to write temporary aliasfile: %s", err)
		}
		for _, target := range targets[1:] {
			_, err := fmt.Fprintf(f, ",%s", target)
			if err != nil {
				log.Fatalf("unable to write temporary aliasfile: %s", err)
			}
		}
		_, err = fmt.Fprintf(f, "\n")
		if err != nil {
			log.Fatalf("unable to write temporary aliasfile: %s", err)
		}
	}
	_ = f.Close()

	// update files
	err = os.Rename(tmpAliasfile, c.AliasFile)
	if err != nil {
		log.Fatalf("unable to replace aliasfile: %s", err)
	}
	err = os.Rename(tmpPassfile, c.PassFile)
	if err != nil {
		log.Fatalf("unable to replace outfile: %s", err)
	}
}
