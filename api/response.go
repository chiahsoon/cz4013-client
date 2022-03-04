package api

type Response struct {
	ErrMsg string
	Data   interface{}
}

func (r *Response) HasError() bool {
	return r.ErrMsg != ""
}
