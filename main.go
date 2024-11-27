package main

import (
	"flag"
	"strings"
	"fmt"
	"net/http"
	"raftik/raft_db"
)

func main() {
	var is_leader bool
	var url string
	var combined_other_nodes string
	var ttls_milliseconds int
	var period_milliseconds int
	var max_master_silence_ms int

	flag.BoolVar(&is_leader, "is_leader", false, "Is first leader or not")
	flag.StringVar(&url, "url", "", "url of this server")
	flag.StringVar(&combined_other_nodes, "other_nodes_url", "", "other nodes urls")
	flag.IntVar(&ttls_milliseconds, "ttls_milliseconds", 0, "ttls_milliseconds")
	flag.IntVar(&period_milliseconds, "period_milliseconds", 0, "period_milliseconds")
	flag.IntVar(&max_master_silence_ms, "max_master_silence_ms", 0, "max_master_silence_ms")

	flag.Parse()

	other_node_urls := strings.Split(combined_other_nodes, ",")
	config := raft_db.NewConfig(is_leader, url, other_node_urls, ttls_milliseconds, period_milliseconds, max_master_silence_ms)
	bootstrap := raft_db.NewBootstrap(config)

	fmt.Println("Start running bootstrap on ", url)

	go bootstrap.Run()

	http.HandleFunc("/read", bootstrap.HandleRead)
	http.HandleFunc("/update", bootstrap.HandleUpdate)
	http.HandleFunc("/delete", bootstrap.HandleDelete)
	http.HandleFunc("/create", bootstrap.HandleCreate)
	http.HandleFunc("/heartbeat", bootstrap.Heartbeat_server.HandleHeartBeat)
	http.HandleFunc("/master", bootstrap.HandleMaster)
	http.HandleFunc("/turn_on", bootstrap.HandleTurnOn)
	http.HandleFunc("/turn_off", bootstrap.HandleTurnOff)
	http.HandleFunc("/timeout_on", bootstrap.HandleTimeoutOn)
	http.HandleFunc("/timeout_off", bootstrap.HandleTimeoutOff)

	http.HandleFunc("/election", bootstrap.Election_server.HandleElection)

	fmt.Println("Serving server ", url)

	err := http.ListenAndServe(url, nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}
