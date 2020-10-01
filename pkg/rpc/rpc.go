package rpc

import (
	"github.com/tmyksj/rlso11n/app/logger"
	"github.com/tmyksj/rlso11n/pkg/context"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"
)

type Rpc int

var clientMap = make(map[string]*rpc.Client)
var clientMu sync.Mutex

var serverIsRunning = false
var serverMu sync.Mutex

func Call(host string, serviceMethod string, req interface{}, res interface{}) error {
	clientMu.Lock()

	client, ok := clientMap[host]
	if !ok {
		c, err := rpc.DialHTTP("tcp", host+":"+strconv.Itoa(context.RpcPort()))
		if err != nil {
			logger.Error(pkg, "failed to connect to %v, %v", host, err)

			clientMu.Unlock()
			return err
		}

		client = c
		clientMap[host] = client

		logger.Info(pkg, "succeed to connect to %v", host)
	}

	clientMu.Unlock()

	err := client.Call(serviceMethod, req, res)
	if err != nil {
		logger.Error(pkg, "failed to call %v, %v", serviceMethod, err)
		return err
	}

	logger.Info(pkg, "succeed to call %v", serviceMethod)

	return nil
}

func Serve() error {
	serverMu.Lock()
	defer serverMu.Unlock()

	if serverIsRunning {
		return nil
	}

	s := new(Rpc)
	if err := rpc.Register(s); err != nil {
		logger.Error(pkg, "failed to register rpc, %v", err)
		return err
	}

	rpc.HandleHTTP()

	l, err := net.Listen("tcp", ":"+strconv.Itoa(context.RpcPort()))
	if err != nil {
		logger.Error(pkg, "failed to listen %v/tcp, %v", context.RpcPort(), err)
		return err
	}

	go func() {
		err := http.Serve(l, nil)
		logger.Info(pkg, "served, %v", err)
	}()

	serverIsRunning = true
	logger.Info(pkg, "succeed to start server")

	return nil
}
