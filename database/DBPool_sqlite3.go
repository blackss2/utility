package database

import (
	_ "github.com/mattn/go-sqlite3"
	"reflect"
	"strings"
)

func init() {
	cm := &ConfigMakerBase{
		keyHash:     make(map[string]string),
		typeHash:    make(map[string]reflect.Kind),
		defaultHash: make(map[string]interface{}),
		templateString: strings.Replace(strings.Replace(`
			{{ .Path }}
		`, "\n", "", -1), "\t", "", -1),
	}
	cm.AddKey(reflect.String, "Path")
	AddDriver("sqlite3", cm)
}
