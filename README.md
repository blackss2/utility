# utility
Set of utilities

# Convert
convert value to target format
simple & useful

- convert.Int(v interface{}) int64
change v to int64 (return 0 when fail)

- convert.IntWith(v interface{}, defatulValue int64) int64
change v to int64 (return defaultValue when fail)

- convert.Float(v interface{}) float64
change v to float64 (return 0 when fail)

- convert.FloatWith(v interface{}, defatulValue float64) float64
change v to float64 (return defaultValue when fail)

- convert.String(v interface{}) string
change v to string
nil => ""
time => "YYYY-MM-DD hh:mm:ss"
else => fmt.Sprintf("%v")

- convert.QueryString(val string) string
replace ' in string to '' for SQL

- convert.MD5(src string) string
chnage src to MD5 hexa formatted string


# Database
Basic DB Wrapper with DBPool
automatically reconnect when disconnected
using driver
mssql : "github.com/denisenkom/go-mssqldb"
mysql :	"github.com/go-sql-driver/mysql"
mymysql : "github.com/ziutek/mymysql/godrv"

- Usage
database.CreateDBPool(DBTYPE, IP_ADDRESS, PORT, DBNAME, ID, PASSWORD, POOLSIZE)

(ex)

import (
	"github.com/blackss2/utility/database"
	"fmt"
)

var dbPool *database.DBPool //make db pool(recommend as global variable)
func init() {
  //arg[0](string) : mssql, mysql, mymysql
  //arg[1](string) : ip or domain
  //arg[2](int) : port for database
  //arg[3](string) : database name
  //arg[4](string) : id
  //arg[5](string) : password
  //arg[6](int) : poolsize(size of concurrent query execution)
  dbPool = database.CreateDBPool("mssql", "127.0.0.1", 1433, "testdb", "testuser", "testpassword", 20) //18080
}

func main() {
  db := dbPool.GetDB()
  defer dbPool.ReleaseDB(db)
  queryStr := `SELECT userid, userpw, usertype FROM t_user`
  
  rows, err := db.Query(queryStr)
  if err != nil {
  	panic(err)
  } else {
    //rows.Cols : column name list
  	for rows.Next() {
  	  // rows.FetchHash() will return hash(column_name, value)
  		row := rows.FetchArray() // return current row as result([]interface{})
  		
  		userid := row[0].(string)
  		userpw := row[1].(string)
  		usertype := row[2].(int64)
  		
  		fmt.Println(userid, userpw, usertype)
  	}
  }
}


# htmlwriter
Simple HTML Builder
Escaping for attr, class, style is not implemented

- htmlwriter.CreateHtmlNode(name string) *htmlwriter.HtmlNode
create root node

- (*htmlwriter.HtmlNode) Add(name string) *htmlwriter.HtmlNode
add & return child node

- (*htmlwriter.HtmlNode) InsertAfter(c *HtmlNode) *htmlwriter.HtmlNode
insert node after target node

- (*htmlwriter.HtmlNode) Eq(idx int) *htmlwriter.HtmlNode
return child at index

- (*htmlwriter.HtmlNode) Append(c *HtmlNode) *htmlwriter.HtmlNode
append target node(c node is child node)

- (*htmlwriter.HtmlNode) AppendTo(c *HtmlNode) *htmlwriter.HtmlNode
append to target node(c node is parent node)

- (*htmlwriter.HtmlNode) Detach() *htmlwriter.HtmlNode
remove node from parent

- (*htmlwriter.HtmlNode) Remove(c *HtmlNode) *htmlwriter.HtmlNode
remove target node(c node is child node)

- (*htmlwriter.HtmlNode) SetText(text string) *htmlwriter.HtmlNode
set text(can be used with children)

- (*htmlwriter.HtmlNode) Write(buffer *bytes.Buffer)
* should be changed to standard Write function
write html to buffer

- (*htmlwriter.HtmlNode) Class(name string) *htmlwriter.HtmlNode
add class

- (*htmlwriter.HtmlNode) RemoveClass(name string) *htmlwriter.HtmlNode
remove class

- (*htmlwriter.HtmlNode) Style(name string, value string) *htmlwriter.HtmlNode
add style

- (*htmlwriter.HtmlNode) RemoveStyle(name string) *htmlwriter.HtmlNode
remove style

- (*htmlwriter.HtmlNode) Attr(name string, value string) *htmlwriter.HtmlNode
add attribute

- (*htmlwriter.HtmlNode) RemoveAttr(name string) *htmlwriter.HtmlNode
remove attribute

jDiv := htmlwriter.CreateHtmlNode("div")
jTr := jDiv.Add("table").Class("table-bordered").Attr("id", "TABLEID")
jTr.Add("td").SetText("Cell1").Add("td").SetText("Cell2").Attr("colspan", "2")
jTr.Add("td").SetText("Cell1").Add("td").SetText("Cell2").Add("td").SetText("Cell3")

var buffer bytes.Buffer
jDiv.Write(&buffer)//not standard write function
fmt.Println(buffer.String())
