package main

import (
	"delman/biz/handler"
	"delman/biz/model"
)

type Response struct {
	Message string `json:"message"`
}

func main() {
	model.InitDBUserBalance()
	handler.InitUserBalanceHandler(model.DBUserBalance)

	InitRouter()
}
