package cmd

import (
	"fmt"
	"time"

	"github.com/dyallo/pricenexus/internal/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history [producto]",
	Short: "Ver historial de precios de un producto",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log := logrus.New()
		query := args[0]

		repo, err := db.NewRepository(dbPath, log)
		if err != nil {
			fmt.Printf("Error conectando a DB: %v\n", err)
			return
		}
		defer repo.Close()

		product, err := repo.GetProduct(query)
		if err != nil {
			fmt.Printf("Producto '%s' no encontrado\n", query)
			return
		}

		prices, err := repo.GetPriceHistoryByProduct(product.ID)
		if err != nil {
			fmt.Printf("Error obteniendo precios: %v\n", err)
			return
		}

		fmt.Printf("Historial de precios para: %s\n\n", product.Name)

		if len(prices) == 0 {
			fmt.Println("No hay precios registrados")
			return
		}

		for _, p := range prices {
			shopName := fmt.Sprintf("Tienda %d", p.ShopID)
			shop, err := repo.GetShopByID(p.ShopID)
			if err == nil && shop.Name != "" {
				shopName = shop.Name
			}

			scrapedAt := p.ScrapedAt
			if parsedTime, err := time.Parse(time.RFC3339, p.ScrapedAt); err == nil {
				scrapedAt = parsedTime.Format("2006-01-02 15:04")
			}

			fmt.Printf("[%s] $%.2f %s - %s\n",
				shopName, p.Price, p.Currency, scrapedAt)
		}
	},
}

func init() {
	RootCmd.AddCommand(historyCmd)
}
