package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"strconv"
	"path/filepath"

	pb "github.com/Jille/raft-grpc-example/proto"
	"github.com/Jille/raft-grpc-leader-rpc/leaderhealth"
	transport "github.com/Jille/raft-grpc-transport"
	"github.com/Jille/raftadmin"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

)

var (
	myAddr = flag.String("address", "localhost:50051", "TCP host+port for this node")
	raftId = flag.String("raft_id", "", "Node id used by Raft")

	raftDir       = flag.String("raft_data_dir", "data/", "Raft data dir")
	raftBootstrap = flag.Bool("raft_bootstrap", false, "Whether to bootstrap the Raft cluster")

	logFile = flag.String("log_file", "./log", "log file")
)


var g_localaddr = ""


func TestWriteApply(svr *rpcInterface) {
	req := &pb.AddWordRequest{}

	index := 0
	for {

		req.Word = "hello_"
		index++

		req.Word = req.Word + strconv.Itoa(index)

		time.Sleep(time.Duration(5)*time.Second)

		LastIndex := svr.raft.LastIndex()
		fmt.Println("LastIndex = ", LastIndex)

		leaderAddr := string(svr.raft.Leader())
		localAddr := g_localaddr
		if leaderAddr != localAddr {
			fmt.Println("---------------------------------")
			continue
		}

		_,err := svr.AddWord(context.Background(), req)
		if nil != err {
			fmt.Println("AddWord fail, err = ", err)
		}

		fmt.Println("---------------------------------")
	}
}

func main() {

	flag.Parse()

	{
		f, err := os.OpenFile(*logFile, os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0755)
		if nil != err {
			fmt.Println("ERROR: os.OpenFile fail, ", logFile)
			return
		}

		os.Stdout = f
		os.Stderr = f
		fmt.Println("user log file = ", *logFile)
		log.SetOutput(f)
	}

	if *raftId == "" {
		log.Fatalf("flag --raft_id is required")
	}

	ctx := context.Background()
	_, port, err := net.SplitHostPort(*myAddr)
	if err != nil {
		log.Fatalf("failed to parse local address (%q): %v", *myAddr, err)
	}
	sock, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	wt := &wordTracker{}

	r, tm, err := NewRaft(ctx, *raftId, *myAddr, wt)
	if err != nil {
		log.Fatalf("failed to start raft: %v", err)
	}
	s := grpc.NewServer()
	svr := &rpcInterface{
		wordTracker: wt,
		raft:        r,
	}

	go TestWriteApply(svr)

	pb.RegisterExampleServer(s, svr)
	tm.Register(s)
	leaderhealth.Setup(r, s, []string{"Example"})
	raftadmin.Register(s, r)
	reflection.Register(s)

	if err := s.Serve(sock); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func NewRaft(ctx context.Context, myID, myAddress string, fsm raft.FSM) (*raft.Raft, *transport.Manager, error) {
	c := raft.DefaultConfig()
	c.LocalID = raft.ServerID(myID)

	baseDir := filepath.Join(*raftDir, myID)

	ldb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "logs.dat"))
	if err != nil {
		return nil, nil, fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "logs.dat"), err)
	}

	sdb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "stable.dat"))
	if err != nil {
		return nil, nil, fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "stable.dat"), err)
	}

	fss, err := raft.NewFileSnapshotStore(baseDir, 3, os.Stderr)
	if err != nil {
		return nil, nil, fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, baseDir, err)
	}

	tm := transport.New(raft.ServerAddress(myAddress), []grpc.DialOption{grpc.WithInsecure()})

	r, err := raft.NewRaft(c, fsm, ldb, sdb, fss, tm.Transport())
	if err != nil {
		return nil, nil, fmt.Errorf("raft.NewRaft: %v", err)
	}

	if *raftBootstrap {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(myID),
					Address:  raft.ServerAddress(myAddress),
				},
			},
		}
		f := r.BootstrapCluster(cfg)
		if err := f.Error(); err != nil {
			return nil, nil, fmt.Errorf("raft.Raft.BootstrapCluster: %v", err)
		}
	}

	g_localaddr = myAddress
	return r, tm, nil
}
