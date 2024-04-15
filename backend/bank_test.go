package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

var test_bank Bank

func Test_Init(t *testing.T) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	err := test_bank.Init(host, port, user, password, dbname)

	if err != nil {
		panic(err)
	}

	assert.NoError(t, err)
	assert.NotNil(t, test_bank.db)
	assert.NotNil(t, test_bank.update)
	assert.Equal(t, 1000, test_bank.total)
}

func TestHandle_Add(t *testing.T) {
	transaction, err := test_bank.db.Begin()

	defer transaction.Rollback()
	if err != nil {
		panic(err)
	}

	test_bank.total = 0

	e := echo.New()
	reqBody := `{"sum":10}`
	req := httptest.NewRequest(http.MethodPost, "/handle", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("Action", "Add")

	err = test_bank.Handle(c)

	if err != nil {
		panic(err)
	}

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "New total = 10", rec.Body.String())
}

func TestHandle_Sub_SufficientFunds(t *testing.T) {
	transaction, err := test_bank.db.Begin()

	defer transaction.Rollback()
	if err != nil {
		panic(err)
	}

	test_bank.total = 20

	e := echo.New()
	reqBody := `{"sum":10}`
	req := httptest.NewRequest(http.MethodPost, "/handle", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("Action", "Sub")

	err = test_bank.Handle(c)

	if err != nil {
		panic(err)
	}

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "New total = 10", rec.Body.String())
}

func TestHandle_Sub_InsufficientFunds(t *testing.T) {
	transaction, err := test_bank.db.Begin()

	defer transaction.Rollback()
	if err != nil {
		panic(err)
	}

	test_bank.total = 0

	e := echo.New()
	reqBody := `{"sum":10}`
	req := httptest.NewRequest(http.MethodPost, "/handle", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("Action", "Sub")

	err = test_bank.Handle(c)

	if err != nil {
		panic(err)
	}

	assert.Equal(t, "Insufficient funds", rec.Body.String())
}

// Test Handle method with unknown action
func TestHandle_UnknownAction(t *testing.T) {
	transaction, err := test_bank.db.Begin()

	defer transaction.Rollback()
	if err != nil {
		panic(err)
	}

	e := echo.New()
	reqBody := `{"sum":10}`
	req := httptest.NewRequest(http.MethodPost, "/handle", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("Action", "InvalidAction")

	err = test_bank.Handle(c)

	if err != nil {
		panic(err)
	}

	assert.Equal(t, "unknown action", rec.Body.String())
}

func Test_Final(t *testing.T) {
	test_bank.update.Exec(1000)
}
