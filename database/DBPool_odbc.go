package database

import (
	_ "github.com/alexbrainman/odbc"
	"reflect"
	"strings"
)

func init() {
	cm := &ConfigMakerBase{
		keyHash:  make(map[string]string),
		typeHash: make(map[string]reflect.Kind),
		templateString: strings.Replace(strings.Replace(`
			DNS={{ .Database }};UID={{ .Id }};PWD={{ .Password }}
		`, "\n", "", -1), "\t", "", -1),
	}
	cm.AddKey(reflect.String, "Database", "db", "dbname", "DNS")
	cm.AddKey(reflect.String, "Id", "UserId", "User")
	cm.AddKey(reflect.String, "Password", "pw", "passwd")
	AddDriver("odbc", cm)
}
