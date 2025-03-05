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
	TransferBalanceRequest struct {
		Sender   string `json:"sender"`
		Receiver string `json:"receiver"`
		Amount   int64  `json:"amount"`
	}
)

func (h *UserBalanceHandler) TransferBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := TransferBalanceRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.Response(ctx, w, nil, errs.New(http.StatusBadRequest, err))
		return
	}

	senderDBKey := strings.ToLower(req.Sender)
	receiverDBKey := strings.ToLower(req.Receiver)
	tx := h.db.StartTransaction([]string{senderDBKey, receiverDBKey})
	defer tx.Rollback()

	senderBalance, ok := tx.Get(senderDBKey)
	if !ok {
		utils.Response(ctx, w, nil, errs.New(http.StatusBadRequest, fmt.Errorf("sender not found")))
		return
	}
	receiverBalance, ok := tx.Get(receiverDBKey)
	if !ok {
		utils.Response(ctx, w, nil, errs.New(http.StatusBadRequest, fmt.Errorf("receiver not found")))
		return
	}
	if receiverBalance.UserName == senderBalance.UserName {
		utils.Response(ctx, w, nil, errs.New(http.StatusForbidden, fmt.Errorf("cannot self transfer")))
		return
	}
	finalSenderBalance := senderBalance.Balance - req.Amount
	if finalSenderBalance < 0 {
		utils.Response(ctx, w, nil, errs.New(http.StatusBadRequest, fmt.Errorf("insufficient balance")))
		return
	}
	finalReceiverBalance := receiverBalance.Balance + req.Amount

	err = tx.Set(senderDBKey, model.UserBalance{
		UserName: senderBalance.UserName,
		Balance:  finalSenderBalance,
	})
	if err != nil {
		utils.Response(ctx, w, nil, errs.New(http.StatusInternalServerError, err))
		return
	}
	err = tx.Set(receiverDBKey, model.UserBalance{
		UserName: receiverBalance.UserName,
		Balance:  finalReceiverBalance,
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
