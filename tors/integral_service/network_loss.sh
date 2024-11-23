#!/bin/bash

docker-compose up -d
docker-compose exec server2 iptables -A INPUT -m statistic --mode random --probability 1 -j DROP
sleep 50
docker-compose logs
docker-compose down