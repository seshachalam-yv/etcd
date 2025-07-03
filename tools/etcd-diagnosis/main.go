package main

import (
	"fmt"
	"os"

	"go.etcd.io/etcd/v3/tools/etcd-diagnosis/cmd"
)

func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
