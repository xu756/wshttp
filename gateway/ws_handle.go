package gateway

import (
	"fmt"
	"log"
)

func (w *WsConnection) WsHandle() {
	var (
		err error
		msg *WSMessage
	)
	for {
		// 读取数据
		if msg, err = w.ReadMsg(); err != nil {
			w.CloseConn()
			return
		}
		switch {
		// 心跳
		case msg.Code == 0:
			//_ = w.ws.SetReadDeadline(time.Now().Add(time.Second * 10))
			_ = w.SendMsg(&WSMessage{Code: 0, Data: "PONG"})
		case msg.Code == 1:
			_ = w.SendMsg(&WSMessage{Code: 2, Data: GetRoomUser()})
		case msg.Code == 2:
			_, err := CreateRoom(msg.Msg, "ceshi")
			if err != nil {
				log.Print("创建房间失败", err)
			}
			_ = w.SendMsg(&WSMessage{Code: 2, Msg: "创建房间成功", Data: GetRoomUser()})
		case msg.Code == 3:
			jerr := JoinRoom(msg.Msg, w.GetWsrId())
			if jerr != nil {
				_ = w.SendMsg(&WSMessage{Code: 3, Msg: "加入房间失败", Data: jerr.Error()})
			}
			_ = w.SendMsg(&WSMessage{Code: 3, Msg: "加入房间成功", Data: "加入房间成功"})
		case msg.Code == 4:
			PushRoom(4, w.wsId, fmt.Sprintf("测试消息 %s", w.GetWsrId()), "房间数据")
		case msg.Code == 5:
			PushAll(5, "测试消息", "全局数据")

		default:
			fmt.Println("OTHER", msg.Code, msg.Data)

		}
	}

}

func (w *WsConnection) CloseConn() {
	w.close()
	GetRoomManage().DelConn(w)
	w.addRoom.Range(func(key, _ interface{}) bool {
		_ = GetRoomManage().LeaveRoom(key.(string), w.GetUid())
		w.addRoom.Delete(key)
		return true
	})
}
