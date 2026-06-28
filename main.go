package main

import (
	"fmt"
	"iglp/handler"
	"net/http"
)

func main() {
	http.HandleFunc("/parse", handler.ParseHandler)

	if err := http.ListenAndServe(":9091", nil); err != nil {
		fmt.Println("Ошибка во время работы HTTP сервера:", err)
	}
}
