package gateway

import "log"

type NewAllMessage struct {
	Msg  string      `json:"msg"`
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

type NewGroupMessage struct {
	Msg    string      `json:"msg"`
	Code   int         `json:"code"`
	RoomId int         `json:"roomId"`
	Data   interface{} `json:"data"`
}
type RoomMessage struct {
	WsId   string `json:"wsId"`
	RoomId int    `json:"roomId"`
}

// PushAll 全部推送
func PushAll(code int, msg string, data interface{}) {
	GetPushManage().Push(&PushJob{
		Type:     1,
		Code:     code,
		PushType: msg,
		Data:     data,
	})
}

// PushRoom 房间推送
func PushRoom(code int, roomId string, msg string, data interface{}) {

	GetPushManage().Push(&PushJob{
		Type:     2,
		Code:     code,
		PushType: msg,
		RoomId:   roomId,
		Data:     data,
	})
}

// JoinRoom 加入房间
func JoinRoom(roomId string, wsId string) error {
	err := GetRoomManage().AddRoom(roomId, wsId)
	if err != nil {
		return err
	}
	return nil
}

// LeaveRoom 离开房间
func LeaveRoom(roomId string, wsId string) error {
	err := GetRoomManage().LeaveRoom(roomId, wsId)
	if err != nil {
		return err
	}
	return nil
}

// Getrooms 获取所有房间
func Getrooms() int {
	return GetRoomManage().Getrooms()
}

// GetRoomCount 获取房间人数
func GetRoomCount(roomId string) int {
	count, err := GetRoomManage().GetRoomCount(roomId)
	if err != nil {
		return 0
	}
	return count
}

// GetRoomUser 获取所有房间的用户
func GetRoomUser() []map[string]interface{} {
	return GetRoomManage().GetRoomList()
}

// GetRoomUsers 获取房间用户
func GetRoomUsers(roomId int) []map[string]interface{} {
	return GetRoomManage().GetRoomUsers(roomId)
}

// CreateRoom 创建房间
func CreateRoom(id, Name string) (string, error) {

	err := GetRoomManage().NewRoom(id, Name)
	if err != nil {
		log.Print("创建错误", err)
		return "", err
	}
	return id, nil
}

// IsUserInRoom 查找用户是否在房间中
func IsUserInRoom(roomId string, wsId string) bool {
	return GetRoomManage().IsUserInRoom(roomId, wsId)
}
