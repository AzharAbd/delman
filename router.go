package main

import (
	"log"
	"net/http"

	"delman/biz/handler"
	"github.com/gorilla/mux"
)

func InitRouter() {
	r := mux.NewRouter()
	r.HandleFunc("/balance/transfer", handler.Handler.TransferBalance).Methods("POST")
	r.HandleFunc("/balance/{user_name}", handler.Handler.GetBalance).Methods("GET")
	r.HandleFunc("/balance/{user_name}", handler.Handler.UpdateBalance).Methods("PUT")
	r.HandleFunc("/balance", handler.Handler.InsertBalance).Methods("POST")

	http.Handle("/", r)
	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
