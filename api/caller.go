package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

/*
//API Caller
builder := api.Builder("http://localhost/api")
caller := builder.PUT("/storage/:id/datas/:did")
caller.Param.Set("id", 1)
caller.Param.Set("did", 3)
caller.QParam.Set("name", "testname")
code, ret := caller.Call()
caller:= caller.GET("/storage/:id/datas/:did")
caller.Param.Set("id", 1)
caller.Param.Set("did", 3)
caller.QParam.Set("offset", "1")
code, ret := caller.Call()
caller := caller.POST("/storage/:id/datas/:did")
caller.Param.Set("id", 1)
caller.Param.Set("did", 3)
caller.Data = "testbody"
code, ret := caller.Call()
caller := caller.DELETE("/storage/:id/datas/:did")
caller.Param.Set("id", 1)
caller.Param.Set("did", 3)
code, ret := caller.Call()
*/

type BuildEngine struct {
	host    string
	isLocal bool
	addr    string
	cookie  string
}

func Builder(host string) *BuildEngine {
	url, err := url.ParseRequestURI(host)
	if err != nil {
		panic(err)
	}
	addr := ""
	isLocal := strings.Index(url.Host, "localhost") >= 0 || strings.Index(url.Host, "127.0.0.1") >= 0
	if isLocal {
		idx := strings.Index(url.Host, ":")
		if idx >= 0 {
			addr = url.Host[idx:]
		} else {
			switch url.Scheme {
			case "http":
				addr = ":80"
			case "https":
				addr = ":443"
			default:
				panic("Unknown scheme : " + url.Scheme)
			}
		}
	}
	return &BuildEngine{
		host:    host,
		addr:    addr,
		isLocal: isLocal,
	}
}

func (this *BuildEngine) createCaller(url string, method int) *Caller {
	var handler localAPIHandler
	if this.isLocal {
		handler = getLocalSupportHandler(this.addr, url, method)
	}
	var client *http.Client
	if handler == nil {
		client = new(http.Client)
	}
	return &Caller{
		engine: this,
		url:    url,
		method: method,
		Params: &CallParams{
			Params: make([]httprouter.Param, 0),
		},
		QParams: make(map[string][]string),
		handler: handler,
		client:  client,
	}
}

func (this *BuildEngine) PUT(url string) *Caller {
	return this.createCaller(url, METHOD_PUT)
}
func (this *BuildEngine) GET(url string) *Caller {
	return this.createCaller(url, METHOD_GET)
}
func (this *BuildEngine) POST(url string) *Caller {
	return this.createCaller(url, METHOD_POST)
}
func (this *BuildEngine) DELETE(url string) *Caller {
	return this.createCaller(url, METHOD_DELETE)
}

type Caller struct {
	engine  *BuildEngine
	url     string
	method  int
	Params  *CallParams
	QParams url.Values
	Data    interface{}
	handler localAPIHandler
	client  *http.Client
}

func (this *Caller) Call() (code int, ret interface{}) {
	method := ""
	switch this.method {
	case METHOD_PUT:
		method = "PUT"
	case METHOD_GET:
		method = "GET"
	case METHOD_POST:
		method = "POST"
	case METHOD_DELETE:
		method = "DELETE"
	}
	if true {

	}
	var buffer bytes.Buffer
	buffer.WriteString(this.engine.host)
	urlPath := this.url
	if len(this.Params.Params) > 0 {
		urlList := strings.Split(urlPath, "/")
		urlHash := make(map[string]int)
		for i, v := range urlList {
			urlHash[strings.ToLower(v)] = i
		}
		for _, v := range this.Params.Params {
			if idx, has := urlHash[strings.ToLower(fmt.Sprintf(":%s", v.Key))]; has {
				urlList[idx] = v.Value
			}
		}
		urlPath = strings.Join(urlList, "/")
	}
	buffer.WriteString(urlPath)
	if len(this.QParams) > 0 {
		buffer.WriteString("?")
		isFirst := true
		for k, v := range this.QParams {
			if isFirst {
				isFirst = false
			} else {
				buffer.WriteString("&")
			}
			buffer.WriteString(k)
			buffer.WriteString("=")
			buffer.WriteString(strings.Join(v, ","))
		}
	}

	url := buffer.String()
	isRoutable := this.engine.isLocal && this.handler != nil

	var reader io.Reader
	if !isRoutable && this.Data != nil {
		b, err := json.Marshal(this.Data)
		if err != nil {
			panic(err)
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Cookie", this.engine.cookie)
	if isRoutable {
		context := &Context{
			code:    -1,
			ret:     nil,
			Data:    this.Data,
			Params:  this.Params.Params,
			QParams: this.QParams,
			Request: req,
			Writer:  newResponse(),
		}
		this.handler(context)
		this.engine.cookie = context.Writer.Header().Get("Set-Cookie")
		code = context.code
		ret = context.ret
	} else {
		req, err := http.NewRequest(method, url, reader)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Add("Cookie", this.engine.cookie)
		if err != nil {
			panic(err)
		}
		res, err := this.client.Do(req)
		if err != nil {
			panic(err)
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			//panic(err)
		}

		var reqRet interface{}
		if body != nil && len(body) > 0 {
			err := json.Unmarshal(body, &reqRet)
			if err != nil {
				reqRet = string(body)
			}
		}

		code = res.StatusCode
		ret = reqRet
	}
	return
}

type CallParams struct {
	httprouter.Params
}

func (this *CallParams) Set(key string, val string) {
	var param httprouter.Param
	param.Key = key
	param.Value = val
	has := false
	fmt.Println(key, "\t", val)
	for _, v := range this.Params {
		fmt.Print(v.Key, "\t", v.Value, "\t")
	}
	fmt.Println("")
	for i, v := range this.Params {
		if strings.ToLower(key) == strings.ToLower(v.Key) {
			this.Params[i].Value = val
			has = true
		}
	}
	if !has {
		this.Params = append(this.Params, param)
	}
	for _, v := range this.Params {
		fmt.Print(v.Key, "\t", v.Value, "\t")
	}
	fmt.Println("")
}

// response implements http.ResponseWriter.
type response struct {
	header      http.Header
	w           bytes.Buffer
	wroteHeader bool
}

func newResponse() *response {
	return &response{
		header: http.Header{},
	}
}

func (this *response) Header() http.Header {
	return this.header
}

func (this *response) Write(data []byte) (int, error) {
	return this.w.Write(data)
}

func (this *response) WriteHeader(code int) {
	if this.wroteHeader {
		return
	}
	this.wroteHeader = true
	if code == http.StatusNotModified {
		// Must not have body.
		this.header.Del("Content-Type")
		this.header.Del("Content-Length")
		this.header.Del("Transfer-Encoding")
	} else if this.header.Get("Content-Type") == "" {
		this.header.Set("Content-Type", "text/html; charset=utf-8")
	}

	if this.header.Get("Date") == "" {
		this.header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}

	w := bufio.NewWriter(&this.w)
	fmt.Fprintf(w, "Status: %d %s\r\n", code, http.StatusText(code))
	this.header.Write(w)
	this.w.WriteString("\r\n")
}
