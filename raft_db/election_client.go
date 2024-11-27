package raft_db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ElectionClient struct {
	config *Config
	heartbeat_client *HeartBeatClient
	shared_state *SharedState
}

func NewElectionClient(config *Config, heartbeat_client *HeartBeatClient) *ElectionClient {
	return &ElectionClient{
		config: config,
		heartbeat_client: heartbeat_client,
		shared_state: heartbeat_client.shared_state,
	}
}

func (client* ElectionClient) Run() {
	if client.shared_state.test_config.turn_on_server == 1 {
		fmt.Println("non processing Run due blocking")
		return
	}
	if client.shared_state.test_config.timeout_server == 1 {
		fmt.Println("Start timeouting")
		for {
			if client.shared_state.test_config.timeout_server == 0 {
				break;
			}
		}
	}

	fmt.Println("Start election")
	results := make(chan int, len(client.config.other_node_urls))	

	for _, server := range client.config.other_node_urls {
		go func(server string, ch chan<- int) {
			url := server + "/election"
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(client.config.ttls_milliseconds) * time.Millisecond)
			defer cancel()

			term_data := map[string]interface{}{
				"term": client.shared_state.term,
			}
		
			jsonData, err := json.Marshal(term_data)

			req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Println("bp1")
				ch <- -1
				return
			}
		
			http_client := &http.Client{}
			resp, err := http_client.Do(req)
			if err != nil {
				fmt.Println("bp2")
				ch <- -1
				return
			}
			defer resp.Body.Close()
		
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("bp3")
				ch <- -1
				return
			}
		
			var parsedBody map[string]interface{}
			if err := json.Unmarshal(body, &parsedBody); err != nil {
				fmt.Println("bp4", len(body), body)
				ch <- -1
				return
			}
			
			if termValue, ok := parsedBody["late"].(float64); ok { 
				termInt := int(termValue)
				ch <- termInt
				ch <- -2
				return
			}

			if termValue, ok := parsedBody["voted"].(float64); ok { 
				termInt := int(termValue)
				ch <- termInt
				return
			}
			fmt.Println("bp5")
			ch <- -1
		}(server, results)
	}
	term_values := make([]int, 0)
	prev_value := 0
	for i := 0; i < len(client.config.other_node_urls); i++ {
		value := <-results
		if value == -2 {
			client.heartbeat_client.shared_state.term = prev_value
			return
		}
		if value == -1 {
			fmt.Println("Follower ", client.config.other_node_urls[i], " become unavailiable (election client)")
		}
		term_values = append(term_values, value)
		prev_value = value
	}

	cnt_votes := 0

	if !client.shared_state.voted[client.shared_state.term] {
		client.shared_state.voted[client.shared_state.term] = true
		cnt_votes += 1
	}
	for _, value := range term_values {
		if value == 1 {
			cnt_votes += 1
		}
	}

	fmt.Println("Count votes for me ", cnt_votes)
	if cnt_votes > len(client.config.other_node_urls) / 2 {
		fmt.Println("Become leader at term ", client.shared_state.term + 1)
		client.shared_state.leader_url = client.config.url
		client.heartbeat_client.UpdateTerm(client.shared_state.term + 1)
		go client.heartbeat_client.Run()
	}
}
