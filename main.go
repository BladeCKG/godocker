package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cenkalti/backoff/v4"
	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/gin-gonic/gin"
	"rsc.io/quote"
)

func main() {
	fmt.Println(quote.Go())
	r := gin.Default()
	db, err := initStore()
	if err != nil {
		log.Fatalf("failed to init db %s", err)
	}
	defer db.Close()
	r.GET("/", func(c *gin.Context) {
		rootHandler(db, c)
		// c.String(200, "Hello world")
	})
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"Status": "ok",
		})
	})
	r.POST("/send", func(ctx *gin.Context) {
		sendHandler(db, ctx)
	})
	log.Fatal(r.Run())
}

type Message struct {
	Value string `json:"value"`
}

func initStore() (*sql.DB, error) {

	pgConnString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		os.Getenv("PGHOST"),
		os.Getenv("PGPORT"),
		os.Getenv("PGDATABASE"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
	)

	log.Println(pgConnString)
	var (
		db  *sql.DB
		err error
	)
	openDB := func() error {
		db, err = sql.Open("postgres", pgConnString)
		return err
	}

	err = backoff.Retry(openDB, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(
		"CREATE TABLE IF NOT EXISTS message (value TEXT PRIMARY KEY)"); err != nil {
		return nil, err
	}

	return db, nil
}

func rootHandler(db *sql.DB, c *gin.Context) {
	r, err := countRecords(db)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("Hello, Docker! (%d)\n", r))
}

func sendHandler(db *sql.DB, c *gin.Context) {

	m := &Message{}

	if err := c.Bind(m); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	err := crdb.ExecuteTx(context.Background(), db, nil,
		func(tx *sql.Tx) error {
			_, err := tx.Exec(
				"INSERT INTO message (value) VALUES ($1) ON CONFLICT (value) DO UPDATE SET value = excluded.value",
				m.Value,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return err
			}
			return nil
		})

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, m)
}

func countRecords(db *sql.DB) (int, error) {

	rows, err := db.Query("SELECT COUNT(*) FROM message")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, err
		}
		rows.Close()
	}

	return count, nil
}
