package raft_db

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/stretchr/testify/assert"
)
func TestHeartbeatServerAndClient(t *testing.T) {
	db_client := NewLocalDataBase()

	db_client.Create("key1", "value1")
	db_client.Update("key1", "value2")
	db_client.Create("key3", "value3")

	db_server := NewLocalDataBase()
	
	handler := http.NewServeMux()
	shared_state_server := NewSharedState(42, "")

	heartbeat_server := NewHeartBeatServer(db_server, shared_state_server, nil)
	handler.HandleFunc("/heartbeat", heartbeat_server.HandleHeartBeat)

	server := httptest.NewServer(handler)
	defer server.Close()

	server_url := server.URL
	shared_state_client := NewSharedState(42, "")
	heartbeat_client := NewHeartBeatClient([]string{server_url}, 1000, 1000, shared_state_client, db_client)
	heartbeat_client.IteratedRun(1)

	value, err := db_server.Read("key1")
	assert.Equal(t, err, nil)
	assert.Equal(t, value, "value2")	
}

func TestMostCommonValue(t *testing.T) {
	assert.Equal(t, 3, GetMostCommonValue([]int{3}))
	assert.Equal(t, 3, GetMostCommonValue([]int{1,3,2,3}))
	assert.Equal(t, 1, GetMostCommonValue([]int{1,3,1,3,2,1}))
}
