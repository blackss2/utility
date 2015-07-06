package convert

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func Int(val interface{}) int64 {
	if val != nil {
		switch val.(type) {
		case float64:
			return int64(val.(float64))
		case int64:
			return val.(int64)
		default:
			ret, err := strconv.ParseInt(String(val), 64)
			if err != nil {
				return 0
			} else {
				return ret
			}
		}
	}
	return 0
}

func IntWith(val interface{}, defaultValue int64) int64 {
	if val != nil {
		return Int(val)
	}
	return defaultValue
}

func Float(val interface{}) float64 {
	if val != nil {
		switch val.(type) {
		case float64:
			return val.(float64)
		case int64:
			return float64(val.(int64))
		default:
			ret, err := strconv.ParseFloat(String(val), 64)
			if err != nil {
				return 0
			} else {
				return float64(ret)
			}
		}
	}
	return 0
}

func FloatWith(val interface{}, defaultValue float64) float64 {
	if val != nil {
		return Float(val)
	}
	return defaultValue
}

func String(val interface{}) string {
	switch val.(type) {
	case string:
		return val.(string)
	case []byte:
		return string(val.([]byte))
	case time.Time:
		t := val.(time.Time)
		return fmt.Sprintf("%4.4d-%2.2d-%2.2d %2.2d:%2.2d:%2.2d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

func QueryString(val string) string {
	return strings.Join(strings.Split(val, "'"), "''")
}

func MD5(src string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(src)))
}
