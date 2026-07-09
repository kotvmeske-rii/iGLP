package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func CreateTable(ctx context.Context, conn *pgx.Conn) {
	sqlQuere := `
	CREATE TABLE IF NOT EXIST formulaBibliothek (
		logik VARCHAR(10) NOT NULL,
		formula TEXT NOT NULL,
		ast JSONB NOT NULL,
		valid BOOLEAN NOT NULL,
		conterexample JSONB,

		PRIMARY KEY (logik, formula)
	);
	`

	conn.Exec(ctx, sqlQuere)
}
