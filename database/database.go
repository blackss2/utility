package database

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/alexbrainman/odbc"
	"github.com/blackss2/utility/convert"
	"github.com/cznic/ql"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/ziutek/mymysql/godrv"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Database struct {
	inst        *sql.DB
	instQL      *ql.DB
	connString  string
	driver      string
	postConnect []string
	isForceUTF8 bool
}

func (db *Database) Open(driver string, connString string) error {
	db.driver = driver
	db.connString = connString
	runtime.SetFinalizer(db, func(f interface{}) {
		f.(*Database).Close()
	})
	return db.executeOpen()
}

func (db *Database) executeOpen() error {
	var err error
	if db.driver == "ql" {
		if db.connString == "mem" {
			db.instQL, err = ql.OpenMem()
		} else {
			opt := &ql.Options{}
			opt.CanCreate = true

			filepath.Walk("./", func(path string, fi os.FileInfo, err error) error {
				if !fi.IsDir() && filepath.Dir(path) == "." {
					if len(fi.Name()) == 41 && fi.Name()[0] == '.' {
						os.Remove(path)
					}
				}
				return nil
			})

			db.instQL, err = ql.OpenFile(db.connString, opt)
		}
	} else {
		db.inst, err = sql.Open(db.driver, db.connString)
	}
	if err == nil && len(db.postConnect) > 0 {
		for _, v := range db.postConnect {
			db.TempQuery(v)
		}
	}
	return err
}

func (db *Database) Close() error {
	var err error

	if db.inst != nil {
		err = db.inst.Close()
	} else if db.instQL != nil {
		err = db.instQL.Close()
	}
	return err
}

type Rows struct {
	inst        *sql.Rows
	qlRows      [][]interface{}
	qlIndex     int
	isFirst     bool
	isNil       bool
	Cols        []string
	isForceUTF8 bool
}

func (db *Database) Query(queryStr string) (*Rows, error) {
	rows := &Rows{nil, nil, 0, true, false, make([]string, 0, 100), db.isForceUTF8}

	QUERYSTR := strings.ToUpper(queryStr)

	if db.inst != nil {
		stmt, err := db.inst.Prepare(queryStr)
		if stmt != nil {
			defer stmt.Close()
		}
		if err != nil {
			db.Close()
			db.executeOpen()
			return db.TempQuery(queryStr)
		}
		rows.inst, err = stmt.Query()

		if err != nil {
			if err.Error() != "Stmt did not create a result set" {
				db.Close()
				db.executeOpen()
				return db.TempQuery(queryStr)
			} else {
				runtime.SetFinalizer(rows, func(f interface{}) {
					f.(*Rows).Close()
				})
				return rows, nil
			}
		}

		rows.Cols, err = rows.inst.Columns()

		if !rows.inst.Next() {
			rows.Close()
		} else {
			runtime.SetFinalizer(rows, func(f interface{}) {
				f.(*Rows).Close()
			})
		}
	} else if db.instQL != nil {
		if !strings.Contains(QUERYSTR, "TRANSACTION") && (strings.Contains(QUERYSTR, "INSERT") || strings.Contains(QUERYSTR, "CREATE") || strings.Contains(QUERYSTR, "UPDATE") || strings.Contains(QUERYSTR, "DELETE")) {
			queryStr = fmt.Sprintf(`
				BEGIN TRANSACTION;
					%s;
				COMMIT;
			`, queryStr)
		}

		ctx := ql.NewRWCtx()
		rs, _, err := db.instQL.Run(ctx, queryStr, nil)
		if err != nil {
			//println("P1 : ", err.Error(), "\n", queryStr)
			return nil, err
		}

		if len(rs) == 0 {
			rows.isNil = true
			rows.isFirst = false
			return rows, nil
		}

		rows.Cols, err = rs[0].Fields()

		rows.qlRows, err = rs[0].Rows(-1, -1)
		if len(rows.qlRows) == 0 || err != nil {
			rows.Close()
		} else {
			runtime.SetFinalizer(rows, func(f interface{}) {
				f.(*Rows).Close()
			})
		}
	} else {
		return nil, errors.New("db is not initialized")
	}

	if strings.HasPrefix(QUERYSTR, "INSERT") && strings.Contains(QUERYSTR, "OUTPUT") && strings.Contains(QUERYSTR, "INSERTED.") {
		if rows.IsNil() {
			return nil, errors.New("insert.fail")
		}
	}
	return rows, nil
}

