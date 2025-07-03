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

package engine

import (
	"encoding/json"
	"log"
	"os"

	"go.etcd.io/etcd/etcdctl/v3/diagnosis/engine/intf"
)

const (
	reportFileName = "etcd_diagnosis_report.json"
)

type report struct {
	Input   any   `json:"input,omitempty"`
	Results []any `json:"results,omitempty"`
}

func Diagnose(input any, plugins []intf.Plugin) {
	rp := report{
		Input: input,
	}
	for i, plugin := range plugins {
		log.Println("---------------------------------------------------------")
		log.Printf("Running %q (%d/%d)...\n", plugin.Name(), i+1, len(plugins))

		result := plugin.Diagnose()
		rp.Results = append(rp.Results, result)

		b, err := json.MarshalIndent(result, "", "\t")
		if err != nil {
			log.Printf("Failed to marshal result for plugin %q: %v", plugin.Name(), err)
			continue
		}
		log.Println(string(b))
	}

	b, err := json.MarshalIndent(rp, "", "\t")
	if err != nil {
		log.Fatalf("Failed to marshal the report: %v", err)
	}

	if err := os.WriteFile(reportFileName, b, 0o644); err != nil {
		log.Fatalf("Failed to write the report to file: %v", err)
	}
}
