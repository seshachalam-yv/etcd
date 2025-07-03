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

package agent

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.etcd.io/etcd/client/pkg/v3/logutil"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func commandCtx(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

func createClient(cfgSpec *clientv3.ConfigSpec) (*clientv3.Client, error) {
	lg, _ := logutil.CreateDefaultZapLogger(zap.InfoLevel)
	cfg, err := clientv3.NewClientConfig(cfgSpec, lg)
	if err != nil {
		return nil, err
	}
	return clientv3.New(*cfg)
}
