package database

import (
	"database/sql"
	"errors"
	"github.com/blackss2/utility/convert"
	"runtime"
	"strings"
)

type Database struct {
	inst        *sql.DB
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
	db.inst, err = sql.Open(db.driver, db.connString)
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
	}
	return err
}

type Rows struct {
	inst        *sql.Rows
	isFirst     bool
	isNil       bool
	Cols        []string
	isForceUTF8 bool
}

func (db *Database) Query(queryStr string) (*Rows, error) {
	rows := &Rows{nil, true, false, make([]string, 0, 100), db.isForceUTF8}

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
	rows := &Rows{nil, true, false, make([]string, 0, 100), db.isForceUTF8}

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
