package common

import "go.etcd.io/etcd/v3/tools/etcd-diagnosis/agent"

type Checker struct {
	agent.GlobalConfig
	Name string
}
