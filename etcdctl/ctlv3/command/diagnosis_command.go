package command

import (
       "time"

       "github.com/spf13/cobra"

       "go.etcd.io/etcd/etcdctl/v3/diagnosis/agent"
       "go.etcd.io/etcd/etcdctl/v3/diagnosis/engine"
       "go.etcd.io/etcd/etcdctl/v3/diagnosis/engine/intf"
       "go.etcd.io/etcd/etcdctl/v3/diagnosis/plugins/epstatus"
       "go.etcd.io/etcd/etcdctl/v3/diagnosis/plugins/membership"
       "go.etcd.io/etcd/etcdctl/v3/diagnosis/plugins/metrics"
       readplugin "go.etcd.io/etcd/etcdctl/v3/diagnosis/plugins/read"
)

var diagCfg = agent.GlobalConfig{}

// NewDiagnosisCommand returns the cobra command for "diagnosis".
func NewDiagnosisCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diagnosis",
		Short: "One-stop etcd diagnosis tool",
		Run:   runDiagnosis,
	}

       cmd.Flags().BoolVar(&diagCfg.UseClusterEndpoints, "cluster", false, "use all endpoints from the cluster member list")
       cmd.Flags().IntVar(&diagCfg.DbQuotaBytes, "etcd-storage-quota-bytes", 2*1024*1024*1024, "etcd storage quota in bytes (the value passed to etcd instance by flag --quota-backend-bytes)")
       cmd.Flags().StringVarP(&diagCfg.OutputFile, "output", "o", "", "output file (if empty, print to stdout)")

	return cmd
}

func runDiagnosis(cmd *cobra.Command, args []string) {
       // populate diagCfg from global flags
       diagCfg.Endpoints = globalFlags.Endpoints
       diagCfg.DialTimeout = globalFlags.DialTimeout
       diagCfg.CommandTimeout = globalFlags.CommandTimeOut
       diagCfg.KeepAliveTime = globalFlags.KeepAliveTime
       diagCfg.KeepAliveTimeout = globalFlags.KeepAliveTimeout
       diagCfg.Insecure = globalFlags.Insecure
       diagCfg.InsecureSkipVerify = globalFlags.InsecureSkipVerify
       diagCfg.InsecureDiscovery = globalFlags.InsecureDiscovery
       diagCfg.CertFile = globalFlags.TLS.CertFile
       diagCfg.KeyFile = globalFlags.TLS.KeyFile
       diagCfg.CaFile = globalFlags.TLS.TrustedCAFile
       diagCfg.DNSDomain = globalFlags.TLS.ServerName
       diagCfg.DNSService = globalFlags.DNSClusterServiceName
       diagCfg.Username = globalFlags.User
       diagCfg.Password = globalFlags.Password

       plugins := []intf.Plugin{
               membership.NewPlugin(diagCfg),
               epstatus.NewPlugin(diagCfg),
               readplugin.NewPlugin(diagCfg, false),
               readplugin.NewPlugin(diagCfg, true),
               metrics.NewPlugin(diagCfg),
       }
       engine.Diagnose(diagCfg, plugins, diagCfg.OutputFile)
}
