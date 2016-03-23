package main

import (
	"fmt"
	"github.com/blackss2/utility/convert"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	cache "github.com/patrickmn/go-cache"
	"github.com/rjeczalik/notify"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func Main(fn func(string)) {
	MainWithPort("", fn)
}

func MainWithPort(port string, fn func(string)) {
	if len(os.Getenv("PORT")) > 0 {
		hotServerWatcher()

		WebPort := os.Getenv("PORT")
		if len(WebPort) == 0 {
			if len(port) == 0 {
				WebPort = "80"
			} else {
				WebPort = port
			}
		}
		fn(WebPort)
	} else {
		hotServerMain(port)
	}
}

func hotServerWatcher() {
	HotPort := os.Getenv("HOT_PORT")
	if len(HotPort) > 0 {
		go func() {
			u := url.URL{Scheme: "ws", Host: "127.0.0.1:" + HotPort, Path: "/__alive_websocket"}
			log.Printf("connecting to %s", u.String())

			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Fatalln("dial:", err)
			}

			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					log.Fatalln(err)
				}
			}
		}()
	}
}

func hotServerMain(port string) {
	if len(port) == 0 {
		port = "80"
	}

	SrcPath, err := filepath.Abs(".")
	if err != nil {
		log.Fatalln(err)
	}

	isRequireReload := true
	newFileWatcher(SrcPath+"/src", func(ev string, path string) {
		isRequireReload = true
	})

	os.Setenv("GOPATH", os.Getenv("GOPATH")+";"+SrcPath)

	var WebServer *os.Process
	var WebPort string
	r := gin.New()
	r.Use(gin.Recovery())

	wsupgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	r.GET("/__alive_websocket", func(c *gin.Context) {
		conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Fatalln("Failed to set websocket upgrade: %+v", err)
			return
		}

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if WebServer != nil {
					log.Println("OldServer Forcely Closed")
					isRequireReload = true
				}
				return
			}
		}
	})
	r.NoRoute(func(c *gin.Context) {
		if isRequireReload {
			if WebServer != nil {
				WebServer.Kill()
				WebServer = nil

				log.Println("Close OldServer")
			}

			isBuildSuccess := true
			if true {
				//Reload Server
				cmd := exec.Command("go", "build", "-v", "-o", "WEB_SERVER.exe", "main")
				res, err := cmd.CombinedOutput()
				if err != nil {
					isBuildSuccess = false
					log.Println("TODO : build fail")
					log.Print(string(res))
				}
			}

			if isBuildSuccess {
				isRequireReload = false

				cmd := exec.Command(SrcPath + "/WEB_SERVER.exe")

				WebPort = convert.String(rand.Int63n(10000) + 50000)
				cmd.Env = append(os.Environ(), []string{"HOT_PORT=" + port, "PORT=" + WebPort}...)

				se, err := cmd.StderrPipe()
				if err != nil {
					log.Fatalln(err)
				}
				go func() {
					for {
						data := make([]byte, 1024)
						n, err := se.Read(data)
						if err != nil {
							return
						}
						log.Print(string(data[:n]))
					}
				}()
				sp, err := cmd.StdoutPipe()
				if err != nil {
					log.Fatalln(err)
				}
				go func() {
					for {
						data := make([]byte, 1024)
						n, err := sp.Read(data)
						if err != nil {
							return
						}
						fmt.Print(string(data[:n]))
					}
				}()

				err = cmd.Start()
				if err != nil {
					log.Fatalln("exec.Command(SrcPath/WEB_SERVER.exe)", err)
				}
				WebServer = cmd.Process
				log.Printf("Start NewServer [:%s]\n", WebPort)
			}
		}

		proxy := &httputil.ReverseProxy{Director: func(req *http.Request) {
			req = c.Request
			req.URL.Scheme = "http"
			req.URL.Host = "127.0.0.1:" + WebPort
		}}
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	err = r.Run(":" + port)
	if err != nil {
		log.Fatalln(err)
	}
}

func newFileWatcher(path string, fn func(ev string, path string)) {
	go func() {
		c := make(chan notify.EventInfo, 1)
		err := notify.Watch(path+"/...", c, notify.All)
		if err != nil {
			log.Fatal(err)
		}
		defer notify.Stop(c)

		timeCache := cache.New(5*time.Minute, 30*time.Second)

		for {
			// Block until an event is received.
			ei := <-c
			eventPath := ei.Path()
			st, err := os.Stat(eventPath)
			if err != nil {
				timeCache.Delete(eventPath)
				continue
			}

			modTime := st.ModTime()
			isEnable := true
			if v, has := timeCache.Get(eventPath); has {
				if oldTime, is := v.(time.Time); is {
					if oldTime.Equal(modTime) {
						isEnable = false
					}
				}
			}
			if isEnable {
				fn(ei.Event().String()[7:], eventPath)
			}
			timeCache.Set(eventPath, modTime, cache.DefaultExpiration)
		}
	}()
}
