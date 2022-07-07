// Copyright 2015 The etcd Authors
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
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	v3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/pkg/v3/report"

	"github.com/dustin/go-humanize"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
	"gopkg.in/cheggaaa/pb.v1"
)

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put",
	Short: "Benchmark put",

	Run: putFunc,
}

var (
	keySize int
	valSize int

	putTotal int
	putRate  int

	keySpaceSize int

	compactInterval   time.Duration
	compactIndexDelta int64

	checkHashkv bool
	repeat      int
)

func init() {
	RootCmd.AddCommand(putCmd)
	putCmd.Flags().IntVar(&repeat, "repeat", 10, "Repeat N times the same test")
	putCmd.Flags().IntVar(&keySize, "key-size", 8, "Key size of put request")
	putCmd.Flags().IntVar(&valSize, "val-size", 8, "Value size of put request")
	putCmd.Flags().IntVar(&putRate, "rate", 0, "Maximum puts per second (0 is no limit)")

	putCmd.Flags().IntVar(&putTotal, "total", 10000, "Total number of put requests")
	putCmd.Flags().IntVar(&keySpaceSize, "key-space-size", 1, "Maximum possible keys")
	putCmd.Flags().DurationVar(&compactInterval, "compact-interval", 0, `Interval to compact database (do not duplicate this with etcd's 'auto-compaction-retention' flag) (e.g. --compact-interval=5m compacts every 5-minute)`)
	putCmd.Flags().Int64Var(&compactIndexDelta, "compact-index-delta", 1000, "Delta between current revision and compact revision (e.g. current revision 10000, compact at 9000)")
	putCmd.Flags().BoolVar(&checkHashkv, "check-hashkv", false, "'true' to check hashkv")
}

func putFunc(cmd *cobra.Command, args []string) {
	if keySpaceSize <= 0 {
		fmt.Fprintf(os.Stderr, "expected positive --key-space-size, got (%v)", keySpaceSize)
		os.Exit(1)
	}

	requests := make(chan v3.Op, totalClients)
	if putRate == 0 {
		putRate = math.MaxInt32
	}
	limit := rate.NewLimiter(rate.Limit(putRate), 1)
	clients := mustCreateClients(totalClients, totalConns)
	v := string(mustRandBytes(valSize))
	k, v := "", string(mustRandBytes(valSize))

	bar = pb.New(putTotal)
	bar.Format("Bom !")
	bar.Start()
	if len(pushgateway) > 0 {
		latencyMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "etcd_benchmark_op_duration_seconds",
			Help: "The duration of the last DB backup in seconds.",
		}, []string{"key", "method", "test_name"})

		registry := prometheus.NewRegistry()
		registry.MustRegister(latencyMetric)
		pusher = push.New(pushgateway, "etcd_benchmark").Gatherer(registry)
	}
	r := newReport()
	for i := range clients {
		wg.Add(1)
		go func(c *v3.Client) {
			defer wg.Done()
			for op := range requests {
				limit.Wait(context.Background())

				st := time.Now()
				_, err := c.Do(context.Background(), op)
				fmt.Println(string(op.KeyBytes()), "put", testName, float64(time.Since(st).Milliseconds()))
				if len(pushgateway) > 0 {
					latencyMetric.WithLabelValues(string(op.KeyBytes()), "put", testName).Set(float64(time.Since(st).Milliseconds()))
				}
				r.Results() <- report.Result{Err: err, Start: st, End: time.Now()}
				bar.Increment()
			}
		}(clients[i])
	}

	go func() {
		for j := 0; j < repeat; j++ {
			for i := 0; i < putTotal; i++ {
				if seqKeys {
					k = fmt.Sprintf("%v", keyStarts)
					keyStarts++
				} else {
					k = fmt.Sprintf("%v", i)
				}
				requests <- v3.OpPut(k, v)
			}
		}
		close(requests)
	}()

	if compactInterval > 0 {
		go func() {
			for {
				time.Sleep(compactInterval)
				compactKV(clients)
			}
		}()
	}

	rc := r.Run()
	wg.Wait()
	close(r.Results())
	bar.Finish()
	fmt.Println(<-rc)
	if len(pushgateway) > 0 {
		if err := pusher.Add(); err != nil {
			fmt.Println("Could not push to Pushgateway:", err)
		}
	}

	if checkHashkv {
		hashKV(cmd, clients)
	}
}

func compactKV(clients []*v3.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := clients[0].KV.Get(ctx, "foo")
	cancel()
	if err != nil {
		panic(err)
	}
	revToCompact := max(0, resp.Header.Revision-compactIndexDelta)
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	_, err = clients[0].KV.Compact(ctx, revToCompact)
	cancel()
	if err != nil {
		panic(err)
	}
}

func max(n1, n2 int64) int64 {
	if n1 > n2 {
		return n1
	}
	return n2
}

func hashKV(cmd *cobra.Command, clients []*v3.Client) {
	eps, err := cmd.Flags().GetStringSlice("endpoints")
	if err != nil {
		panic(err)
	}
	for i, ip := range eps {
		eps[i] = strings.TrimSpace(ip)
	}
	host := eps[0]

	st := time.Now()
	rh, eh := clients[0].HashKV(context.Background(), host, 0)
	if eh != nil {
		fmt.Fprintf(os.Stderr, "Failed to get the hashkv of endpoint %s (%v)\n", host, eh)
		panic(err)
	}
	rt, es := clients[0].Status(context.Background(), host)
	if es != nil {
		fmt.Fprintf(os.Stderr, "Failed to get the status of endpoint %s (%v)\n", host, es)
		panic(err)
	}

	rs := "HashKV Summary:\n"
	rs += fmt.Sprintf("\tHashKV: %d\n", rh.Hash)
	rs += fmt.Sprintf("\tEndpoint: %s\n", host)
	rs += fmt.Sprintf("\tTime taken to get hashkv: %v\n", time.Since(st))
	rs += fmt.Sprintf("\tDB size: %s", humanize.Bytes(uint64(rt.DbSize)))
	fmt.Println(rs)
}
