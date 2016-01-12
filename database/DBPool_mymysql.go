package database

import (
	_ "github.com/ziutek/mymysql/godrv"
	"reflect"
	"strings"
)

func init() {
	cm := &ConfigMakerBase{
		keyHash:     make(map[string]string),
		typeHash:    make(map[string]reflect.Kind),
		defaultHash: make(map[string]interface{}),
		templateString: strings.Replace(strings.Replace(`
			tcp:{{ .Address }}:{{ .Port }}*{{ .Database }}/{{ .Id }}/{{ .Password }}
		`, "\n", "", -1), "\t", "", -1),
	}
	cm.AddKey(reflect.String, "Address", "Addr", "IP")
	cm.AddKey(reflect.Int, "Port")
	cm.AddKey(reflect.String, "Database", "db", "dbname")
	cm.AddKey(reflect.String, "Id", "UserId", "User")
	cm.AddKey(reflect.String, "Password", "pw", "passwd")
	cm.AddKey(reflect.Int, "Timeout")
	cm.SetDefault("Port", 3306)
	AddDriver("mymysql", cm)
}
