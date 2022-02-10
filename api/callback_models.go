package api

type CallbackFunctionId int

const (
	StopMonitoringCallback CallbackFunctionId = iota
	UpdateCallback
)

type CallbackPayload struct {
	FunctionId CallbackFunctionId
	Data       interface{}
}
