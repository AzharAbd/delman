package model

import (
	"strings"

	"delman/utils"
)

type UserBalance struct {
	UserName string
	Balance  int64
}

func InitDBUserBalance() utils.DatabaseInterface[string, UserBalance] {
	DBUserBalance := utils.NewDatabase[string, UserBalance]()
	DBUserBalance.Set(strings.ToLower("Mark"), UserBalance{
		UserName: "Mark",
		Balance:  100,
	})
	DBUserBalance.Set(strings.ToLower("Jane"), UserBalance{
		UserName: "Jane",
		Balance:  30,
	})
	DBUserBalance.Set(strings.ToLower("Adam"), UserBalance{
		UserName: "Adam",
		Balance:  0,
	})

	return DBUserBalance
}
