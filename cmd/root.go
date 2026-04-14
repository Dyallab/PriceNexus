package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	dbPath  string
)

var RootCmd = &cobra.Command{
	Use:   "pricenexus",
	Short: "PriceNexus - Sistema de búsqueda de precios",
	Long:  `CLI para buscar y trackear precios de productos en tiendas de Argentina`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			fmt.Println("Verbose mode enabled")
		}
	},
}

func init() {

	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().StringVar(&dbPath, "db-path", "prices.db", "database path")
}
