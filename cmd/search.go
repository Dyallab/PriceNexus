package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dyallo/pricenexus/internal/agent/orchestrator"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	searchCmd = &cobra.Command{
		Use:   "search [producto]",
		Short: "Buscar precios de un producto",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			query := args[0]

			// Create logs directory if it doesn't exist
			logsDir := filepath.Join(".", "logs")
			if err := os.MkdirAll(logsDir, 0755); err != nil {
				fmt.Printf("Warning: Could not create logs directory: %v\n", err)
			}

			// Setup logger to write to both file and console
			log := logrus.New()
			log.SetLevel(logrus.InfoLevel)

			// Log to file
			logFile, err := os.OpenFile(
				filepath.Join(logsDir, "search.log"),
				os.O_CREATE|os.O_WRONLY|os.O_APPEND,
				0666,
			)
			if err != nil {
				fmt.Printf("Warning: Could not open log file: %v\n", err)
			} else {
				log.SetOutput(logFile)
				defer logFile.Close()
			}

			// Also log to console
			log.SetFormatter(&logrus.TextFormatter{
				FullTimestamp: true,
			})

			orch, err := orchestrator.NewOrchestrator(dbPath, log)
			if err != nil {
				fmt.Printf("Error inicializando sistema: %v\n", err)
				return
			}
			defer orch.Close()

			fmt.Printf("Buscando: %s\n\n", query)

			results, err := orch.Search(context.Background(), query)
			if err != nil {
				fmt.Printf("Error en búsqueda: %v\n", err)
				return
			}

			if len(results) == 0 {
				fmt.Println("No se encontraron resultados.")
				return
			}

			for _, r := range results {
				fmt.Printf("  - %s: $%.2f %s\n", r.ProductName, r.Price, r.Currency)
				fmt.Printf("    URL: %s\n", r.URL)
			}
		},
	}
)

func init() {
	RootCmd.AddCommand(searchCmd)
}
