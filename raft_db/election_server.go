package raft_db

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type ElectionServer struct {
	shared_state *SharedState
}

func NewElectionServer(shared_state *SharedState) *ElectionServer {
	return &ElectionServer{
		shared_state: shared_state,
	}
}

func (server *ElectionServer) HandleElection(w http.ResponseWriter, r *http.Request) {
	if server.shared_state.test_config.turn_on_server == 1 {
		fmt.Println("non processing handleElection due blocking")
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
		return
	}


	fmt.Println("handleElection")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("handleElection bp1")
		http.Error(w, "Error while reading data", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var parsedBody map[string]interface{}
	if err := json.Unmarshal(data, &parsedBody); err != nil {
		fmt.Println("handleElection bp2")
		http.Error(w, "Error while reading data", http.StatusInternalServerError)
		return
	}

	termInt := 0
	if termValue, ok := parsedBody["term"].(float64); ok { 
		termInt = int(termValue)
	} else {
		fmt.Println("handleElection bp3")
		http.Error(w, "Error while reading data", http.StatusInternalServerError)
		return
	}

	term_data := map[string]interface{}{}

	if termInt < server.shared_state.term {
		term_data["late"] = server.shared_state.term
	}
	if !server.shared_state.voted[server.shared_state.term] {
		term_data["voted"] = 1
	} else {
		server.shared_state.voted[server.shared_state.term] = true
		term_data["voted"] = 0
	}

	jsonData, err := json.Marshal(term_data)
	if err != nil {
		fmt.Println("handleElection bp4")
		http.Error(w, "Error while serialize response data", http.StatusInternalServerError)
		return
	}

	fmt.Println("Write response on election server ", len(jsonData), jsonData)
	w.Write(jsonData)
}
