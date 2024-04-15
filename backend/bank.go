package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo"
)

type Bank struct {
	db     *sql.DB
	update *sql.Stmt
	total  int
}

type Request struct {
	Sum int `json:"sum"`
}

func (b *Bank) Init(host, port, user, password, dbname string) error {
	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	log.Printf("Connecting to %s\n", psqlconn)
	db, err := sql.Open("postgres", psqlconn)

	if err != nil {
		return err
	}

	err = db.Ping()

	if err != nil {
		return err
	}

	b.db = db

	b.update, err = b.db.Prepare(
		`
			UPDATE bank_data 
			SET total = $1;
		`)

	if err != nil {
		return err
	}

	rows, err := b.db.Query("SELECT * FROM bank_data;")

	b.total = 1000
	if err != nil {
		return err
	}

	any := false

	for rows.Next() {
		rows.Scan(&b.total)
		log.Print("\t total updated", b.total)
		any = true
	}

	if !any {
		log.Println("Table is empty, adding one row")
		_, err := b.db.Query("INSERT INTO bank_data VALUES (1000);")

		if err != nil {
			return err
		}
	}

	b.update.Exec(b.total)
	log.Println("Bank is ready")
	return nil
}

func (b *Bank) Handle(e echo.Context) error {
	fmt.Println("handling")
	action := e.Get("Action")

	r := Request{}

	err := e.Bind(&r)

	if err != nil {
		log.Println("Error: \"sum\" value is not an int")
		return echo.NewHTTPError(http.StatusBadRequest, "sum must be an int")
	}

	if action == "Add" {
		log.Print("Adding ", r.Sum)
		b.total += r.Sum
	} else if action == "Sub" {
		if b.total-r.Sum >= 0 {
			log.Print("Substructing ", r.Sum)
			b.total -= r.Sum
		} else {
			log.Println("Insufficient funds")

			return e.String(http.StatusBadRequest, "Insufficient funds")
		}
	} else {
		log.Println("Error: unknown action")
		return e.String(http.StatusBadRequest, "unknown action")
	}

	b.update.Exec(b.total)

	resp := fmt.Sprintf("New total = %d", b.total)

	log.Println(resp)
	return e.String(http.StatusOK, resp)
}

func (b *Bank) Total(e echo.Context) error {
	rows, err := b.db.Query("SELECT * FROM bank_data;")

	if err != nil {
		return err
	}

	for rows.Next() {
		rows.Scan(&b.total)
	}

	log.Print("Actual total is ", b.total)

	return e.String(http.StatusOK, fmt.Sprint(b.total))
}

func ActionBasedOnRole(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role := c.Request().Header.Get("User-Role")

		if role == "" {
			return echo.NewHTTPError(http.StatusForbidden, "acces denied")
		}

		if role == "client" {
			c.Set("Action", "Add")
		} else if role == "admin" {
			c.Set("Action", "Sub")
		} else {
			return echo.NewHTTPError(http.StatusForbidden, "acces denied")
		}

		return next(c)
	}
}
