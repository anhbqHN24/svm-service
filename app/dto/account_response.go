package dto

import "svm_whiteboard/app/core"

type ExecuteAccountResponse struct {
	ProgAddr    core.Address `json:"prog_addr"`
	DataAddr    core.Address `json:"data_addr"`
	ComputeCost int          `json:"compute_cost"`
	Logs        []string     `json:"logs"`
}
