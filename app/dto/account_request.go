package dto

import "svm_whiteboard/app/model"

type ExecuteAccountRequest struct {
	ProgAddr model.Address `json:"prog_addr"`
	DataAddr model.Address `json:"data_addr"`
	Params   struct {
		Param1 any `json:"param_1"`
		Param2 any `json:"param_2"`
	} `json:"params"`
}

type WriteAccountRequest struct {
	Owner      model.Address `json:"owner"`
	Data       any           `json:"data"`
	Executable bool          `json:"executable"`
}

type CompileCodeRequest struct {
	SourceCode string `json:"source_code"`
}
