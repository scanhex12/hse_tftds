#!/bin/bash

go build -o bootstrap

./bootstrap --is_leader=false --url="localhost:8081" --other_nodes_url="http://localhost:8080,http://localhost:8082,http://localhost:8083" --ttls_milliseconds=500 --period_milliseconds=1000 --max_master_silence_ms=14000
./bootstrap --is_leader=false --url="localhost:8082" --other_nodes_url="http://localhost:8081,http://localhost:8080,http://localhost:8083" --ttls_milliseconds=500 --period_milliseconds=1000 --max_master_silence_ms=18000
./bootstrap --is_leader=false --url="localhost:8083" --other_nodes_url="http://localhost:8081,http://localhost:8082,http://localhost:8080" --ttls_milliseconds=500 --period_milliseconds=1000 --max_master_silence_ms=19000

./bootstrap --is_leader=true --url="localhost:8080" --other_nodes_url="http://localhost:8081,http://localhost:8082,http://localhost:8083" --ttls_milliseconds=500 --period_milliseconds=1500 --max_master_silence_ms=19000
