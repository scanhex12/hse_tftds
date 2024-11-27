package raft_db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"sync"
)

type ConcurrentMap struct {
	mu sync.RWMutex
	m  map[string]int
}

func NewConcurrentMap() *ConcurrentMap {
	return &ConcurrentMap{
		m: make(map[string]int),
	}
}

func (c *ConcurrentMap) Get(key string) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.m[key]
	return value, ok
}

func (c *ConcurrentMap) Set(key string, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
}


type HeartBeatClient struct {
	other_nodes []string
	ttls_milliseconds int
	period_milliseconds int
	shared_state *SharedState
	db_state *LocalDataBase
	current_changelog_length *ConcurrentMap
}

func NewHeartBeatClient(other_nodes []string, ttls_milliseconds, period_milliseconds int, shared_state *SharedState, db_state *LocalDataBase) *HeartBeatClient {
	return &HeartBeatClient{
		other_nodes: other_nodes,
		ttls_milliseconds: ttls_milliseconds,
		period_milliseconds: period_milliseconds,
		shared_state: shared_state,
		db_state: db_state,
		current_changelog_length: NewConcurrentMap(),
	}
}

// Return true if we are not in current term and should not be leader
func (client *HeartBeatClient) SendHeartBeats() bool {
	if client.shared_state.test_config.turn_on_server == 1 {
		fmt.Println("non processing SendHeartBeats due blocking")
		return false
	}

	if client.shared_state.test_config.timeout_server == 1 {
		fmt.Println("Start timeouting")
		for {
			if client.shared_state.test_config.timeout_server == 0 {
				break;
			}
		}
		return false
	}

	results := make(chan int, len(client.other_nodes))

	for _, server := range client.other_nodes {
		go func(server string, ch chan<- int) {
			cur_changelog_len, ok := client.current_changelog_length.Get(server)
			if !ok {
				cur_changelog_len = client.db_state.Size()
			}
			fmt.Println("Current server position ", cur_changelog_len, server)
			data, err := client.db_state.GetChangelogFromPosition(cur_changelog_len)
			if err != nil {
				panic(err)
			}
		
			message := HeartBeatMessage{
				Master_url: client.shared_state.leader_url,
				Changelog: data,
				Term: client.shared_state.term,
			}
			fmt.Println("Sending message", message, len(message.Changelog))
		
			serialized_message, err := json.Marshal(message)
			if err != nil {
				fmt.Println("bp0")
				ch <- -1
				return
			}

			url := server + "/heartbeat"
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(client.ttls_milliseconds) * time.Millisecond)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(serialized_message))
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
				fmt.Println("Follower ", server, " become unavailiable (heartbeat client)")
				fmt.Println("bp4 hb ", body, len(body), url)
				ch <- -1
				return
			}
		
			if termValue, ok := parsedBody["term"].(float64); ok { 
				termInt := int(termValue)
				ch <- termInt

				changelog_len, _ := parsedBody["state_len"].(float64)
				client.current_changelog_length.Set(server, int(changelog_len))
				fmt.Println("Set current server position ", changelog_len, server, client.db_state.Size(), client.shared_state.test_config.turn_on_server)

				return
			}
			fmt.Println("bp5")
			ch <- -1
		}(server, results)
	}

	term_values := make([]int, 0)
	for i := 0; i < len(client.other_nodes); i++ {
		value := <-results
		if value == -1 {
			continue
		}
		term_values = append(term_values, value)
	}
	current_term := GetMostCommonValue(term_values)
	fmt.Println("Terms of followers and master ", current_term, client.shared_state.term)
	return client.shared_state.term < current_term
}

func (client *HeartBeatClient) UpdateTerm(term int) {
	client.shared_state.term = term
}

type HeartBeatMessage struct {
	Master_url string 		`json:"master_url"`
	Changelog []*Change 	`json:"changelog"`
	Term int 				`json:"term"`
}

func (client *HeartBeatClient) RunBody() bool {
	res := client.SendHeartBeats()
	time.Sleep(time.Duration(client.period_milliseconds) * time.Millisecond)
	return res
}

func (client *HeartBeatClient) Run() {
	for {
		res := client.RunBody()
		if res {
			return
		}
	}
}

func (client *HeartBeatClient) IteratedRun(iterations int) {
	for i := 0; i < iterations; i++ {
		res := client.RunBody()
		if res {
			return
		}
	}
}
