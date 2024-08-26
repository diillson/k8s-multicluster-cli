package main

import (
	"github.com/sirupsen/logrus"
	"multicluster/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logrus.Fatalf("Error executing command: %v", err)
	}
}
