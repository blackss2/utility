package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
)

const unresolvedCode = -1

type Context struct {
	code    int
	ret     interface{}
	Data    interface{}
	Params  gin.Params
	QParams url.Values
	Request *http.Request
	Writer  http.ResponseWriter
}

func (this *Context) Abort(code int) {
	this.code = code
}

func (this *Context) Resolve(code int, ret interface{}) {
	this.code = code
	this.ret = ret
}

func (this *Context) IsResolved() bool {
	return this.code != unresolvedCode
}

var localSupportHash map[string]*EngineGroup = make(map[string]*EngineGroup)

func apiLocalSupportRegister(addr string, router *EngineGroup) {
	localSupportHash[addr] = router
}

func updateHandlerHash(handlerHash map[string][]localAPIHandler, router *EngineGroup, path string) {
	for k, v := range router.handlerHash {
		handlerHash[path+k] = v
	}
	for _, c := range router.children {
		updateHandlerHash(handlerHash, c, path+c.path)
	}
}

func getLocalSupportHandlerHash(addr string) map[string][]localAPIHandler {
	if router, has := localSupportHash[addr]; !has {
		return nil
	} else {
		handlerHash := make(map[string][]localAPIHandler)
		updateHandlerHash(handlerHash, router, router.path)
		return handlerHash
	}
}

func getLocalSupportHandler(addr string, path string, method int) localAPIHandler {
	hash := getLocalSupportHandlerHash(addr)
	if hash != nil {
		if apiList, has := hash[path]; has {
			return apiList[method]
		}
	}
	return nil
}
