package utils

import (
	"context"
	"encoding/json"
	"net/http"

	"delman/utils/errs"
)

type Resp struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func Response(ctx context.Context, w http.ResponseWriter, data interface{}, err error) {
	resp := Resp{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    data,
	}
	if err != nil {
		if customError, ok := err.(*errs.Error); ok {
			resp.Code = customError.Code
			resp.Message = customError.Message
		} else {
			resp.Code = http.StatusInternalServerError
			resp.Message = err.Error()
		}
	}
	respJSON, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.Code)
	w.Write(respJSON)
}
