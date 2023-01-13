package gateway

import (
	"errors"
	"sync"
)

var (
	manage *RoomManage
)

type RoomManage struct {
	AllRoom sync.Map
	AllConn sync.Map
}

func NewRoomManage() {
	manage = &RoomManage{}
	return
}
func GetRoomManage() *RoomManage {
	return manage
}

func (r *RoomManage) NewRoom(id string, name string) error {
	_, ok := r.AllRoom.Load(id)
	if ok {
		return errors.New("already exists")
	}
	r.AllRoom.Store(id, newRoom(id, name))
	return nil
}
func (r *RoomManage) AddConn(ws *WsConnection) {
	r.AllConn.Store(ws.Uid, ws)
}

func (r *RoomManage) DelConn(ws *WsConnection) {
	r.AllConn.Delete(ws.GetUid())
}

func (r *RoomManage) AddRoom(id string, wsId string) error {
	uid := GetwsCache(FormatwsId(wsId))
	var room *Room
	var ws *WsConnection
	val, ok := r.AllRoom.Load(id)
	if !ok {
		return errors.New("add not find room")
	}
	wsVal, ok := r.AllConn.Load(uid)
	if !ok {
		return errors.New("add not find conn")
	}
	room = val.(*Room)
	ws = wsVal.(*WsConnection)
	if err := room.JoinRoom(ws); err != nil {
		return err
	}
	ws.addRoom.Store(id, true)
	return nil
}

func (r *RoomManage) LeaveRoom(id string, wsId string) error {
	uid := GetwsCache(FormatwsId(wsId))
	var room *Room
	val, ok := r.AllRoom.Load(id)
	if !ok {
		return errors.New("leave not find room")
	}
	room = val.(*Room)
	if err := room.LeaveRoom(uid); err != nil {
		return err
	}
	if room.Count() <= 0 {
		//房间的人数为0 删除房间
		r.AllRoom.Delete(room.id)
	}
	return nil
}

func (r *RoomManage) PushRoom(id string, msg *WSMessage) error {
	val, ok := r.AllRoom.Load(id)
	if !ok {
		return errors.New("not find room")
	}
	room := val.(*Room)
	room.Push(msg)
	return nil
}

func (r *RoomManage) PushAll(msg *WSMessage) {
	r.AllConn.Range(func(_, value interface{}) bool {
		if ws, ok := value.(*WsConnection); ok {
			_ = ws.SendMsg(msg)
		}
		return true
	})
}

func (r *RoomManage) GetRoomCount(id string) (int, error) {
	val, ok := r.AllRoom.Load(id)
	if !ok {
		return 0, errors.New("not find room")
	}
	room := val.(*Room)
	return room.Count(), nil
}
func (r *RoomManage) Getrooms() int {
	var count int
	r.AllRoom.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// GetRoomList 获取所有房间
func (r *RoomManage) GetRoomList() []map[string]interface{} {
	var list []map[string]interface{}
	r.AllRoom.Range(func(key, value interface{}) bool {
		room := value.(*Room)
		list = append(list, map[string]interface{}{
			"id":    key,
			"name":  room.Name,
			"count": room.Count(),
		})
		return true
	})
	return list
}

func (r *RoomManage) GetRoomUsers(id int) []map[string]interface{} {
	val, ok := r.AllRoom.Load(id)
	if !ok {
		return nil
	}
	room := val.(*Room)
	var users []map[string]interface{}
	room.RConn.Range(func(key, value interface{}) bool {
		ws := value.(*WsConnection)
		users = append(users, map[string]interface{}{
			"id":   key,
			"wsid": ws.GetWsrId(),
			"uid":  ws.GetUid(),
		})
		return true
	})
	return users
}

// GetUserRooms 获取用户所在的房间
func (r *RoomManage) GetUserRooms(uid int) []map[string]interface{} {
	var rooms []map[string]interface{}
	r.AllRoom.Range(func(key, value interface{}) bool {
		room := value.(*Room)
		_, ok := room.RConn.Load(uid)
		if ok {
			rooms = append(rooms, map[string]interface{}{
				"id":    key,
				"name":  room.Name,
				"count": room.Count(),
			})
		}
		return true
	})
	return rooms
}

// IsUserInRoom 查找用户是否在房间中
func (r *RoomManage) IsUserInRoom(roomId string, wsId string) bool {
	uid := GetwsCache(FormatwsId(wsId))
	val, ok := r.AllRoom.Load(roomId)
	if !ok {
		return false
	}
	room := val.(*Room)
	_, ok = room.RConn.Load(uid)
	return ok
}

// DelUserAllRoom 删除用户所有房间
func (r *RoomManage) DelUserAllRoom(wsId string) {
	uid := GetwsCache(FormatwsId(wsId))
	r.AllRoom.Range(func(key, value interface{}) bool {
		room := value.(*Room)
		if room.id == "all" {
			return true
		}
		_ = room.LeaveRoom(uid)
		return true
	})
}

// CloseOldConn 关闭旧的连接
func (r *RoomManage) CloseOldConn(wsId, uid string) {
	val, ok := r.AllConn.LoadAndDelete(uid)
	if !ok {
		return
	}
	ws := val.(*WsConnection)
	PushRoom(404, wsId, "您的账号在其他地方登录", nil)
	ws.CloseConn()
}
