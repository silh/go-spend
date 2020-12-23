package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	pgxpool.Connect(context.Background(), "")

}
