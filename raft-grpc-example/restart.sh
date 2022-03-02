#!/bin/bash


rm ./raft-data -rf
mkdir ./raft-data
mkdir ./raft-data/nodeA
mkdir ./raft-data/nodeB
mkdir ./raft-data/nodeC

rm ./logs/* -rf

pkill -9 raft

sleep 2s

./raft-grpc-example --raft_bootstrap --raft_id=nodeA --address=localhost:50051 --raft_data_dir ./raft-data --log_file ./logs/node_a.log &

./raft-grpc-example --raft_bootstrap --raft_id=nodeB --address=localhost:50052 --raft_data_dir ./raft-data --log_file ./logs/node_b.log &

./raft-grpc-example --raft_bootstrap --raft_id=nodeC --address=localhost:50053 --raft_data_dir ./raft-data --log_file ./logs/node_c.log &



# ./raft-grpc-example --raft_id=nodeB --address=localhost:50052 --raft_data_dir ./raft-data --log_file ./logs/node_b.log &
# ./raft-grpc-example  --raft_id=nodeC --address=localhost:50053 --raft_data_dir ./raft-data --log_file ./logs/node_c.log &


ps -ef | grep raft-grpc-example | grep -v grep

