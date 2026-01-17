package main

import (
	"fmt"
	"os"

	"github.com/jackyhe0402/daemonflow/internal/daemon"
	"github.com/spf13/cobra"
)

var (
	version    = "0.1.0-dev"
	foreground bool
	configPath string
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
		RunE: func(cmd *cobra.Command, args []string) error {
			d := daemon.New()
			d.Foreground = foreground
			d.ConfigPath = configPath
			return d.Start()
		},
	}
	startCmd.Flags().BoolVar(&foreground, "foreground", false, "Run in foreground (don't daemonize)")
	startCmd.Flags().StringVar(&configPath, "config", "", "Path to config file (default: ~/.daemonflow/config.yaml)")

	// Stop command
	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop the daemon",
		Long:  `Stop the running DaemonFlow daemon.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			d := daemon.New()
			return d.Stop()
		},
	}

	// Status command
	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Check daemon status",
		Long:  `Check if the DaemonFlow daemon is running.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			d := daemon.New()
			running, pid, err := d.Status()
			if err != nil {
				return err
			}
			if running {
				fmt.Printf("DaemonFlow daemon is running (PID: %d)\n", pid)
			} else {
				fmt.Println("DaemonFlow daemon is not running")
			}
			return nil
		},
	}

	rootCmd.AddCommand(startCmd, stopCmd, statusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
