// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package llm_runner

import (
	"net"
	"net/rpc"
	"strings"
	"time"
)

const rpcDialTimeout = 3 * time.Second

func RpcExecuteRun(address, workDir, command string) (resp RpcRunResponse) {
	return RpcExecuteRunWithEnv(address, workDir, command, nil)
}

func RpcExecuteRunWithEnv(address, workDir, command string, env map[string]string) (resp RpcRunResponse) {
	return RpcExecuteRunWithIDAndEnv(address, workDir, ``, command, env)
}

func RpcExecuteRunWithID(address, workDir, runID, command string) (resp RpcRunResponse) {
	return RpcExecuteRunWithIDAndEnv(address, workDir, runID, command, nil)
}

func RpcExecuteRunWithIDAndEnv(address, workDir, runID, command string, env map[string]string) (resp RpcRunResponse) {
	client, err := dialRpcClient(address)
	if err != nil {
		resp.IsError = true
		resp.ErrorMsg = err.Error()
		return
	}
	defer func(client *rpc.Client) {
		_ = client.Close()
	}(client)
	err = client.Call(RpcServiceName+`.Run`, RpcRunRequest{Command: command, WorkDir: workDir, RunID: runID, Env: env}, &resp)
	if err != nil {
		resp.IsError = true
		resp.ErrorMsg = err.Error()
	}
	return
}

func RpcCancelRun(address, runID string) (resp RpcCancelResponse) {
	if strings.TrimSpace(runID) == `` {
		resp.ErrorMsg = `run id is required`
		return
	}
	client, err := dialRpcClient(address)
	if err != nil {
		resp.ErrorMsg = err.Error()
		return
	}
	defer func(client *rpc.Client) {
		_ = client.Close()
	}(client)
	err = client.Call(RpcServiceName+`.Cancel`, RpcCancelRequest{RunID: runID}, &resp)
	if err != nil {
		resp.ErrorMsg = err.Error()
	}
	return
}

func dialRpcClient(address string) (*rpc.Client, error) {
	if !strings.Contains(address, `:`) {
		address = address + RpcServicePort
	}
	conn, err := net.DialTimeout(`tcp`, address, rpcDialTimeout)
	if err != nil {
		return nil, err
	}
	return rpc.NewClient(conn), nil
}
