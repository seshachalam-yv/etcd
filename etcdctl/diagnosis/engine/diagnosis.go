package engine

import (
       "encoding/json"
       "fmt"
       "log"
       "os"

       "go.etcd.io/etcd/etcdctl/v3/diagnosis/engine/intf"
)

type report struct {
	Input   any   `json:"input,omitempty"`
	Results []any `json:"results,omitempty"`
}

// Diagnose runs all provided plugins and outputs a report. If outputFile is an
// empty string, the report is written to stdout, otherwise it's written to the
// specified file.
func Diagnose(input any, plugins []intf.Plugin, outputFile string) {
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

       if outputFile == "" {
               fmt.Fprintln(os.Stdout, string(b))
               return
       }

       if err := os.WriteFile(outputFile, b, 0644); err != nil {
               log.Fatalf("Failed to write the report to file: %v", err)
       }
}
