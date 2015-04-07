package convert

import (
	"crypto/md5"
	"fmt"
	"io"
	"strconv"
)

func Int(val interface{}) int64 {
	if val != nil {
		switch val.(type) {
		case float64:
			return int64(val.(float64))
		case int64:
			return val.(int64)
		case string:
			ret, err := strconv.Atoi(val.(string))
			if err != nil {
				return 0
			} else {
				return int64(ret)
			}
		}
	}
	return 0
}

func IntWith(val interface{}, defaultValue int64) int64 {
	if val != nil {
		switch val.(type) {
		case float64:
			return int64(val.(float64))
		case int64:
			return val.(int64)
		case string:
			ret, err := strconv.Atoi(val.(string))
			if err == nil {
				return int64(ret)
			}
		}
	}
	return defaultValue
}

func String(val interface{}) string {
	switch val.(type) {
	case string:
		return val.(string)
	case []byte:
		return string(val.([]byte))
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

func MD5(src string) string {
	h := md5.New()
	io.WriteString(h, src)
	return h.Sum(nil)
}
