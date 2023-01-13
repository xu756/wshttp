package gateway

import (
	"errors"
	"sync"
)

// Room 一个房间代表一个订阅推送类型
type Room struct {
	id    string
	Name  string
	RConn sync.Map
}

func newRoom(id string, name string) *Room {
	return &Room{
		id:   id,
		Name: name,
	}
}

func (r *Room) JoinRoom(ws *WsConnection) error {
	if _, ok := r.RConn.Load(ws.GetUid()); ok {
		return errors.New("already exists")
	}
	r.RConn.Store(ws.GetWsrId(), ws)
	return nil
}

func (r *Room) LeaveRoom(uid string) error {
	if _, ok := r.RConn.Load(uid); !ok {
		return errors.New("already delete")
	}
	r.RConn.Delete(uid)
	return nil
}

func (r *Room) Count() int {
	var count int
	r.RConn.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

func (r *Room) Push(msg *WSMessage) {
	var (
		ws *WsConnection
		ok bool
	)
	r.RConn.Range(func(_, value interface{}) bool {
		if ws, ok = value.(*WsConnection); ok {
			_ = ws.SendMsg(msg)
		}
		return true
	})
}
