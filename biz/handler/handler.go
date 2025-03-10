package handler

import (
	"delman/biz/model"
	"delman/utils"
)

//type UserBalanceHandler struct {
//	db utils.DatabaseInterface[string, model.UserBalance]
//}

func InitUserBalanceHandler(db utils.DatabaseInterface[string, model.UserBalance]) *UserBalanceHandler {
	return &UserBalanceHandler{
		db: db,
	}
}
