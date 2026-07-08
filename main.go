package main

import (
	"context"
	"fmt"
	"iglp/database"
	"iglp/handler"
	"net/http"
)

func main() {
	ctx := context.Background()
	conn, err := database.CreateConnection(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	database.CreateTable(ctx, conn)

	http.HandleFunc("/parse", handler.ParseHandler)

	if err := http.ListenAndServe(":9091", nil); err != nil {
		fmt.Println("Ошибка во время работы HTTP сервера:", err)
	}
}
