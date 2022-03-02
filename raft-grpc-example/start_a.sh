#!/bin/bash


echo `date` ./raft-grpc-example --raft_id=nodeA --address=localhost:50051 --raft_data_dir /tmp/my-raft-cluster --log_file ./logs/node_a.log &
./raft-grpc-example --raft_id=nodeA --address=localhost:50051 --raft_data_dir /tmp/my-raft-cluster --log_file ./logs/node_a.log &

