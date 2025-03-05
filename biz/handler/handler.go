package handler

import (
	"delman/biz/model"
	"delman/utils"
)

var Handler *UserBalanceHandler

type UserBalanceHandler struct {
	db utils.DatabaseInterface[string, model.UserBalance]
}

func InitUserBalanceHandler(db utils.DatabaseInterface[string, model.UserBalance]) {
	Handler = &UserBalanceHandler{
		db: db,
	}
}
