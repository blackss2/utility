package CURM

import (
	"bytes"
	"fmt"
	"strings"
)

// Insert Query
type Insert struct {
	table  string
	output []string
	column []string
	data   []interface{}
	raw    []bool
}

func CreateInsert(table string) *Insert {
	return &Insert{table, make([]string, 0, 16), make([]string, 0, 16), make([]interface{}, 0, 16), make([]bool, 0, 16)}
}

func (q *Insert) AddOutput(s string) {
	q.output = append(q.output, s)
}

func (q *Insert) AddColumn(s string) {
	q.column = append(q.column, s)
}

func (q *Insert) AddColumns(ss ...string) {
	for _, s := range ss {
		q.AddColumn(s)
	}
}

func (q *Insert) AddData(s interface{}) {
	q.data = append(q.data, s)
	q.raw = append(q.raw, false)
}

func (q *Insert) AddDataWithNull(s string) {
	if len(s) > 0 {
		q.data = append(q.data, s)
	} else {
		q.data = append(q.data, nil)
	}
	q.raw = append(q.raw, false)
}

func (q *Insert) AddRawData(s interface{}) {
	q.data = append(q.data, s)
	q.raw = append(q.raw, true)
}

func (q *Insert) ClearData() {
	q.data = make([]interface{}, 0, 16)
	q.raw = make([]bool, 0, 16)
}

func (q *Insert) String() string {
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
	if len(q.column) == 0 {
		buffer.WriteString("DEFAULT VALUES")
	} else {
		buffer.WriteString("VALUES")
		buffer.WriteString("(")
		for i := 0; i < len(q.data); i++ {
			if i > 0 {
				buffer.WriteString(", ")
			}
			data := q.data[i]
			isRaw := q.raw[i]

			s := typeToString(data)
			if data == nil {
				buffer.WriteString("NULL")
			} else {
				if isRaw {
					buffer.WriteString(typeToString(s))
				} else {
					buffer.WriteString("'")
					buffer.WriteString(strings.Replace(typeToString(s), "'", "''", -1))
					buffer.WriteString("'")
				}
			}
		}
		buffer.WriteString(")\n")
	}

	return buffer.String()
}

type Update struct {
	table  string
	output []string
	column []string
	data   []interface{}
	raw    []bool
	where  Where
}

type Where struct {
	columm []string
	data   []interface{}
	raw    []bool
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