func (db *Database) TempQuery(queryStr string) (*Rows, error) {
	rows := &Rows{nil, nil, 0, true, false, make([]string, 0, 100), db.isForceUTF8}

	if db.inst != nil {
		stmt, err := db.inst.Prepare(queryStr)
		if stmt != nil {
			defer stmt.Close()
		}
		if err != nil {
			//println("P1 : ", err.Error())
			return nil, err
		}
		rows.inst, err = stmt.Query()

		if err != nil {
			if err.Error() != "Stmt did not create a result set" {
				//println("P2 : ", err.Error(), "\n", queryStr)
				return nil, err
			} else {
				runtime.SetFinalizer(rows, func(f interface{}) {
					f.(*Rows).Close()
				})
				return rows, nil
			}
		}

		rows.Cols, err = rows.inst.Columns()
		if err != nil {
			//println("P2 : ", err.Error(), "\n", queryStr)
			return nil, err
		}

		if !rows.inst.Next() {
			rows.Close()
		} else {
			runtime.SetFinalizer(rows, func(f interface{}) {
				f.(*Rows).Close()
			})
		}
	} else if db.instQL != nil {
		return nil, errors.New("ql not use TempQuery")
	} else {
		return nil, errors.New("db is not initialized")
	}

	QUERYSTR := strings.ToUpper(queryStr)
	if strings.HasPrefix(QUERYSTR, "INSERT") && strings.Contains(QUERYSTR, "OUTPUT") && strings.Contains(QUERYSTR, "INSERTED.") {
		if rows.IsNil() {
			return nil, errors.New("insert.fail")
		}
	}

	return rows, nil
}

func (rows *Rows) Next() bool {
	if !rows.isNil && rows.isFirst {
		rows.isFirst = false
		return true
	}
	if rows.inst != nil {
		if !rows.inst.Next() {
			rows.Close()
		}
	} else if rows.qlRows != nil {
		rows.qlIndex++
		if len(rows.qlRows) <= rows.qlIndex {
			rows.Close()
		}
	} else {
		return false
	}
	return !rows.isNil
}

func (rows *Rows) FetchArray() []interface{} {
	if rows.isNil {
		return nil
	}
	if rows.inst != nil {
		cols, err := rows.inst.Columns()
		if err != nil {
			return nil
		}

		rawResult := make([]*interface{}, len(cols))
		result := make([]interface{}, len(cols))

		dest := make([]interface{}, len(cols))
		for i, _ := range rawResult {
			dest[i] = &rawResult[i]
		}
		rows.inst.Scan(dest...)
		for i, raw := range rawResult {
			if raw != nil {
				v := (*raw)
				switch v.(type) {
				case []byte:
					v = convert.String(v)
				}
				if rows.isForceUTF8 {
					switch v.(type) {
					case string:
						v = convert.UTF8(v.(string))
					}
				}
				result[i] = v
			} else {
				result[i] = nil
			}
		}
		return result
	} else if rows.qlRows != nil {
		if len(rows.qlRows) <= rows.qlIndex {
			return nil
		}
		return rows.qlRows[rows.qlIndex]
	} else {
		return nil
	}

}

func (rows *Rows) FetchHash() map[string]interface{} {
	if rows.isNil {
		return nil
	}
	cols, err := rows.inst.Columns()
	if err != nil {
		return nil
	}

	result := make(map[string]interface{}, len(cols))

	row := rows.FetchArray()

	for i, v := range row {
		if v != nil {
			switch v.(type) {
			case []byte:
				v = convert.String(v)
			}
			if rows.isForceUTF8 {
				switch v.(type) {
				case string:
					v = convert.UTF8(v.(string))
				}
			}
		}
		result[cols[i]] = v
		result[strings.ToUpper(cols[i])] = v
		result[strings.ToLower(cols[i])] = v
	}
	return result
}

func (rows *Rows) Close() error {
	if rows != nil {
		rows.isNil = true
		if rows.inst != nil {
			return rows.inst.Close()
		}
	}
	return nil
}

func (rows *Rows) IsNil() bool {
	return rows.isNil
}
