package database

import (
	"fmt"
)

func CreateDBPool(driver string, ip string, port int, name string, id string, pw string, poolSize int) *DBPool {
	var connString string
	switch driver {
	case "mssql":
		timeout := 300
		connString = fmt.Sprintf("Server=%s;Port=%d;Database=%s;User Id=%s;Password=%s;connection timeout=%d", ip, port, name, id, pw, timeout)
	case "mysql":
		connString = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", id, pw, ip, port, name)
	case "mymysql":
		connString = fmt.Sprintf("tcp:%s:%d*%s/%s/%s", ip, port, name, id, pw)
	case "odbc":
		connString = fmt.Sprintf("DSN=%s;UID=%s;PWD=%s", name, id, pw)
	case "postgres":
		timeout := 300
		connString = fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s connect_timeout=%d", ip, port, name, id, pw, timeout)
	//case "ql":
	//	connString = name
	default:
		panic("Unsupported driver : " + driver)
	}
	return CreateDBPoolByConnString(driver, connString, poolSize)
}
