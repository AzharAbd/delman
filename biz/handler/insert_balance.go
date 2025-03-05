package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"delman/biz/model"
	"delman/utils"
	"delman/utils/errs"
)

type (
	InsertBalanceRequest struct {
		Name    string `json:"name"`
		Balance int64  `json:"balance"`
	}
)

func (h *UserBalanceHandler) InsertBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := InsertBalanceRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.Response(ctx, w, nil, errs.New(http.StatusBadRequest, err))
		return
	}

	dbKey := strings.ToLower(req.Name)
	tx := h.db.StartTransaction([]string{dbKey})
	defer tx.Rollback()

	_, ok := tx.Get(dbKey)
	if ok {
		utils.Response(ctx, w, nil, errs.New(http.StatusBadRequest, fmt.Errorf("user already exists")))
		return
	}
	err = tx.Set(dbKey, model.UserBalance{
		UserName: req.Name,
		Balance:  req.Balance,
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
