package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"delman/biz/model"
	"delman/utils"
	"delman/utils/errs"
	"github.com/gorilla/mux"
)

type (
	UpdateBalanceRequest struct {
		Amount int64 `json:"amount"`
	}
)

func (h *UserBalanceHandler) UpdateBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	userName := vars["user_name"]

	req := UpdateBalanceRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.Response(ctx, w, nil, errs.New(http.StatusBadRequest, err))
		return
	}

	dbKey := strings.ToLower(userName)
	tx := h.db.StartTransaction([]string{dbKey})
	defer tx.Rollback()

	balance, ok := tx.Get(dbKey)
	if !ok {
		utils.Response(ctx, w, nil, errs.New(http.StatusBadRequest, fmt.Errorf("user not found")))
		return
	}

	finalBalance := balance.Balance + req.Amount
	if finalBalance < 0 {
		utils.Response(ctx, w, nil, errs.New(http.StatusBadRequest, fmt.Errorf("insufficient balance")))
		return
	}

	err = tx.Set(dbKey, model.UserBalance{
		UserName: balance.UserName,
		Balance:  finalBalance,
	})
	if err != nil {
		utils.Response(ctx, w, nil, errs.New(http.StatusInternalServerError, err))
		return
	}

	err = tx.Commit()
	if err != nil {
		utils.Response(ctx, w, nil, errs.New(http.StatusInternalServerError, err))
		return
	}

	utils.Response(ctx, w, nil, nil)
	return
}
