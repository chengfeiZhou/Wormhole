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

// Unmarshal []]byte => 报文结构
func (msg *HTTPMessage) Unmarshal(b []byte) error {
	return json.Unmarshal(b, msg)
}

// Marshal 报文结构 => []byte
func (msg *HTTPMessage) Marshal() ([]byte, error) {
	return json.Marshal(msg)
}

func TransHTTPMessage(line []byte) (Message, error) {
	data := new(HTTPMessage)
	if err := data.Unmarshal(line); err != nil {
		return nil, err
	}
	return data, nil
}

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
func FilterHost(u *url.URL) string {
	uri := u.String()
	return strings.Replace(uri, u.Host, "%s", 1)
}
