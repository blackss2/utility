package database

import (
	"database/sql"
	"errors"
	_ "github.com/alexbrainman/odbc"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/ziutek/mymysql/godrv"
	"runtime"
	"strings"
)

type Database struct {
	inst       *sql.DB
	connString string
	driver     string
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
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) Close() error {
	if db.inst != nil {
		return db.inst.Close()
	}
	return nil
}

type Rows struct {
	inst    *sql.Rows
	isFirst bool
	isNil   bool
	Cols    []string
}

func (db *Database) prepare(queryStr string, retCount int) (*sql.Stmt, error) {
	stmt, err := db.inst.Prepare(queryStr)
	if err != nil {
		db.Close()
		db.executeOpen()
		if retCount > 0 {
			return db.prepare(queryStr, retCount-1)
		}
		return nil, err
	}
	return stmt, err
}

func (db *Database) Query(queryStr string) (*Rows, error) {
	stmt, err := db.prepare(queryStr, 1)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		db.Close()
		db.executeOpen()
		return db.TempQuery(queryStr)
	}

	rows := &Rows{nil, true, false, make([]string, 0, 100)}
	rows.inst, err = stmt.Query()

	if err != nil {
		db.Close()
		db.executeOpen()
		return db.TempQuery(queryStr)
	}

	rows.Cols, err = rows.inst.Columns()

	if !rows.inst.Next() {
		rows.Close()
	}

	QUERYSTR := strings.ToUpper(queryStr)
	if strings.HasPrefix(QUERYSTR, "INSERT") && strings.Contains(QUERYSTR, "OUTPUT") && strings.Contains(QUERYSTR, "INSERTED.") {
		if rows.IsNil() {
			return nil, errors.New("insert.fail")
		}
	}

	runtime.SetFinalizer(rows, func(f interface{}) {
		f.(*Rows).Close()
	})

	return rows, nil
}

func (db *Database) TempQuery(queryStr string) (*Rows, error) {
	stmt, err := db.prepare(queryStr, 1)
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		println("P1 : ", err.Error())
		return nil, err
	}

	rows := &Rows{nil, true, false, make([]string, 0, 100)}
	rows.inst, err = stmt.Query()

	if err != nil {
		if err.Error() != "Stmt did not create a result set" {
			println("P2 : ", err.Error(), "\n", queryStr)
			return nil, err
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
	if !rows.inst.Next() {
		rows.Close()
	}
	return !rows.isNil
}

func (rows *Rows) FetchArray() []interface{} {
	if rows.isNil {
		return nil
	}
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
			result[i] = (*raw)
		} else {
			result[i] = nil
		}
	}
	return result
}

func (rows *Rows) FetchHash() map[string]interface{} {
	if rows.isNil {
		return nil
	}
	cols, err := rows.inst.Columns()
	if err != nil {
		return nil
	}

	rawResult := make([]*interface{}, len(cols))
	result := make(map[string]interface{}, len(cols))

	dest := make([]interface{}, len(cols))
	for i, _ := range rawResult {
		dest[i] = &rawResult[i]
	}
	rows.inst.Scan(dest...)
	for i, raw := range rawResult {
		if raw != nil {
			result[cols[i]] = (*raw)
			result[strings.ToUpper(cols[i])] = (*raw)
			result[strings.ToLower(cols[i])] = (*raw)
		} else {
			result[cols[i]] = nil
			result[strings.ToUpper(cols[i])] = nil
			result[strings.ToLower(cols[i])] = nil
		}
	}
	return result
}

func (rows *Rows) Close() error {
	if rows != nil && rows.inst != nil {
		rows.isNil = true
		return rows.inst.Close()
	}
	return nil
}

func (rows *Rows) IsNil() bool {
	return rows.isNil
}
