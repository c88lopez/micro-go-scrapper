package postgres

import (
	"database/sql"
	"fmt"
	"log"

	"sarasa/schemas"

	_ "github.com/lib/pq"
)

type Client struct {
	connection *sql.DB
	txn        *sql.Tx
}

func (postgres *Client) Init(pc schemas.PostgresConfig) error {
	log.Println("Initializing Postgres client...")

	var err error

	dataSource := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", pc.User, pc.Password, pc.Host, pc.Database)

	postgres.connection, err = sql.Open("postgres", dataSource)
	if err != nil {
		return err
	}

	return nil
}

func (postgres Client) Close() error {
	return postgres.connection.Close()
}

func (rabbitmq Client) String() string {
	return "Postgres"
}
