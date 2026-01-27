package dto

import "svm_whiteboard/app/model"

type ExecuteParam struct {
	Value   any            `json:"value"`
	Type    int            `json:"type"` // 0: primitive, 1: address
	Account *model.Account `json:"-"`    // Loaded account (if Type == 1)
}

type ExecuteAccountRequest struct {
	ProgAddr model.Address `json:"prog_addr"`
	Params   struct {
		Param1 ExecuteParam `json:"param_1"`
		Param2 ExecuteParam `json:"param_2"`
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
