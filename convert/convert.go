package convert

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/suapapa/go_hangul/encoding/cp949"
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
			ret, err := strconv.ParseInt(String(val), 10, 64)
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
		switch val.(type) {
		case float64:
			return int64(val.(float64))
		case int64:
			return val.(int64)
		default:
			ret, err := strconv.ParseInt(String(val), 10, 64)
			if err != nil {
				return defaultValue
			} else {
				return ret
			}
		}
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
		switch val.(type) {
		case float64:
			return val.(float64)
		case int64:
			return float64(val.(int64))
		default:
			ret, err := strconv.ParseFloat(String(val), 64)
			if err != nil {
				return defaultValue
			} else {
				return float64(ret)
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
	case time.Time:
		t := val.(time.Time)
		return fmt.Sprintf("%4.4d-%2.2d-%2.2d %2.2d:%2.2d:%2.2d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	case nil:
		return ""
	default:
		if val == nil {
			return ""
		} else {
			return fmt.Sprintf("%v", val)
		}
	}
}

func QueryString(val string) string {
	return strings.Join(strings.Split(val, "'"), "''")
}

func MD5(src string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(src)))
}

func SHA256(src string) string {
	hasher := sha256.New()
	hasher.Write([]byte(src))
	return hex.EncodeToString(hasher.Sum(nil))
}

func Time(val interface{}) *time.Time {
	if val != nil {
		v := String(val)
		if len(v) > 0 {
			if t, err := time.Parse("2006-01-02 15:04:05.000", v); err == nil && t.Year() > 0 {
				return &t
			} else if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil && t.Year() > 0 {
				return &t
			} else if t, err := time.Parse("2006-01-02", v); err == nil && t.Year() > 0 {
				return &t
			} else if t, err := time.Parse("20060102150405", v); err == nil && t.Year() > 0 {
				return &t
			} else if t, err := time.Parse("20060102", v); err == nil && t.Year() > 0 {
				return &t
			} else if t, err := time.Parse("01-02-06", v); err == nil && t.Year() > 0 {
				return &t
			} else if t, err := time.Parse("01-02-2006", v); err == nil && t.Year() > 0 {
				return &t
			} else if t, err := time.Parse("15:04:05.000", v); err == nil {
				return &t
			} else if t, err := time.Parse("15:04:05", v); err == nil {
				return &t
			}
		}
	}
	return nil
}

func UTF8(ustr string) (str string) {
	defer func() {
		if r := recover(); r != nil {
			ustr = str
			return
		}
	}()

	bytes, err := cp949.From([]byte(ustr))
	if err != nil {
		str = ustr
	} else {
		str = string(bytes)
	}
	return
}

func EUCKR(str string) (ustr string) {
	defer func() {
		if r := recover(); r != nil {
			ustr = str
			return
		}
	}()

	ubytes, err := cp949.To([]byte(str))
	if err != nil {
		ustr = str
	} else {
		ustr = string(ubytes)
	}
	return
}
