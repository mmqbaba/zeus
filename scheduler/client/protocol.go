package bus

import (
	"context"
	"github.com/coreos/etcd/clientv3"
)

const (
	KeyCreateChangeEvent = iota
	KeyUpdateChangeEvent
	KeyDeleteChangeEvent
)

// key 变化事件
type KeyChangeEvent struct {
	Type  int
	Key   string
	Value []byte
}

// 监听key 变化响应
type WatchKeyChangeResponse struct {
	Event      chan *KeyChangeEvent
	CancelFunc context.CancelFunc
	Watcher    clientv3.Watcher
}

type TxResponse struct {
	Success   bool
	LeaseID   clientv3.LeaseID
	Lease     clientv3.Lease
	Key       string
	Value     string
	StateChan chan bool
}


type JobSnapshot struct {
	Id         string `json:"id"`
	JobId      string `json:"jobId"`
	Name       string `json:"name"`
	Ip         string `json:"ip"`
	Group      string `json:"group"`
	Cron       string `json:"cron"`
	Target     string `json:"target"`
	Params     string `json:"params"`
	Mobile     string `json:"mobile"`
	Remark     string `json:"remark"`
	CreateTime string `json:"createTime"`
}


type JobExecuteSnapshot struct {
	Id         string `json:"id",xorm:"pk"`
	JobId      string `json:"jobId",xorm:"job_id"`
	Name       string `json:"name",xorm:"name"`
	Ip         string `json:"ip",xorm:"ip"`
	Group      string `json:"group",xorm:"group"`
	Cron       string `json:"cron",xorm:"cron"`
	Target     string `json:"target",xorm:"target"`
	Params     string `json:"params",xorm:"params"`
	Mobile     string `json:"mobile",xorm:"mobile"`
	Remark     string `json:"remark",xorm:"remark"`
	CreateTime string `json:"createTime",xorm:"create_time"`
	StartTime  string `json:"startTime",xorm:"start_time"`
	FinishTime string `json:"finishTime",xorm:"finish_time"`
	Times      int    `json:"times",xorm:"times"`
	Status     int    `json:"status",xorm:"status"`
	Result     string `json:"result",xorm:"result"`
}