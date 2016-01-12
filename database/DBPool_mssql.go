package database

import (
	_ "github.com/denisenkom/go-mssqldb"
	"reflect"
	"strings"
)

func init() {
	cm := &ConfigMakerBase{
		keyHash:     make(map[string]string),
		typeHash:    make(map[string]reflect.Kind),
		defaultHash: make(map[string]interface{}),
		templateString: strings.Replace(strings.Replace(`
			{{ with .Address }}Server={{ . }}{{ end }}
			{{ with .Port }};Port={{ . }}{{ end }}
			{{ with .Database }};Database={{ . }}{{ end }}
			{{ with .Id }};User Id={{ . }}{{ end }}
			{{ with .Password }};Password={{ . }}{{ end }}
			{{ with .Timeout }};connection timeout={{ . }}{{ end }}
		`, "\n", "", -1), "\t", "", -1),
	}
	cm.AddKey(reflect.String, "Address", "Addr", "IP")
	cm.AddKey(reflect.Int, "Port")
	cm.AddKey(reflect.String, "Database", "db", "dbname")
	cm.AddKey(reflect.String, "Id", "UserId", "User")
	cm.AddKey(reflect.String, "Password", "pw", "passwd")
	cm.AddKey(reflect.Int, "Timeout")
	cm.SetDefault("Port", 1433)
	AddDriver("mssql", cm)
}
