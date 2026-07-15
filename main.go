package main

import (
	"iglp/handler"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/parse", handler.ParseHandler)

	server := &http.Server{
		Addr:              ":9091",
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Println("Ошибка во время работы HTTP сервера:", err)
	}
}
