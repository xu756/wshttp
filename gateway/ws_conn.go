package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"log"
	"sync"
	"time"
)

type WsConnection struct {
	mu        sync.Mutex
	wsId      string // 用户id  user_id staff_id
	wsSort    int    // 用户类型 1:普通用户 2:职工
	Uid       string // 用户 uid  唯一标识
	ws        *websocket.Conn
	readChan  chan *WSMessage
	writeChan chan *WSMessage
	closeChan chan bool
	isOpen    bool
	addRoom   *sync.Map
	clientIp  string
}

func FormatwsId(wsId string) string {
	return fmt.Sprintf("%s:%s", "wsid", wsId)

}

type WSMessage struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func SetWsCache(key, value string) error {
	redisObj := redis.New("localhost:5379")
	return redisObj.Set(key, value)
}
func GetwsCache(key string) string {
	redisObj := redis.New("localhost:5379")
	val, err := redisObj.Get(key)
	if err != nil {
		return ""
	}
	return val
}
func ExistswsCache(key string) bool {
	redisObj := redis.New("localhost:5379")
	is, err := redisObj.Exists(key)
	if err != nil {
		return false
	}
	return is
}

var (
	WsErrConnLoss = errors.New("conn already close")
)

func NewWsConnection(conn *websocket.Conn) *WsConnection {
	ws := &WsConnection{}
	ws.ws = conn
	ws.readChan = make(chan *WSMessage, 10)
	ws.writeChan = make(chan *WSMessage, 10)
	ws.closeChan = make(chan bool)
	ws.isOpen = true
	ws.addRoom = new(sync.Map)
	go ws.read()
	go ws.send()
	return ws
}
func Init(w *WsConnection) {
	GetRoomManage().AllConn.Load(w)
	err := JoinRoom("all", w.wsId)
	if err != nil {
		fmt.Println("加入全体房间失败", err)
		return
	}
	roomid, cerr := CreateRoom(w.wsId, w.Uid)
	if cerr != nil {
		fmt.Println("创建个人房间失败", cerr)
		w.close()
		return
	}
	jerr := JoinRoom(roomid, w.wsId)
	if jerr != nil {
		_, nerr := CreateRoom(w.wsId, w.Uid)
		if nerr != nil {
			fmt.Println("创建个人房间失败", nerr)
			w.close()
			return
		}
	}
	_ = w.SendMsg(&WSMessage{
		Code: 111,
		Data: map[string]interface{}{"wsId": w.GetWsrId(), "uid": w.GetUid()},
	})
}

func (w *WsConnection) SetWsrId(sort string, id int) {
	w.wsId = fmt.Sprintf("%s_%d", sort, id)
}
func (w *WsConnection) GetWsrId() string {
	return w.wsId
}
func (w *WsConnection) SetWsSort(id int) {
	w.wsSort = id
}
func (w *WsConnection) GetWsSort() int {
	return w.wsSort
}

func (w *WsConnection) SetUid() {
	uid := uuid.NewV5(uuid.NewV4(), "ws").String()
	w.Uid = uid
	oldUid := GetwsCache(FormatwsId(w.wsId))
	if oldUid != "" {
		//删除旧的连接
		GetRoomManage().CloseOldConn(w.wsId, oldUid)
	}

	err := SetWsCache(FormatwsId(w.wsId), uid)
	if err != nil {
		log.Print("设置uid失败", err)
	}
}
func (w *WsConnection) GetUid() string {
	return w.Uid
}

func (w *WsConnection) SetIp(ip string) {
	w.clientIp = ip
}
func (w *WsConnection) GetIp() string {
	return w.clientIp
}

func (w *WsConnection) read() {
	var (
		Data []byte
		err  error
	)
	w.ws.SetReadLimit(1024)
	_ = w.ws.SetReadDeadline(time.Now().Add(time.Second * 1000))
	for {
		if _, Data, err = w.ws.ReadMessage(); err != nil {
			w.close()
			return
		}
		message := &WSMessage{}
		if err = json.Unmarshal(Data, message); err != nil {
			w.close()
			return
		}
		select {
		case w.readChan <- message:
		case <-w.closeChan:
			return
		}
	}
}
func (w *WsConnection) send() {
	var (
		err     error
		message *WSMessage
	)
	for {
		select {
		case message = <-w.writeChan:
			if err = w.ws.WriteJSON(&message); err != nil {
				w.close()
			}
		case <-w.closeChan:
			return
		}
	}
}

func (w *WsConnection) ReadMsg() (message *WSMessage, err error) {
	select {
	case message = <-w.readChan:
	case <-w.closeChan:
		err = WsErrConnLoss
	}
	return
}

func (w *WsConnection) SendMsg(msg *WSMessage) (err error) {
	select {
	case w.writeChan <- msg:
	case <-w.closeChan:
		err = WsErrConnLoss
	}
	return
}

func (w *WsConnection) close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.isOpen {
		GetRoomManage().AllRoom.Delete(w.wsId)
		GetRoomManage().DelUserAllRoom(w.Uid)
		fmt.Println("关闭链接: ", w.GetIp(), "用户Wsid", w.GetWsrId())
		_ = w.ws.Close()
		w.isOpen = false
		w.closeChan <- true
	}
}
