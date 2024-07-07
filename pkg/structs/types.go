package types

// Response 接口返回规范
// https://ones.baimaohui.net:24688/wiki/#/team/2hTgeDe2/space/X5zE4Pp3/page/3TsatzhX?a=preview&resource=JUWiprvA
// nolint
type HttpRespData struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}
