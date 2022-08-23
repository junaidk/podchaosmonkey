package main

import (
	"github.com/spf13/cobra"
	"log"
	"os"
	"podchaosmonkey/pkg/cli"
)

func main() {
	cmd, err := rootCmd()

	if err != nil {
		log.Fatalf("error configuring commands : %s", err.Error())
	}
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "podchaosmonkey",
		Short: "Run the pod chaos monkey process",
	}
	processCmd, err := cli.NewProcessCMD()
	cmd.AddCommand(processCmd)
	return cmd, err
}
