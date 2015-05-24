package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/blackss2/utility/convert"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

/*
//API Register
router := api.Default()
apiRouter := router.Group("/api")
apiRouter.RequirePermission(true)
apiStoragesRouter := apiRouter.Group("/storages")
apiStoragesRouter.RequirePermission(true)
apiStoragesRouter.PUT("/", func(c *api.Context) {
	c.Abort(http.StatusOK)
})
apiStoragesRouter.GET("/", func(c *api.Context) {
	c.Resolve(http.StatusOK, "123")
})
apiStoragesRouter.POST("/", func(c *api.Context) {
	c.Resolve(http.StatusOK, nil)
})
apiStoragesRouter.DELETE("/", func(c *api.Context) {
	c.Abort(http.StatusOK)
})
*/
type Engine struct {
	gin   *gin.Engine
	addr  string
}

type EngineGroup struct {
	engine      *Engine
	children    []*EngineGroup
	routerGroup *gin.RouterGroup
	path        string
	handlerHash map[string][]localAPIHandler
}

type localAPIHandler func(*Context)

type APIHandler func(*Context)

func Default(addr string) *EngineGroup {
	engine := &Engine{
		gin:   gin.Default(),
		addr:  addr,
	}
	router := &EngineGroup{
		engine:      engine,
		children:    make([]*EngineGroup, 0, 16),
		routerGroup: nil,
		path:        "",
		handlerHash: make(map[string][]localAPIHandler),
	}
	engine.gin.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))     //TEMP
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")                        //TEMP
		c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT, GET, POST, DELETE, OPTIONS") //TEMP
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")                    //TEMP
		c.Next()
	})
	engine.gin.OPTIONS("/*all", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusOK) //TEMP
	})
	apiLocalSupportRegister(addr, router)
	return router
}

func (this *EngineGroup) Group(path string) *EngineGroup {
	var routerGroup *gin.RouterGroup
	for _, child := range this.children {
		if path == child.path {
			return child
		}
	}
	if this.routerGroup != nil {
		routerGroup = this.routerGroup.Group(path)
	} else {
		routerGroup = this.engine.gin.Group(path)
	}
	child := &EngineGroup{
		engine:      this.engine,
		routerGroup: routerGroup,
		path:        path,
		handlerHash: make(map[string][]localAPIHandler),
	}
	this.children = append(this.children, child)
	return child
}

const (
	METHOD_PUT    = iota
	METHOD_GET    = iota
	METHOD_POST   = iota
	METHOD_DELETE = iota
	method_COUNT  = iota
)

func (this *EngineGroup) register(method int, path string, handler APIHandler) {
	apiList, has := this.handlerHash[path]
	if !has {
		apiList = make([]localAPIHandler, method_COUNT)
		this.handlerHash[path] = apiList
	}
	apiList[method] = func(c *Context) {
		handler(c)
	}

	handlerImp := this.getHandlerImp(handler)
	switch method {
	case METHOD_PUT:
		if this.routerGroup != nil {
			this.routerGroup.PUT(path, handlerImp)
		} else {
			this.engine.gin.PUT(path, handlerImp)
		}
	case METHOD_GET:
		if this.routerGroup != nil {
			this.routerGroup.GET(path, handlerImp)
		} else {
			this.engine.gin.GET(path, handlerImp)
		}
	case METHOD_POST:
		if this.routerGroup != nil {
			this.routerGroup.POST(path, handlerImp)
		} else {
			this.engine.gin.POST(path, handlerImp)
		}
	case METHOD_DELETE:
		if this.routerGroup != nil {
			this.routerGroup.DELETE(path, handlerImp)
		} else {
			this.engine.gin.DELETE(path, handlerImp)
		}
	}
}

func (this *EngineGroup) PUT(path string, handler APIHandler) {
	this.register(METHOD_PUT, path, handler)
}

func (this *EngineGroup) GET(path string, handler APIHandler) {
	this.register(METHOD_GET, path, handler)
}

func (this *EngineGroup) POST(path string, handler APIHandler) {
	this.register(METHOD_POST, path, handler)
}

func (this *EngineGroup) DELETE(path string, handler APIHandler) {
	this.register(METHOD_DELETE, path, handler)
}

func (this *EngineGroup) getHandlerImp(handler APIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data interface{}
		body, _ := ioutil.ReadAll(c.Request.Body)
		err := json.Unmarshal(body, &data)
		if err != nil {
			var buffer bytes.Buffer
			buffer.Write(body)
			c.Request.Body = ioutil.NopCloser(bufio.NewReader(&buffer))
		}
		context := &Context{
			code:    unresolvedCode,
			ret:     nil,
			Data:    data,
			Params:  c.Params,
			QParams: c.Request.URL.Query(),
			Request: c.Request,
			Writer:  c.Writer,
		}
		handler(context)
		switch context.code {
		case unresolvedCode:
			c.AbortWithStatus(http.StatusInternalServerError)
		case http.StatusMovedPermanently:
			c.Redirect(context.code, context.ret.(string))
		case http.StatusTemporaryRedirect:
			c.Redirect(context.code, context.ret.(string))
		default:
			switch context.ret.(type) {
			case nil:
				c.AbortWithStatus(context.code)
			case string:
				fallthrough
			case int:
				fallthrough
			case int64:
				fallthrough
			case float32:
				fallthrough
			case float64:
				c.Header("Content-Type", "text/html; charset=utf-8")
				c.String(context.code, convert.String(context.ret))
			default:
				c.JSON(context.code, context.ret)
			}
		}
	}
}

func (this *EngineGroup) Run() error {
	return this.engine.gin.Run(this.engine.addr)
}

func (this *EngineGroup) RunTLS(cert string, key string) error {
	return this.engine.gin.RunTLS(this.engine.addr, cert, key)
}
