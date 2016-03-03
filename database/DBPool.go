package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"text/template"
)

type Applyer interface {
	Apply(data map[string]interface{}, key string, value interface{}) bool
}

type ConfigMaker interface {
	Init(data map[string]interface{})
	TemplateString() string
	Applyer
}

type Config struct {
	driver  string
	applyer Applyer
	tp      *template.Template
	data    map[string]interface{}
}

func newConfig(driver string, ap Applyer, templateString string) *Config {
	t, err := template.New("_").Parse(templateString)
	if err != nil {
		panic(err)
	}
	c := &Config{
		driver:  driver,
		applyer: ap,
		tp:      t,
		data:    make(map[string]interface{}),
	}
	return c
}

func (c *Config) Set(key string, value interface{}) *Config {
	if !c.applyer.Apply(c.data, key, value) {
		log.Fatalln("unacceptable data", key, value)
	}
	return c
}

func (c *Config) Json(data string) *Config {
	hash := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &hash)
	if err != nil {
		log.Fatalln(err, data)
	} else {
		for k, v := range hash {
			c.applyer.Apply(c.data, k, v)
		}
	}
	return c
}

func (c *Config) ConnectionString() string {
	var buffer bytes.Buffer
	c.tp.Execute(&buffer, c.data)
	return buffer.String()
}

func (c *Config) Open(poolsize int) *DBPool {
	conn := c.ConnectionString()
	return CreateDBPoolByConnString(c.driver, conn, poolsize)
}

func (c *Config) OpenDirect() *Database {
	db := new(Database)
	err := db.Open(c.driver, c.ConnectionString())
	if err != nil {
		panic(err)
	}
	return db
}

var gDriverHash map[string]ConfigMaker = make(map[string]ConfigMaker)

func AddDriver(driver string, cm ConfigMaker) {
	gDriverHash[strings.ToLower(driver)] = cm
}

func SupportDrivers() []string {
	list := make([]string, len(gDriverHash))
	for k, _ := range gDriverHash {
		list = append(list, k)
	}
	sort.Strings(list)
	return list
}

type DBPool struct {
	driver      string
	connString  string
	poolSize    int
	dbQueue     chan *Database
	PostConnect []string
	IsForceUTF8 bool
}

func NewPool(driver string) *Config {
	cm, has := gDriverHash[strings.ToLower(driver)]
	if !has {
		log.Fatalln("not supported driver", driver)
		return nil
	}
	c := newConfig(driver, cm, cm.TemplateString())
	cm.Init(c.data)
	return c
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
