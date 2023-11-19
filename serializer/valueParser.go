package serializer

import "errors"

type ActionAllRespValue struct {
	Status []bool `json:"status"`
}

func ActionAllRespParse(raw []byte) (*ActionAllRespValue, error) {
	var resp ActionAllRespValue = ActionAllRespValue{
		Status: make([]bool, 15),
	}
	for i, v := range raw {
		if v == 0x1 {
			resp.Status[i] = true
		}
		if v == 0x2 {
			resp.Status[i] = false
		}
	}
	if len(raw) == 0 {
		return nil, errors.New("ActionAllRespParse input:raw empty")
	}
	return &resp, nil
}
