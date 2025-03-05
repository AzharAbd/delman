package handler

import (
	"fmt"
	"net/http"
	"strings"

	"delman/utils"
	"delman/utils/errs"
	"github.com/gorilla/mux"
)

type (
	GetBalanceRequest struct {
		Name string `json:"name"`
	}
	GetBalanceResponse struct {
		Name    string `json:"name"`
		Balance int64  `json:"balance"`
	}
)

func (h *UserBalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	userName := vars["user_name"]

	balance, ok := h.db.Get(strings.ToLower(userName))
	if !ok {
		utils.Response(ctx, w, nil, errs.New(http.StatusBadRequest, fmt.Errorf("user not found")))
		return
	}

	resp := GetBalanceResponse{
		Name:    balance.UserName,
		Balance: balance.Balance,
	}
	utils.Response(ctx, w, resp, nil)
	return
}
