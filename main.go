package main

import (
	"delman/biz/handler"
	"delman/biz/model"
)

type Response struct {
	Message string `json:"message"`
}

func main() {
	db := model.InitDBUserBalance()
	uHandler := handler.InitUserBalanceHandler(db)

	InitRouter(uHandler)
}
