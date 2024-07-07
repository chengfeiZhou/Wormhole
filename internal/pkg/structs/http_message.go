package structs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// HTTPMessage 请求报文结构体
// nolint
type HTTPMessage struct {
	Method string              `json:"method"`
	URL    string              `json:"url"` // 被去除了HOST的 scheme://userinfo@/{host}/path?query#fragment
	Header map[string][]string `json:"header"`
	Body   string              `json:"body"`

	RemoteAddr       string              `json:"remoteAddr"`
	Proto            string              `json:"proto"`         // "HTTP/1.0"
	ContentLength    int64               `json:"contentLength"` // TODO: 如果报错就设置成-1
	TransferEncoding []string            `json:"transferEncoding"`
	Form             map[string][]string `json:"form"`
	PostForm         map[string][]string `json:"postForm"`
	Trailer          map[string][]string `json:"trailer"`
	// MultipartForm    *multipart.Form     `json:"multipartForm"` // TODO: 文件解析会有内容的问题
}

// Unmarshal 是一个HTTPMessage类型的方法，用于将JSON格式的字节数组解析到HTTPMessage结构体中
//
// 参数：
//
//	b []byte：待解析的JSON格式的字节数组
//
// 返回值：
//
//	error：如果解析成功，返回nil；否则返回错误信息
func (msg *HTTPMessage) Unmarshal(b []byte) error {
	return json.Unmarshal(b, msg)
}

// Marshal 函数将HTTPMessage结构体对象序列化为JSON格式的字节切片
//
// 参数：
//   - msg: 指向待序列化的HTTPMessage结构体的指针
//
// 返回值：
//   - []byte: 序列化后的JSON格式字节切片
//   - error: 如果序列化过程中发生错误，则返回非零的错误码；否则返回nil
func (msg *HTTPMessage) Marshal() ([]byte, error) {
	return json.Marshal(msg)
}

// TransHTTPMessage 将给定的字节切片转换为HTTPMessage类型
// 参数：
//
//	line []byte - 待转换的字节切片
//
// 返回值：
//
//	Message - 转换后的HTTPMessage类型指针，若转换失败则为nil
//	error - 如果转换过程中发生错误，则返回非零的错误码；否则返回nil
func TransHTTPMessage(line []byte) (Message, error) {
	data := new(HTTPMessage)
	if err := data.Unmarshal(line); err != nil {
		return nil, err
	}
	return data, nil
}

// MakeRequest 创建一个HTTP请求
//
// 参数：
//
//	msg *HTTPMessage - 指向HTTPMessage结构体的指针，包含HTTP请求的详细信息
//	host string - 请求的主机名
//
// 返回值：
//
//	*http.Request - 创建的HTTP请求对象
//	error - 如果在创建请求过程中发生错误，则返回非零的错误码
//
// 该函数根据给定的HTTPMessage结构体中的信息，创建一个HTTP请求。
// 使用给定的主机名（host）与HTTPMessage中的URL拼接成完整的URL，
// 使用HTTPMessage中的Method字段作为请求方法，
// 使用HTTPMessage中的Body字段作为请求体，
// 使用HTTPMessage中的Header字段作为请求头。
// 如果在创建请求过程中发生错误，则返回nil的请求对象和非零的错误码。
func (msg *HTTPMessage) MakeRequest(host string) (*http.Request, error) {
	body := bytes.NewReader([]byte(msg.Body))
	_url := fmt.Sprintf("http://"+msg.URL, host)
	req, err := http.NewRequest(msg.Method, _url, body)
	if err != nil {
		return nil, err
	}
	for key, values := range msg.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return req, nil
}

// ParseRequest 将请求中的报文解析成结构
// ParseRequest 是一个函数，它解析传入的HTTP请求r，并返回一个HTTPMessage指针和一个错误。
// 如果解析过程中出现错误，将返回nil的HTTPMessage指针和错误对象。
// HTTPMessage是一个自定义结构体，用于封装HTTP请求的相关信息。
func ParseRequest(r *http.Request) (*HTTPMessage, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	message := &HTTPMessage{
		Method:           r.Method,
		URL:              FilterHost(r.URL),
		Proto:            r.Proto,
		Header:           r.Header,
		Body:             string(body),
		ContentLength:    r.ContentLength,
		TransferEncoding: r.TransferEncoding,
		Form:             r.Form,
		PostForm:         r.PostForm,
		// MultipartForm:    r.MultipartForm,
		Trailer:    r.Trailer,
		RemoteAddr: r.RemoteAddr,
	}
	return message, nil
}

// FilterHost 过滤掉URL中host并以%s填充
// scheme://userinfo@/host/path?query#fragment ==> scheme://userinfo@/%s/path?query#fragment
// FilterHost 函数接收一个 *url.URL 类型的参数 u，返回一个 string 类型的值
// 该函数将 u.String() 返回的 URL 字符串中的 Host 部分替换为 "%s"
// 替换操作只执行一次
func FilterHost(u *url.URL) string {
	uri := u.String()
	return strings.Replace(uri, u.Host, "%s", 1)
}
