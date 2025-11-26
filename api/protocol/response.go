package protocol

// Response : api 回應格式
type Response struct {
	Code    int    `json:"Code"`
	Message string `json:"Message"`
	Result  any    `json:"Result"`
}
