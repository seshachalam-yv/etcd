// Copyright 2025 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"go.etcd.io/etcd/etcdctl/v3/diagnosis/agent"
	"go.etcd.io/etcd/etcdctl/v3/diagnosis/engine"
	intf "go.etcd.io/etcd/etcdctl/v3/diagnosis/engine/intf"
	"go.etcd.io/etcd/etcdctl/v3/diagnosis/offline"
	"go.etcd.io/etcd/etcdctl/v3/diagnosis/plugins/epstatus"
	"go.etcd.io/etcd/etcdctl/v3/diagnosis/plugins/membership"
	"go.etcd.io/etcd/etcdctl/v3/diagnosis/plugins/metrics"
	readplugin "go.etcd.io/etcd/etcdctl/v3/diagnosis/plugins/read"
)

var diagCfg = agent.GlobalConfig{}

// NewRootCommand returns the Cobra command for etcd-diagnosis.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "etcd-diagnosis",
		Short: "One-stop etcd diagnosis tool",
		Run:   runDiagnosis,
	}

	cmd.Flags().StringSliceVar(&diagCfg.Endpoints, "endpoints", []string{"127.0.0.1:2379"}, "comma separated etcd endpoints")
	cmd.Flags().BoolVar(&diagCfg.UseClusterEndpoints, "cluster", false, "use all endpoints from the cluster member list")

	cmd.Flags().DurationVar(&diagCfg.DialTimeout, "dial-timeout", 2*time.Second, "dial timeout for client connections")
	cmd.Flags().DurationVar(&diagCfg.CommandTimeout, "command-timeout", 5*time.Second, "command timeout (excluding dial timeout)")
	cmd.Flags().DurationVar(&diagCfg.KeepAliveTime, "keepalive-time", 2*time.Second, "keepalive time for client connections")
	cmd.Flags().DurationVar(&diagCfg.KeepAliveTimeout, "keepalive-timeout", 5*time.Second, "keepalive timeout for client connections")

	cmd.Flags().BoolVar(&diagCfg.Insecure, "insecure-transport", true, "disable transport security for client connections")
	cmd.Flags().BoolVar(&diagCfg.InsecureSkipVerify, "insecure-skip-tls-verify", false, "skip server certificate verification")
	cmd.Flags().StringVar(&diagCfg.CertFile, "cert", "", "identify secure client using this TLS certificate file")
	cmd.Flags().StringVar(&diagCfg.KeyFile, "key", "", "identify secure client using this TLS key file")
	cmd.Flags().StringVar(&diagCfg.CaFile, "cacert", "", "verify certificates of TLS-enabled secure servers using this CA bundle")

	cmd.Flags().StringVar(&diagCfg.Username, "user", "", "username[:password] for authentication (prompt if password is not supplied)")
	cmd.Flags().StringVar(&diagCfg.Password, "password", "", "password for authentication (if this option is used, --user option shouldn't include password)")
	cmd.Flags().StringVarP(&diagCfg.DNSDomain, "discovery-srv", "d", "", "domain name to query for SRV records describing cluster endpoints")
	cmd.Flags().StringVarP(&diagCfg.DNSService, "discovery-srv-name", "", "", "service name to query when using DNS discovery")
	cmd.Flags().BoolVar(&diagCfg.InsecureDiscovery, "insecure-discovery", true, "accept insecure SRV records describing cluster endpoints")

	cmd.Flags().IntVar(&diagCfg.DbQuotaBytes, "etcd-storage-quota-bytes", 2*1024*1024*1024, "etcd storage quota in bytes (the value passed to etcd instance by flag --quota-backend-bytes)")

	cmd.Flags().BoolVar(&diagCfg.PrintVersion, "version", false, "print the version and exit")

	cmd.Flags().BoolVar(&diagCfg.Offline, "offline", false, "offline analysis")
	cmd.Flags().StringVar(&diagCfg.DataDir, "data-dir", "", "path to data directory")

	return cmd
}

func runDiagnosis(cmd *cobra.Command, args []string) {
	if diagCfg.Offline {
		offline.AnalyzeOffline(diagCfg.DataDir)
		return
	}

	plugins := []intf.Plugin{
		membership.NewPlugin(diagCfg),
		epstatus.NewPlugin(diagCfg),
		readplugin.NewPlugin(diagCfg, false),
		readplugin.NewPlugin(diagCfg, true),
		metrics.NewPlugin(diagCfg),
	}
	engine.Diagnose(diagCfg, plugins)
}
