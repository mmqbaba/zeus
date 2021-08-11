package bus

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
)

const clientPathPrefix = "/forest/client/%s/clients/%s"
const (
	URegistryState = iota
	RegistryState
)

type ForestClient struct {
	etcd         *Etcd
	jobs         map[string]Job
	group        string
	ip           string
	running      bool
	quit         chan bool
	state        int
	clientPath   string
	snapshotPath string
	txResponse   *TxResponse

	snapshotProcessor *JobSnapshotProcessor
}

// new a client
func NewForestClient(group, ip string, etcd *Etcd) (*ForestClient) {

	return &ForestClient{
		etcd:  etcd,
		group: group,
		ip:    ip,
		jobs:  make(map[string]Job, 0),
		quit:  make(chan bool, 0),
		state: URegistryState,
	}

}

// bootstrap client
func (client *ForestClient) Bootstrap() (err error) {

	if err = client.validate(); err != nil {
		return err
	}

	client.clientPath = fmt.Sprintf(clientPathPrefix, client.group, client.ip)
	client.snapshotPath = fmt.Sprintf(jobSnapshotPrefix, client.group, client.ip)

	client.snapshotProcessor = NewJobSnapshotProcessor(client.group, client.ip, client.etcd)

	client.addJobs()

	go client.registerNode()
	go client.lookup()

	client.running = true
	<-client.quit
	client.running = false
	return
}

// stop  client
func (client *ForestClient) Stop() {
	if client.running {
		client.quit <- true
	}

	return
}

// add jobs
func (client *ForestClient) addJobs() {

	if len(client.jobs) == 0 {
		return
	}

	for name, job := range client.jobs {

		client.snapshotProcessor.PushJob(name, job)
	}
}

// pre validate params
func (client *ForestClient) validate() (err error) {

	if client.ip == "" {
		return errors.New(fmt.Sprint("ip not allow null"))
	}

	if client.group == "" {
		return errors.New(fmt.Sprint("group not allow null"))
	}
	return
}

// push a new job to job list
func (client ForestClient) PushJob(name string, job Job) (err error) {

	if client.running {
		return errors.New(fmt.Sprintf("the forest client is running not allow push a job "))
	}
	if _, ok := client.jobs[name]; ok {
		return errors.New(fmt.Sprintf("the job %s name exist!", name))
	}
	client.jobs[name] = job

	return
}

func (client *ForestClient) registerNode() {

RETRY:
	var (
		txResponse *TxResponse
		err        error
	)

	if client.state == RegistryState {
		log.Printf("the forest client has already registry to:%s", client.clientPath)
		return
	}
	if txResponse, err = client.etcd.TxKeepaliveWithTTL(client.clientPath, client.ip, 10); err != nil {
		log.Printf("the forest client fail registry to:%s", client.clientPath)
		time.Sleep(time.Second * 3)
		goto RETRY
	}

	if !txResponse.Success {
		log.Printf("the forest client fail registry to:%s", client.clientPath)
		time.Sleep(time.Second * 3)
		goto RETRY
	}

	log.Printf("the forest client success registry to:%s", client.clientPath)
	client.state = RegistryState
	client.txResponse = txResponse

	select {
	case <-client.txResponse.StateChan:
		client.state = URegistryState
		log.Printf("the forest client fail registry to----->:%s", client.clientPath)
		goto RETRY
	}

}

// look up
func (client *ForestClient) lookup() {

	for {

		keys, values, err := client.etcd.GetWithPrefixKeyLimit(client.snapshotPath, 50)

		if err != nil {

			log.Printf("the forest client load job snapshot error:%v", err)
			time.Sleep(time.Second * 3)
			continue
		}

		if client.state == URegistryState {
			time.Sleep(time.Second * 3)
			continue
		}

		if len(keys) == 0 || len(values) == 0 {
			log.Printf("the forest client :%s load job snapshot is empty", client.clientPath)
			time.Sleep(time.Second * 3)
			continue
		}

		for i := 0; i < len(values); i++ {

			key := keys[i]

			if client.state == URegistryState {
				time.Sleep(time.Second * 3)
				break
			}

			if err := client.etcd.Delete(string(key)); err != nil {
				log.Printf("the forest client :%s delete job snapshot fail:%v", client.clientPath, err)
				continue
			}

			value := values[i]
			if len(value) == 0 {
				log.Printf("the forest client :%s found job snapshot value is nil ", client.clientPath)
				continue
			}

			snapshot := new(JobSnapshot)
			err := json.Unmarshal(value, snapshot)
			if err != nil {

				log.Printf("the forest client :%s found job snapshot value is cant not parse the json value：%v ", client.clientPath, err)
				continue
			}

			// push a job snapshot
			client.snapshotProcessor.pushJobSnapshot(snapshot)

		}

	}
}
