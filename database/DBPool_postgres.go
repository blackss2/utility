package database

import (
	_ "github.com/lib/pq"
	"reflect"
	"strings"
)

func init() {
	cm := &ConfigMakerBase{
		keyHash:     make(map[string]string),
		typeHash:    make(map[string]reflect.Kind),
		defaultHash: make(map[string]interface{}),
		templateString: strings.Replace(strings.Replace(`
			{{ with .Address }}host={{ . }}{{ end }}
			{{ with .Port }} port={{ . }}{{ end }}
			{{ with .Database }} dbname={{ . }}{{ end }}
			{{ with .Id }} user={{ . }}{{ end }}
			{{ with .Password }} password={{ . }}{{ end }}
			{{ with .Timeout }} connect_timeout={{ . }}{{ end }}
			{{ with .SSL }} sslmode={{ . }}{{ end }}
		`, "\n", "", -1), "\t", "", -1),
	}
	cm.AddKey(reflect.String, "Address", "Addr", "IP")
	cm.AddKey(reflect.Int, "Port")
	cm.AddKey(reflect.String, "Database", "db", "dbname")
	cm.AddKey(reflect.String, "Id", "UserId", "User")
	cm.AddKey(reflect.String, "Password", "pw", "passwd")
	cm.AddKey(reflect.Int, "Timeout")
	cm.AddKey(reflect.String, "SSL", "sslmode")
	cm.SetDefault("Port", 5432)
	AddDriver("postgres", cm)
}
