package main

import (
	"github.com/diillson/k8s-multicluster-cli/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logrus.Fatalf("Error executing command: %v", err)
	}
}
