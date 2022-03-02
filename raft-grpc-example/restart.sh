#!/bin/bash


rm /tmp/my-raft-cluster -rf
mkdir /tmp/my-raft-cluster
mkdir /tmp/my-raft-cluster/nodeA
mkdir /tmp/my-raft-cluster/nodeB
mkdir /tmp/my-raft-cluster/nodeC

rm ./logs/* -rf

pkill -9 raft

sleep 2s

./raft-grpc-example --raft_bootstrap --raft_id=nodeA --address=localhost:50051 --raft_data_dir /tmp/my-raft-cluster --log_file ./logs/node_a.log &

./raft-grpc-example --raft_id=nodeB --address=localhost:50052 --raft_data_dir /tmp/my-raft-cluster --log_file ./logs/node_b.log &

./raft-grpc-example  --raft_id=nodeC --address=localhost:50053 --raft_data_dir /tmp/my-raft-cluster --log_file ./logs/node_c.log &


ps -ef | grep raft-grpc-example | grep -v grep

