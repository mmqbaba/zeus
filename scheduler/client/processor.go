package bus

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

const jobSnapshotPrefix = "/forest/client/snapshot/%s/%s/"
const jobExecuteSnapshotPrefix = "/forest/client/execute/snapshot/%s/%s/"

const (
	// 执行中
	JobExecuteDoingStatus = 1
	// 执行成功
	JobExecuteSuccessStatus = 2
	// 未知
	JobExecuteUkonwStatus = 3
	// 执行失败
	JobExecuteErrorStatus = -1
)

type JobSnapshotProcessor struct {
	etcd                *Etcd
	snapshotPath        string
	snapshotExecutePath string
	snapshots           chan *JobSnapshot
	jobs                map[string]Job
	lk                  *sync.RWMutex
}

// new a job snapshot processor
func NewJobSnapshotProcessor(group, ip string, etcd *Etcd) (*JobSnapshotProcessor) {

	processor := &JobSnapshotProcessor{
		etcd:      etcd,
		snapshots: make(chan *JobSnapshot, 100),
		jobs:      make(map[string]Job),
		lk:        &sync.RWMutex{},
	}
	processor.snapshotPath = fmt.Sprintf(jobSnapshotPrefix, group, ip)
	processor.snapshotExecutePath = fmt.Sprintf(jobExecuteSnapshotPrefix, group, ip)

	go processor.lookup()

	return processor
}

// lookup the job snapshot
func (processor *JobSnapshotProcessor) lookup() {

	for {

		select {
		case snapshot := <-processor.snapshots:

			go processor.handleSnapshot(snapshot)

		}
	}
}

func (processor *JobSnapshotProcessor) pushJobSnapshot(snapshot *JobSnapshot) {

	processor.snapshots <- snapshot
}

// handle the snapshot
func (processor *JobSnapshotProcessor) handleSnapshot(snapshot *JobSnapshot) {

	target := snapshot.Target

	now := time.Now()

	executeSnapshot := &JobExecuteSnapshot{
		Id:         snapshot.Id,
		JobId:      snapshot.JobId,
		Name:       snapshot.Name,
		Group:      snapshot.Group,
		Ip:         snapshot.Ip,
		Cron:       snapshot.Cron,
		Target:     snapshot.Target,
		Params:     snapshot.Params,
		Status:     JobExecuteDoingStatus,
		CreateTime: snapshot.CreateTime,
		Remark:     snapshot.Remark,
		Mobile:     snapshot.Mobile,
		StartTime:  now.Format("2006-01-02 15:04:05"),
		Times:      0,
	}

	if target == "" {

		log.Printf("the snapshot:%v target is nil", snapshot)
		executeSnapshot.Status = JobExecuteUkonwStatus
		return
	}

	job, ok := processor.jobs[target]
	if !ok || job == nil {
		log.Printf("the snapshot:%v target is not found in the job list", snapshot)
		executeSnapshot.Status = JobExecuteUkonwStatus
	}

	value, _ := json.Marshal(executeSnapshot)

	key := processor.snapshotExecutePath + executeSnapshot.Id
	if err := processor.etcd.Put(key, string(value)); err != nil {
		log.Printf("the snapshot:%v put snapshot execute fail:%v", executeSnapshot, err)
		return
	}

	if executeSnapshot.Status !=JobExecuteDoingStatus {
		return
	}

	result, err := job.Execute(snapshot.Params)
	after := time.Now()
	executeSnapshot.Status = JobExecuteSuccessStatus
	executeSnapshot.Result = result
	if err != nil {
		executeSnapshot.Status = JobExecuteErrorStatus
	}

	duration := after.Sub(now)

	times := duration / time.Second
	executeSnapshot.Times = int(times)
	executeSnapshot.FinishTime = after.Format("2006-01-02 15:04:05")
	log.Printf("the execute snapshot:%v execute success ", executeSnapshot)

	// store the execute job snapshot
	value, _ = json.Marshal(executeSnapshot)
	if err := processor.etcd.Put(key, string(value)); err != nil {
		log.Printf("the snapshot:%v put update snapshot execute fail:%v", executeSnapshot, err)
		return
	}

}

// push a job to job list
func (processor *JobSnapshotProcessor) PushJob(name string, job Job) {

	processor.lk.Lock()
	defer processor.lk.Unlock()

	if _, ok := processor.jobs[name]; ok {
		log.Printf("the job %s :%v  has exist!", name, job)
		return
	}

	processor.jobs[name] = job
}
