import requests
import json

class RaftClient:
    def __init__(self, leader_server):
        self.leader_server = leader_server

    def create(self, key, value):
        url = self.leader_server + "/create"
        data = {
            "key": key,
            "value": value
        }
        response = requests.get(url, json=data)
        if response.status_code == 200:
            return 
        else:
            raise Exception("Not success create query")

    def update(self, key, value):
        url = self.leader_server + "/update"
        data = {
            "key": key,
            "value": value
        }
        response = requests.get(url, json=data)
        if response.status_code == 200:
            return 
        else:
            raise Exception("Not success create query")

    def delete(self, key):
        url = self.leader_server + "/delete"
        data = {
            "key": key,
        }
        response = requests.get(url, json=data)
        if response.status_code == 200:
            return 
        else:
            raise Exception("Not success create query")

    def read(self, key):
        url = self.leader_server + "/read"
        data = {
            "key": key,
        }
        response = requests.get(url, json=data)
        if response.status_code == 200:
            return response.json()["value"]
        else:
            raise Exception("Not success create query")

    def get_master(self):
        url = self.leader_server + "/master"
        response = requests.get(url)
        if response.status_code == 200:
            return response.json()["master"]
        else:
            raise Exception("Not success master query")

def turn_off_server(url):
    rsp = requests.get(url + "/turn_off")
    if rsp.status_code != 200:
        raise Exception("Not success turn_off_server")

def turn_on_server(url):
    rsp = requests.get(url + "/turn_on")
    if rsp.status_code != 200:
        raise Exception("Not success turn_on_server")

def start_timeouting(url):
    requests.get(url + "/timeout_off")

def end_timeouting(url):
    requests.get(url + "/timeout_on")


