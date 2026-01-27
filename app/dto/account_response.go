package dto

import "svm_whiteboard/app/model"

type ExecuteAccountResponse struct {
	ProgAddr    model.Address `json:"prog_addr"`
	ComputeCost int           `json:"compute_cost"`
	Logs        []string      `json:"logs"`
}
