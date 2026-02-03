package postgres_connection

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

func CheckConnection() {
	ctx := context.Background()

	con, err := pgx.Connect(ctx, "postgres://postgres:pass@localhost:5433/postgres")
	if err != nil {
		panic(err)
	}

	if err := con.Ping(ctx); err != nil {
		panic(err)
	}

	fmt.Println("все ок")

}
