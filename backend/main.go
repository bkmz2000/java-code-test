package main

import (
	"log"
	"os"

	"github.com/labstack/echo"
	_ "github.com/lib/pq"
)

func main() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	bank := &Bank{}
	err := bank.Init(host, port, user, password, dbname)
	if err != nil {
		log.Fatal("Failed to initialize bank:", err)
	}

	e := echo.New()

	e.POST("/bank", bank.Handle, ActionBasedOnRole)
	e.GET("/total", bank.Total)

	e.Logger.Fatal(e.Start(":8080"))
}
