package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gocolly/colly"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Dictionary map[string]string

type RecipeSpecs struct {
	Difficulty  string
	PrepTime    string
	CookingTime string
	ServingSize string
	PriceTier   string
}

type Recipe struct {
	ID          uint `gorm:"primaryKey"`
	URL         string
	Name        string
	Difficulty  string
	PrepTime    string
	CookingTime string
	ServingSize string
	PriceTier   string
	Ingredients string `gorm:"type:TEXT"`
}

func main() {
	args := os.Args
	url := args[1]

	// Connect to the PostgreSQL database using GORM
	dsn := "host=localhost user=postgres password=root dbname=recipes port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}

	// Create the recipes table if it doesn't exist
	err = db.AutoMigrate(&Recipe{})
	if err != nil {
		fmt.Println("Error migrating table:", err)
		return
	}

	collector := colly.NewCollector()
	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	collector.OnResponse(func(r *colly.Response) {
		fmt.Println("Got a response from", r.Request.URL)
	})

	collector.OnError(func(r *colly.Response, e error) {
		fmt.Println("An error occurred!:", e)
	})

	collector.OnHTML("main", func(main *colly.HTMLElement) {
		recipe := Recipe{}

		recipe.URL = url
		recipe.Name = main.ChildText(".gz-title-recipe")
		fmt.Println("Scraping recipe for:", recipe.Name)

		main.ForEach(".gz-name-featured-data", func(i int, specListElement *colly.HTMLElement) {
			if strings.Contains(specListElement.Text, "Difficoltà: ") {
				recipe.Difficulty = specListElement.ChildText("strong")
			}
			if strings.Contains(specListElement.Text, "Preparazione: ") {
				recipe.PrepTime = specListElement.ChildText("strong")
			}
			if strings.Contains(specListElement.Text, "Cottura: ") {
				recipe.CookingTime = specListElement.ChildText("strong")
			}
			if strings.Contains(specListElement.Text, "Dosi per: ") {
				recipe.ServingSize = specListElement.ChildText("strong")
			}
			if strings.Contains(specListElement.Text, "Costo: ") {
				recipe.PriceTier = specListElement.ChildText("strong")
			}
		})

		// Ingredients will be stored as a string
		var ingredients []string
		main.ForEach(".gz-ingredient", func(i int, ingredient *colly.HTMLElement) {
			ingredients = append(ingredients, fmt.Sprintf("%s: %s", ingredient.ChildText("a"), ingredient.ChildText("span")))
		})
		recipe.Ingredients = strings.Join(ingredients, ", ")

		// Save the recipe to the database
		err := db.Create(&recipe).Error
		if err != nil {
			fmt.Println("Error saving recipe:", err)
			return
		}

		fmt.Println("Recipe saved successfully")
	})

	collector.Visit(url)
}
