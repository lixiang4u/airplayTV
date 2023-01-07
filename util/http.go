package util

import (
	"compress/flate"
	"compress/gzip"
	"github.com/andybalholm/brotli"
	"github.com/zc310/headers"
	"io"
	"net/http"
	"strings"
)

type HttpWrapper struct {
	headers map[string]string
}

//	x.httpWrapper.SetHeader("origin", czHost)
//	x.httpWrapper.SetHeader("authority", util.HandleHostname(czHost))
//	x.httpWrapper.SetHeader("referer", czHost)
//	x.httpWrapper.SetHeader("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36")
//	x.httpWrapper.SetHeader("cookie", "")

func (x *HttpWrapper) SetHeader(k, v string) {
	if x.headers == nil {
		x.headers = make(map[string]string)
	}
	x.headers[k] = v
}

func (x *HttpWrapper) SetHeaders(h map[string]string) {
	x.headers = h
}

func (x *HttpWrapper) GetHeaders() map[string]string {
	return x.headers
}

func (x *HttpWrapper) addHeaderParams(req *http.Request) {
	for k, v := range x.headers {
		req.Header.Set(k, v)
	}
}

// 解码返回的编码数据，需要根据response头的Content-Encoding确定
func (x *HttpWrapper) decodeEncoding(resp *http.Response) ([]byte, error) {
	switch resp.Header.Get(headers.ContentEncoding) {
	case "br":
		return io.ReadAll(brotli.NewReader(resp.Body))
	case "gzip":
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		return io.ReadAll(gr)
	case "deflate":
		zr := flate.NewReader(resp.Body)
		defer func() { _ = zr.Close() }()
		return io.ReadAll(zr)
	default:
		return io.ReadAll(resp.Body)
	}
}

func (x *HttpWrapper) Get(requestUrl string) ([]byte, error) {
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}
	x.addHeaderParams(req)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}

	return x.decodeEncoding(resp)
}

func (x *HttpWrapper) Post(requestUrl, rawBody string) ([]byte, error) {
	req, err := http.NewRequest("POST", requestUrl, strings.NewReader(rawBody))
	if err != nil {
		return nil, err
	}
	x.addHeaderParams(req)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}

	return x.decodeEncoding(resp)
}

func (x *HttpWrapper) GetResponse(requestUrl string) (map[string][]string, []byte, error) {
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, nil, err
	}
	x.addHeaderParams(req)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, nil, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return resp.Header, b, nil
}

func (x *HttpWrapper) PostResponse(requestUrl, rawBody string) (map[string][]string, []byte, error) {
	req, err := http.NewRequest("POST", requestUrl, strings.NewReader(rawBody))
	if err != nil {
		return nil, nil, err
	}
	x.addHeaderParams(req)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, nil, err
	}

	b, err := x.decodeEncoding(resp)

	return resp.Header, b, err
}
