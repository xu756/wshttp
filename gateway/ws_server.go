package gateway

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var wsUpgrader = websocket.Upgrader{
	// 允许所有CORS跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func InitWsServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/connect", getConn)
	NewRoomManage()
	//初始化房间
	NewPushTask(3000)
	go pushTask.distributionTask()
	err := GetRoomManage().NewRoom("all", "all")
	if err != nil {
		fmt.Println("初始化房间失败", err)
		return
	}
	// HTTP服务
	server := http.Server{
		Addr:         "0.0.0.0:7088",
		ReadTimeout:  time.Duration(10) * time.Millisecond,
		WriteTimeout: time.Duration(10) * time.Millisecond,
		Handler:      mux,
	}
	fmt.Println("启动WS服务器成功 ：", 7088)
	_ = server.ListenAndServe()

}

func getConn(res http.ResponseWriter, req *http.Request) {
	var (
		err    error
		wsConn *websocket.Conn
	)
	if wsConn, err = wsUpgrader.Upgrade(res, req, nil); err != nil {
		return
	}
	ws := NewWsConnection(wsConn)
	ws.SetIp(ClientIP(req))
	ws.SetWsrId(GetUserSort(req), GetUserId(req))
	ws.SetUid()
	GetRoomManage().AddConn(ws)
	Init(ws)
	ws.WsHandle()
}

func ClientIP(c *http.Request) string {
	clientIP := c.Header.Get("X-Forwarded-For")
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if clientIP == "" {
		clientIP = strings.TrimSpace(c.Header.Get(("X-Real-Ip")))
	}
	if clientIP != "" {
		return clientIP
	}
	addr := c.Header.Get("X-Appengine-Remote-Addr")
	if addr != "" {
		return addr
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}
func GetUserId(c *http.Request) int {
	// 从from表单中获取用户ID
	userId, err := strconv.Atoi(c.FormValue("userId"))
	if err != nil {
		return 0
	}
	return userId

}
func GetUserSort(c *http.Request) string {
	userSort := c.FormValue("userSort")
	if userSort == "" {
		return "user"
	}
	return userSort
}
