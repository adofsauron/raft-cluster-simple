#!/bin/bash

rm raft-grpc-example -f

go build -gcflags=all="-N -l" -o raft-grpc-example .


