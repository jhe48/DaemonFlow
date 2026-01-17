package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jackyhe0402/daemonflow/internal/daemon"
	"github.com/jackyhe0402/daemonflow/internal/ipc"
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
		Long:  `Stop the running DaemonFlow daemon via IPC.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			d := daemon.New()
			socketPath := d.GetSocketPath()

			// Check if socket exists
			if _, err := os.Stat(socketPath); os.IsNotExist(err) {
				fmt.Println("DaemonFlow daemon is not running")
				return nil
			}

			// Try to stop via IPC
			client := ipc.NewClient(socketPath)
			if err := client.Stop(); err != nil {
				// Fall back to PID-based stop if IPC fails
				fmt.Printf("IPC stop failed (%v), trying signal-based stop...\n", err)
				return d.Stop()
			}

			fmt.Println("Stop request sent to daemon")

			// Wait briefly for shutdown
			time.Sleep(500 * time.Millisecond)

			// Check if daemon is still running
			if client.IsConnected() {
				fmt.Println("Daemon is still shutting down...")
			} else {
				fmt.Println("Daemon stopped successfully")
			}

			return nil
		},
	}

	// Status command
	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Check daemon status",
		Long:  `Check if the DaemonFlow daemon is running and display status via IPC.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			d := daemon.New()
			socketPath := d.GetSocketPath()

			// Check if socket exists
			if _, err := os.Stat(socketPath); os.IsNotExist(err) {
				fmt.Println("DaemonFlow daemon is not running")
				return nil
			}

			// Try to get status via IPC
			client := ipc.NewClient(socketPath)
			status, err := client.Status()
			if err != nil {
				// Socket exists but can't connect - may be stale
				fmt.Printf("DaemonFlow daemon is not responding (socket exists but connection failed)\n")
				fmt.Printf("Socket: %s\n", socketPath)
				fmt.Println("\nTry 'daemonflow stop' to clean up stale socket")
				return nil
			}

			// Display nice formatted status
			fmt.Println("DaemonFlow Status")
			fmt.Println("-----------------")
			if status.Running {
				fmt.Println("Running: yes")
			} else {
				fmt.Println("Running: no")
			}
			fmt.Printf("Uptime: %s\n", formatDuration(time.Duration(status.Uptime)*time.Second))
			fmt.Printf("Watch Dir: %s\n", status.WatchDir)
			fmt.Printf("Socket: %s\n", socketPath)

			return nil
		},
	}

	rootCmd.AddCommand(startCmd, stopCmd, statusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)

	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour

	hours := d / time.Hour
	d -= hours * time.Hour

	minutes := d / time.Minute
	d -= minutes * time.Minute

	seconds := d / time.Second

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
