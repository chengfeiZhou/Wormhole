package types

// Response 接口返回规范
// nolint
type HttpRespData struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}
