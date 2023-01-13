package gateway

import (
	"fmt"
	"log"
)

type PushJob struct {
	Code     int         // 返回码
	Type     int         // 1: 全部推送 2: 房间推送
	PushType string      // 推送类型
	RoomId   string      // 房间id
	Data     interface{} // 推送内容
}

var pushTask *PushTask

type PushTask struct {
	JobChan          map[string]chan *PushJob
	DistributionTask chan *PushJob
}

type PushManage interface {
	Push(job *PushJob)
}

func GetPushManage() PushManage {
	return pushTask
}

func NewPushTask(taskNum int) {

	pushTask = &PushTask{
		JobChan:          make(map[string]chan *PushJob),
		DistributionTask: make(chan *PushJob, taskNum),
	}
}

func (p *PushTask) Push(job *PushJob) {
	p.DistributionTask <- job
}

func (p *PushTask) distributionTask() {
	var (
		pushJob *PushJob
	)
	for {
		select {
		case pushJob = <-p.DistributionTask:
			// 分发
			if pushJob.Type == 1 {
				GetRoomManage().PushAll(&WSMessage{
					Code: pushJob.Code,
					Msg:  pushJob.PushType,
					Data: pushJob.Data,
				})
			} else {
				if _, ok := p.JobChan[pushJob.RoomId]; !ok {
					log.Print("发生堵塞")
					p.JobChan[pushJob.RoomId] = make(chan *PushJob)
					go p.pushWork(pushJob.RoomId)
				}
				p.JobChan[pushJob.RoomId] <- pushJob
			}
		}
	}
}

func (p *PushTask) pushWork(roomId string) {
	var (
		err error
		job *PushJob
	)
	for {
		select {
		case job = <-p.JobChan[roomId]:
			if err = GetRoomManage().PushRoom(roomId, &WSMessage{
				Code: job.Code,
				Msg:  job.PushType,
				Data: job.Data,
			}); err != nil {
				fmt.Println(err)
			}
		}
	}
}
