package api

import (
	"gocTask"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	httpServer *http.Server
}

var (
	GApiServer *Server
)

func InitServer() (err error) {
	var (
		listener   net.Listener
		mux        *http.ServeMux
		httpServer *http.Server
	)
	mux = http.NewServeMux()
	mux.HandleFunc("/task/add", handleTaskSave)

	httpServer = &http.Server{
		ReadTimeout:  time.Duration(gocTask.GConfig.ApiReadTimout) * time.Millisecond,
		WriteTimeout: time.Duration(gocTask.GConfig.ApiWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}

	GApiServer = &Server{
		httpServer: httpServer,
	}

	// 启动tcp监听
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(gocTask.GConfig.ApiPort)); err != nil {
		return
	}

	// 后台开启api服务
	go httpServer.Serve(listener)

	return
}

// 任务添加api
func handleTaskSave(w http.ResponseWriter, r *http.Request) {

}
