import unittest
from raft_client import RaftClient, turn_off_server, turn_on_server, start_timeouting, end_timeouting
import time

class TestRaft(unittest.TestCase):
    def test_simple(self):
        client = RaftClient("http://localhost:8080")
        client.create("key1", "value1")
        self.assertEqual(client.read("key1"), "value1")
        client.update("key1", "value2")
        self.assertEqual(client.read("key1"), "value2")
        client.delete("key1")
        try:
            client.read("key1")
            self.assertEqual(True, False)
        except:
            pass

    def test_replication(self):
        client = RaftClient("http://localhost:8080")
        replica_client = RaftClient("http://localhost:8081")

        client.create("rkey1", "value1")
        try:
            replica_client.read("rkey1")
            self.assertEqual(True, False)
        except:
            pass

        time.sleep(2)
        self.assertEqual(replica_client.read("rkey1"), "value1")    

    def test_master_address(self):
        client = RaftClient("http://localhost:8080")
        replica_client = RaftClient("http://localhost:8081")

        time.sleep(2)

        self.assertEqual(client.get_master(), "localhost:8080")  
        self.assertEqual(replica_client.get_master(), "localhost:8080")  


    #### NOTE: this test is not match with timeout test
    def test_turn_off(self):
        client = RaftClient("http://localhost:8080")
        replica1_client = RaftClient("http://localhost:8081")
        replica2_client = RaftClient("http://localhost:8082")
        replica3_client = RaftClient("http://localhost:8082")

        client.create("k1", "value1")
        time.sleep(2)
        turn_off_server("http://localhost:8080")
        print("Server is turn off")
        time.sleep(15)

        self.assertEqual(replica1_client.get_master(), "localhost:8081")  
        self.assertEqual(replica2_client.get_master(), "localhost:8081")  
        self.assertEqual(replica3_client.get_master(), "localhost:8081")  
        
        self.assertEqual(replica1_client.read("k1"), "value1")  
        self.assertEqual(replica2_client.read("k1"), "value1")  
        self.assertEqual(replica3_client.read("k1"), "value1")

        replica1_client.create("k2", "value2")
        self.assertEqual(replica1_client.read("k2"), "value2")  
        try:
            replica2_client.read("k2")
            self.assertEqual(True, False)
        except:
            pass

        try:
            replica3_client.read("k2")
            self.assertEqual(True, False)
        except:
            pass

        time.sleep(2)

        self.assertEqual(replica2_client.read("k2"), "value2")  
        self.assertEqual(replica3_client.read("k2"), "value2")

        ###### RECOVERY #######

        turn_on_server("http://localhost:8080")

        time.sleep(2)
        self.assertEqual(client.read("k1"), "value1")  
        self.assertEqual(client.read("k2"), "value2")

    #### NOTE: this test is not match with turn off test. Run them separately
    #def test_timeouting(self):
    #    client = RaftClient("http://localhost:8080")
    #    replica1_client = RaftClient("http://localhost:8081")
    #    replica2_client = RaftClient("http://localhost:8082")
    #    replica3_client = RaftClient("http://localhost:8082")

    #    client.create("tkey1", "value1")
    #    time.sleep(1)

    #    start_timeouting("http://localhost:8080")

    #    time.sleep(15)

    #    self.assertEqual(replica1_client.get_master(), "localhost:8081")  
    #    self.assertEqual(replica2_client.get_master(), "localhost:8081")  
    #    self.assertEqual(replica3_client.get_master(), "localhost:8081")  
        
    #    self.assertEqual(replica1_client.read("tkey1"), "value1")  
    #    self.assertEqual(replica2_client.read("tkey1"), "value1")  
    #    self.assertEqual(replica3_client.read("tkey1"), "value1")

    #    replica1_client.create("tkey2", "value2")
    #    self.assertEqual(replica1_client.read("tkey2"), "value2")  
    #    try:
    #        replica2_client.read("tkey2")
    #        self.assertEqual(True, False)
    #    except:
    #        pass

    #    try:
    #        replica3_client.read("tkey2")
    #        self.assertEqual(True, False)
    #    except:
    #        pass

    #    time.sleep(2)
    #    replica1_client.create("tkey2", "value2")

    #    self.assertEqual(replica2_client.read("tkey2"), "value2")  
    #    self.assertEqual(replica3_client.read("tkey2"), "value2")  
    #    end_timeouting("http://localhost:8080")

if __name__ == "__main__":
    unittest.main()
