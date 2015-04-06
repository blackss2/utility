package sqlQuery

import (
	"bytes"
	"fmt"
	"strings"
)

type Query struct {
	table  string
	output []string
	column []string
	data   []interface{}
	raw    []bool
}

type Insert Query
type Update Query
type Where Query

var maxCol = 16;
func CreateInsert(table string) *Query {
	return &Query{table, make([]string, 0, maxCol), make([]string, 0, maxCol), make([]interface{}, 0, maxCol), make([]bool, 0, maxCol)}
}

func (q *Query) AddOutput(ss ...string) {
	for _, s := range ss {
		q.output = append(q.output, s)	
	}
}

func (q *Query) AddColumn(ss ...string) {
	for _, s := range ss {
		q.column = append(q.column, s)
	}
}

func (q *Query) AddData(s interface{}) {
	q.data = append(q.data, s)
	q.raw = append(q.raw, false)
}

func (q *Query) AddDataWithNull(ss ...string) {
	for _, s := range ss {
		if len(s) > 0 {
				q.data = append(q.data, s)
			} else {
				q.data = append(q.data, nil)
		}
		q.raw = append(q.raw, false)
	}
}

func (q *Query) AddRawData(s interface{}) {
	q.data = append(q.data, s)
	q.raw = append(q.raw, true)
}

func (q *Query) ClearData() {
	q.data = make([]interface{}, 0, maxCol)
	q.raw = make([]bool, 0, maxCol)
}

func (q *Query) ToInsertString() string {
	var buffer bytes.Buffer
	buffer.WriteString("INSERT INTO ")
	buffer.WriteString(q.table)
	buffer.WriteString("\n")
	if len(q.column) > 0 {
		buffer.WriteString("(")
		for i := 0; i < len(q.column); i++ {
			if i > 0 {
				buffer.WriteString(", ")
			}
			s := q.column[i]
			buffer.WriteString(s)
		}
		buffer.WriteString(")\n")
	}
	if len(q.output) > 0 {
		buffer.WriteString("OUTPUT")
		for i := 0; i < len(q.output); i++ {
			if i > 0 {
				buffer.WriteString(",")
			}
			s := q.output[i]
			buffer.WriteString(" inserted.")
			buffer.WriteString(s)
		}
		buffer.WriteString("\n")
	}
	buffer.WriteString("VALUES")
	buffer.WriteString("(")
	for i := 0; i < len(q.data); i++ {
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(_s(q.data[i], q.raw[i]))
	}
	buffer.WriteString(")\n")

	return buffer.String()
}

// UPDATE table_name SET column_name = value [, column_name = value ...] [WHERE condition]
func (q *Insert) toUpdateString(where *Where) string {
	var buffer bytes.Buffer
	buffer.WriteString("UPDATE ")
	buffer.WriteString(q.table)
	buffer.WriteString("\n")
	buffer.WriteString("SET ")
	
	if len(q.column) > 0 && len(q.column) == len(q.data) {
		for i:=0; i < len(q.column);i++ {
			buffer.WriteString( fmt.Sprintf("%s = %s", q.column[i], _s(q.data[i], q.raw[i])))	
		}
	}
	if len(q.output) > 0 {
		buffer.WriteString("OUTPUT")
		for i := 0; i < len(q.output); i++ {
			if i > 0 {
				buffer.WriteString(",")
			}
			s := q.output[i]
			buffer.WriteString(" inserted.")
			buffer.WriteString(s)
		}
		buffer.WriteString("\n")
	}
	if where != nil {
		
	}

	return buffer.String()
}
// Utils
func typeToString(s interface{}) string {
	switch s.(type) {
	case int64:
		return fmt.Sprintf("%d", s.(int64))
	case int32:
		return fmt.Sprintf("%d", s.(int32))
	case float64:
		return fmt.Sprintf("%f", s.(float64))
	case float32:
		return fmt.Sprintf("%f", s.(float32))
	case string:
		return s.(string)
	}
	return ""
}
func TypeToString(s interface{}) string {
	switch s.(type) {
	case int64:
		return fmt.Sprintf("%d", s.(int64))
	case int32:
		return fmt.Sprintf("%d", s.(int32))
	case float64:
		return fmt.Sprintf("%f", s.(float64))
	case float32:
		return fmt.Sprintf("%f", s.(float32))
	case string:
		return s.(string)
	}
	return ""
}

func _s(data interface{}, isRaw bool) (s string) {
	s = typeToString(data)
	if len(s) > 0 {
		s = "'" + strings.Replace(s, "'", "''", -1) + "'"
	} else if !isRaw {
		s = "NULL"
	}
	return
}

func S(data interface{}, isRaw bool) (s string) {
	s = typeToString(data)
	if data == nil {
		s = "NULL"
	} else if !isRaw {
		s = "'" + strings.Replace(s, "'", "''", -1) + "'"
	}
	return
}
/*
func hasQueryError(queryStr string, err error, ms *interface {}, errMsg string) bool {
	if err != nil {
		ms.Error = 1
		println(queryStr)
		println(err.Error())
		ms.ErrorMsg = "permission_assign_approve.UPDATE.fail"
		return true
	} 
	return false
}
*/

