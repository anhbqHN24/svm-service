package dto

import "svm_whiteboard/app/core"

type ExecuteAccountRequest struct {
	ProgAddr core.Address `json:"prog_addr"`
	DataAddr core.Address `json:"data_addr"`
	Params   struct {
		Param1 any `json:"param_1"`
		Param2 any `json:"param_2"`
	} `json:"params"`
}

type WriteAccountRequest struct {
	Owner      core.Address `json:"owner"`
	Data       any          `json:"data"`
	Executable bool         `json:"executable"`
}
