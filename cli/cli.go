package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/imafaz/cdng/api"
	"github.com/imafaz/cdng/controller"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cdng",
	Short: "NGINX Proxy server Tool",
	Long:  "A command-line tool for managing NGINX configurations and operations with api:)",
}

var addDomainCmd = &cobra.Command{
	Use:   "add-domain [domain] [ip]",
	Short: "Add a new domain",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := controller.AddDomain(args[0], args[1]); err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println("Domain added successfully.")
	},
}

var deleteDomainCmd = &cobra.Command{
	Use:   "delete-domain [domain]",
	Short: "Delete a domain",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := controller.DeleteDomain(args[0]); err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println("Domain deleted successfully.")
	},
}

var addPortCmd = &cobra.Command{
	Use:   "add-port [port]",
	Short: "Add a new port",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := controller.AddPort(args[0]); err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println("Port added successfully.")
	},
}

var deletePortCmd = &cobra.Command{
	Use:   "delete-port [port]",
	Short: "Delete a port",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := controller.DeletePort(args[0]); err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println("Port deleted successfully.")
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart Nginx",
	Run: func(cmd *cobra.Command, args []string) {
		if err := controller.RestartNginx(); err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println("Nginx restarted successfully.")
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get system statistics",
	Run: func(cmd *cobra.Command, args []string) {
		stats, err := controller.GetStats()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Printf("%+v\n", stats)
	},
}

var startApiCmd = &cobra.Command{
	Use:   "start-api [port] [apiKey]",
	Short: "Start the API server",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		port := args[0]
		apiKey := args[1]
		fmt.Printf("Starting API server on port %s with API key %s\n", port, apiKey)
		api.StartServer(port, apiKey)
	},
}

func init() {
	rootCmd.AddCommand(addDomainCmd, deleteDomainCmd, addPortCmd, deletePortCmd, restartCmd, statsCmd, startApiCmd)
}

func StartCLI() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
