package raft_db

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type HeartBeatServer struct {
	db_state *LocalDataBase
	shared_state *SharedState
	deadline_tracker *DeadlineTracker
}

func NewHeartBeatServer(db_state *LocalDataBase, shared_state *SharedState, deadline_tracker *DeadlineTracker) *HeartBeatServer {
	return &HeartBeatServer{
		db_state: db_state,
		shared_state: shared_state,
		deadline_tracker: deadline_tracker,
	}
}

func (server *HeartBeatServer) HandleHeartBeat(w http.ResponseWriter, r *http.Request) {
	if server.shared_state.test_config.turn_on_server == 1 {
		fmt.Println("non processing HandleHeartBeat due blocking")
		http.Error(w, "Error while reading data", http.StatusInternalServerError)
		return
	}

	if server.shared_state.test_config.timeout_server == 1 {
		fmt.Println("Start timeouting")
		for {
			if server.shared_state.test_config.timeout_server == 0 {
				break;
			}
		}
	}


	fmt.Println("HandleHeartBeat")
	if server.deadline_tracker != nil {
		server.deadline_tracker.Warmup()
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("HandleHeartBeat: error bp1")
		http.Error(w, "Error while reading data", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	var message HeartBeatMessage
	err = json.Unmarshal(data, &message)
	if err != nil {
		fmt.Println("HandleHeartBeat: error bp2")
		http.Error(w, "Error while unmarshal data", http.StatusInternalServerError)
		return
	}

	server.shared_state.leader_url = message.Master_url
	if server.shared_state.term < message.Term {
		server.shared_state.term = message.Term
	}

	fmt.Println("Current length ", len(message.Changelog), len(server.db_state.Changelog))
	change_db := NewLocalDataBase()
	change_db.Changelog = message.Changelog
	server.db_state.AppendNewChanges(change_db)
	fmt.Println("Current length ", len(server.db_state.Changelog))

	term_data := map[string]interface{}{
		"term": server.shared_state.term,
		"state_len": server.db_state.Size(),
	}
	fmt.Println("Changelog len ", server.db_state.Size())

	jsonData, err := json.Marshal(term_data)
	if err != nil {
		fmt.Println("HandleHeartBeat: error bp3")
		http.Error(w, "Error while serialize response data", http.StatusInternalServerError)
		return
	}
	fmt.Println("HandleHeartBeat result data", len(jsonData), jsonData)

	w.Write(jsonData)
}

func (server *HeartBeatServer) UpdateTerm(term int) {
	server.shared_state.term = term
}
