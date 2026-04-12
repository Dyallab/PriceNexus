package cmd

import (
	"fmt"

	"github.com/dyallo/pricenexus/internal/db"
	"github.com/dyallo/pricenexus/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var addShopCmd = &cobra.Command{
	Use:   "shop [nombre] [url]",
	Short: "Agregar una nueva tienda",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		log := logrus.New()
		name := args[0]
		url := args[1]

		repo, err := db.NewRepository(dbPath, log)
		if err != nil {
			fmt.Printf("Error conectando a DB: %v\n", err)
			return
		}
		defer repo.Close()

		shop := models.Shop{
			Name:   name,
			URL:    url,
			Active: true,
		}

		id, err := repo.AddShop(shop)
		if err != nil {
			fmt.Printf("Error guardando tienda: %v\n", err)
			return
		}

		fmt.Printf("Tienda '%s' agregada con ID: %d\n", name, id)
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Agregar datos al sistema",
}

func init() {
	addCmd.AddCommand(addShopCmd)
	RootCmd.AddCommand(addCmd)
}
