package database

import (
	"fmt"
)

type DBPool struct {
	driver      string
	connString  string
	poolSize    int
	dbQueue     chan *Database
	PostConnect []string
	IsForceUTF8 bool
}

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

func CreateDBPoolByConnString(driver string, connString string, poolSize int) *DBPool {
	pool := &DBPool{driver, connString, poolSize, make(chan *Database, poolSize), make([]string, 0), false}
	err := pool.fill()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return pool
}

func (p *DBPool) fill() error {
	if p.driver == "ql" {
		db := new(Database)
		err := db.Open(p.driver, p.connString)
		if err != nil {
			return err
		}
		for len(p.dbQueue) < p.poolSize {
			p.dbQueue <- db
		}
	} else {
		for len(p.dbQueue) < p.poolSize {
			db := new(Database)
			err := db.Open(p.driver, p.connString)
			if err != nil {
				return err
			}
			p.dbQueue <- db
		}
	}
	return nil
}

func (p *DBPool) AddPostConnect(v string) {
	p.PostConnect = append(p.PostConnect, v)
}

func (p *DBPool) GetDB() *Database {
	db := <-p.dbQueue
	db.postConnect = p.PostConnect
	db.isForceUTF8 = p.IsForceUTF8
	return db
}

func (p *DBPool) ReleaseDB(db *Database) {
	p.dbQueue <- db
}
