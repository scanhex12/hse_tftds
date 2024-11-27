package raft_db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
)

type TestConfig struct {
	turn_on_server int32
	timeout_server int32
}

type SharedState struct {
	term int
	leader_url string
	voted map[int]bool
	test_config *TestConfig
}

func NewSharedState(term int, master_url string) *SharedState {
	return &SharedState{
		term: term,
		leader_url: master_url,
		voted: make(map[int]bool),
		test_config: &TestConfig{turn_on_server: 0, timeout_server: 0},
	}
}

type Bootstrap struct {
	db_state 			*LocalDataBase
	Heartbeat_server 	*HeartBeatServer
	heartbeat_client 	*HeartBeatClient
	config   			*Config
	shared_state 		*SharedState
	deadline_tracker    *DeadlineTracker
	election_client     *ElectionClient
	Election_server     *ElectionServer
}

func NewBootstrap(config *Config) *Bootstrap {
	db_state := NewLocalDataBase()
	leader_url := ""
	if config.is_leader {
		leader_url = config.url
	}
	shared_state := NewSharedState(0,leader_url)
	heartbeat_client := NewHeartBeatClient(config.other_node_urls, config.ttls_milliseconds, config.period_milliseconds, shared_state, db_state)
	election_client := NewElectionClient(config, heartbeat_client)
	deadline_tracker := NewDeadlineTracker(election_client, config.max_master_silence_ms)
	return &Bootstrap{
		db_state: db_state,
		Heartbeat_server: NewHeartBeatServer(db_state, shared_state, deadline_tracker),
		heartbeat_client: heartbeat_client,
		config: config,
		shared_state: shared_state,
		deadline_tracker: deadline_tracker,
		election_client: election_client,
		Election_server: NewElectionServer(shared_state),
	}
}

func (bootstrap *Bootstrap) HandleCreate(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading data", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var parsedBody map[string]interface{}
	if err := json.Unmarshal(data, &parsedBody); err != nil {
		http.Error(w, "Error while unmarshal", http.StatusInternalServerError)
		return
	}

	key, ok := parsedBody["key"].(string)
	if !ok { 
		http.Error(w, "Error while extracting key", http.StatusInternalServerError)
		return
	}
	value, ok := parsedBody["value"].(string)
	if !ok { 
		http.Error(w, "Error while extracting value", http.StatusInternalServerError)
		return
	}
	fmt.Println("Create ", key, value)
	bootstrap.db_state.Create(key, value)
}

func (bootstrap *Bootstrap) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading data", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var parsedBody map[string]interface{}
	if err := json.Unmarshal(data, &parsedBody); err != nil {
		http.Error(w, "Error while unmarshal", http.StatusInternalServerError)
		return
	}

	key, ok := parsedBody["key"].(string)
	if !ok { 
		http.Error(w, "Error while extracting key", http.StatusInternalServerError)
		return
	}
	value, ok := parsedBody["value"].(string)
	if !ok { 
		http.Error(w, "Error while extracting value", http.StatusInternalServerError)
		return
	}
	bootstrap.db_state.Update(key, value)
}

func (bootstrap *Bootstrap) HandleDelete(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading data", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var parsedBody map[string]interface{}
	if err := json.Unmarshal(data, &parsedBody); err != nil {
		http.Error(w, "Error while unmarshal", http.StatusInternalServerError)
		return
	}

	key, ok := parsedBody["key"].(string)
	if !ok { 
		http.Error(w, "Error while extracting key", http.StatusInternalServerError)
		return
	}
	bootstrap.db_state.Delete(key)
}

func (bootstrap *Bootstrap) HandleRead(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handleRead")
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading data", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var parsedBody map[string]interface{}
	if err := json.Unmarshal(data, &parsedBody); err != nil {
		http.Error(w, "Error while unmarshal", http.StatusInternalServerError)
		return
	}

	key, ok := parsedBody["key"].(string)
	if !ok { 
		http.Error(w, "Error while extracting key", http.StatusInternalServerError)
		return
	}

	value, err := bootstrap.db_state.Read(key)
	if err != nil {
		http.Error(w, "Error while extracting key", http.StatusNotFound)
		return
	}
	fmt.Println("Read ", key, value)

	term_data := map[string]interface{}{
		"value": value,
	}

	jsonData, err := json.Marshal(term_data)
	if err != nil {
		http.Error(w, "Error while serialize response data", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

func (bootstrap *Bootstrap) HandleMaster(w http.ResponseWriter, r *http.Request) {
	term_data := map[string]interface{}{
		"master": bootstrap.shared_state.leader_url,
	}

	jsonData, err := json.Marshal(term_data)
	if err != nil {
		http.Error(w, "Error while serialize response data", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

func (bootstrap *Bootstrap) HandleTurnOff(w http.ResponseWriter, r *http.Request) {
	atomic.StoreInt32(&bootstrap.shared_state.test_config.turn_on_server, 1)
}

func (bootstrap *Bootstrap) HandleTurnOn(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Turn on server")
	atomic.StoreInt32(&bootstrap.shared_state.test_config.turn_on_server, 0)
}

func (bootstrap *Bootstrap) HandleTimeoutOff(w http.ResponseWriter, r *http.Request) {
	atomic.StoreInt32(&bootstrap.shared_state.test_config.timeout_server, 1)
}

func (bootstrap *Bootstrap) HandleTimeoutOn(w http.ResponseWriter, r *http.Request) {
	atomic.StoreInt32(&bootstrap.shared_state.test_config.timeout_server, 0)
}


func (bootstrap *Bootstrap) Run() {
	if (bootstrap.config.is_leader) {
		bootstrap.heartbeat_client.Run()
	} else {
		bootstrap.deadline_tracker.Warmup()
	}
}
