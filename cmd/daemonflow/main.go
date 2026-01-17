package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0-dev"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "daemonflow",
		Short: "DaemonFlow - Background automation daemon",
		Long:  `DaemonFlow is a daemon that monitors your development environment and runs automation scripts in the background.`,
	}

	// Version flag
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("daemonflow version {{.Version}}\n")

	// Start command
	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the daemon",
		Long:  `Start the DaemonFlow daemon in the background.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting DaemonFlow daemon...")
		},
	}

	// Stop command
	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop the daemon",
		Long:  `Stop the running DaemonFlow daemon.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Stopping DaemonFlow daemon...")
		},
	}

	// Status command
	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Check daemon status",
		Long:  `Check if the DaemonFlow daemon is running.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Checking DaemonFlow daemon status...")
		},
	}

	rootCmd.AddCommand(startCmd, stopCmd, statusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
