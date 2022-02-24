package api

type Response struct {
	ErrMsg string
	Data   interface{} // !REVIEW Usages
}

func (r *Response) HasError() bool {
	return r.ErrMsg != ""
}
