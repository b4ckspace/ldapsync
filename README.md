# ldapsync

Since ldap replication sucks we can just mirror it to a password file.

[See Dovecot documentation](https://doc.dovecot.org/configuration_manual/authentication/passwd_file/)

## Usage

```sh
export BIND_DN="cn=reader,ou=example,dc=com"  # user to login to ldap
export PASSWORD="secret-password-337"         # password to login to ldap
export URL="ldap://localhost:389"             # ldap url
export SERVER_NAME="ldap.example.com"         # certificate hostname

export BASE_DN="ou=member,dc=example"         # dn where the users are
export FILTER="((serviceEnabled=mail))"       # filter
export OUTFILE="mail.passwd"                  # file to write
```
