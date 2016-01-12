package database

import (
	"github.com/blackss2/utility/convert"
	"math"
	"reflect"
	"strings"
)

type ConfigMakerBase struct {
	keyHash        map[string]string
	requriedHash   map[string]bool
	templateString string
	typeHash       map[string]reflect.Kind
	defaultHash    map[string]interface{}
}

func (cm *ConfigMakerBase) Init(data map[string]interface{}) {
	for k, v := range cm.defaultHash {
		cm.Apply(data, k, v)
	}
}

func (cm *ConfigMakerBase) SetDefault(key string, value interface{}) {
	cm.defaultHash[key] = value
}

func (cm *ConfigMakerBase) AddKey(kind reflect.Kind, key string, morekeys ...string) {
	cm.typeHash[key] = kind
	cm.keyHash[strings.ToLower(key)] = key
	for _, n := range morekeys {
		cm.keyHash[strings.ToLower(n)] = key
	}
}

func (cm *ConfigMakerBase) TemplateString() string {
	return cm.templateString
}

func (cm *ConfigMakerBase) Apply(data map[string]interface{}, key string, value interface{}) bool {
	if k, has := cm.keyHash[strings.ToLower(key)]; has {
		kind := cm.typeHash[k]
		switch kind {
		case reflect.Int:
			if v := convert.IntWith(value, math.MaxInt64); v == math.MaxInt64 {
				return false
			} else {
				data[k] = v
			}
		case reflect.String:
			data[k] = convert.String(value)
		}
	}
	return true
}
