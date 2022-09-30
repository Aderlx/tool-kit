package toolkit_http

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	DefaultReadTimeout  = 10 * time.Second
	DefaultWriteTimeout = 10 * time.Second
	MaxConnTimeout      = 60 * time.Second
	MaxConnsPerHost     = 2000
)

type HttpClient struct {
	*fasthttp.Client
	//config *HttpClientConfig
	Cookie  []*fasthttp.Cookie
	Headers map[string]string
}

type HttpClientConfig struct {
	Name                          *string
	NoDefaultUserAgentHeader      *bool
	Dial                          *fasthttp.DialFunc
	DialDualStack                 *bool
	TLSConfig                     *tls.Config
	MaxConnsPerHost               *int
	MaxIdleConnDuration           *time.Duration
	MaxConnDuration               *time.Duration
	MaxIdemponentCallAttempts     *int
	ReadBufferSize                *int
	WriteBufferSize               *int
	ReadTimeout                   *time.Duration
	WriteTimeout                  *time.Duration
	MaxResponseBodySize           *int
	DisableHeaderNamesNormalizing *bool
	DisablePathNormalizing        *bool
	MaxConnWaitTimeout            *time.Duration
	RetryIf                       *fasthttp.RetryIfFunc
	ConfigureClient               *func(hc *fasthttp.HostClient) error
	mLock                         *sync.Mutex
	m                             *map[string]*fasthttp.HostClient
	ms                            *map[string]*fasthttp.HostClient
	readerPool                    *sync.Pool
	writerPool                    *sync.Pool
}

func NewFastClient(client *fasthttp.Client) *HttpClient {
	if client != nil {
		return &HttpClient{Client: client}
	}
	var httpClient = &HttpClient{Client: &fasthttp.Client{

		NoDefaultUserAgentHeader: true,
		TLSConfig:                &tls.Config{InsecureSkipVerify: true},
		MaxConnsPerHost:          MaxConnsPerHost,
		MaxIdleConnDuration:      MaxConnTimeout,
		MaxConnDuration:          MaxConnTimeout,
		ReadTimeout:              DefaultReadTimeout,
		WriteTimeout:             DefaultWriteTimeout,
		MaxConnWaitTimeout:       MaxConnTimeout,
	}}

	return httpClient

}
func toValues(args map[string]string) string {
	var values string

	if args == nil {
		return values
	}
	for k, v := range args {
		if k[:2] != "__" {
			v = url.QueryEscape(v)
		} else {
			k = k[2:len(k)]
		}
		values += fmt.Sprintf("%s=%s&", k, v)
	}

	return values[:len(values)-1]
}

func (c *HttpClient) Get(urlString string, param map[string]string) ([]byte, error) {

	var (
		request  = fasthttp.AcquireRequest()
		response = fasthttp.AcquireResponse()
	)
	defer func() {
		// 用完需要释放资源
		fasthttp.ReleaseRequest(request)
		fasthttp.ReleaseResponse(response)
	}()
	var paramValues = toValues(param)
	request.SetRequestURI(fmt.Sprintf("%s?%s", urlString, paramValues))
	request.Header.SetMethod(fasthttp.MethodGet)
	for k, v := range c.Headers {
		request.Header.Set(k, v)
	}
	if c.Cookie != nil {
		for _, cookie := range c.Cookie {
			request.Header.SetCookieBytesKV(cookie.Key(), cookie.Value())
		}
	}
	if err := c.Do(request, response); err != nil {
		return nil, err
	}
	response.Header.VisitAllCookie(func(key, value []byte) {
		var cookie = fasthttp.AcquireCookie()
		cookie.ParseBytes(value)
		c.Cookie = append(c.Cookie, cookie)
	})
	return response.Body(), nil
}

func (c *HttpClient) Post(url string, param map[string]interface{}) ([]byte, error) {
	var (
		request  = fasthttp.AcquireRequest()
		response = fasthttp.AcquireResponse()
	)
	defer func() {
		// 用完需要释放资源
		fasthttp.ReleaseRequest(request)
		fasthttp.ReleaseResponse(response)
	}()
	request.SetRequestURI(url)
	var bodyJson, _ = json.Marshal(param)

	request.Header.SetMethod(fasthttp.MethodPost)
	request.Header.SetContentTypeBytes([]byte("application/json"))
	for k, v := range c.Headers {
		request.Header.Set(k, v)
	}
	if c.Cookie != nil {
		for _, cookie := range c.Cookie {
			request.Header.SetCookieBytesKV(cookie.Key(), cookie.Value())
		}
	}
	request.SetBodyRaw(bodyJson)
	if err := c.Do(request, response); err != nil {
		return nil, err
	}
	response.Header.VisitAllCookie(func(key, value []byte) {
		var cookie = fasthttp.AcquireCookie()
		if err := cookie.ParseBytes(value); err != nil {
			return
		}
		c.Cookie = append(c.Cookie, cookie)
	})

	return response.Body(), nil
}

func (c *HttpClient) CleanCookie() {
	for _, cookie := range c.Cookie {
		fasthttp.ReleaseCookie(cookie)
	}
	c.Cookie = nil
}

func (c *HttpClient) GetCookie() []*fasthttp.Cookie {
	return c.Cookie
}

func (c *HttpClient) SetRequestHeaders(headers map[string]string) {
	c.Headers = headers
}

func (c *HttpClient) SetCookie(cookieValue map[string]string) {
	for k, v := range cookieValue {
		var cookie = fasthttp.AcquireCookie()
		cookie.SetKey(k)
		cookie.SetValue(v)
		c.Cookie = append(c.Cookie, cookie)
	}
}

func (c *HttpClient) SetClientReadTimeout(duration time.Duration) {
	c.Client.ReadTimeout = duration
}

func (c *HttpClient) SetClientWriteTimeout(duration time.Duration) {
	c.Client.WriteTimeout = duration
}

func (c *HttpClient) SetSkipVerifyTLSConfig(isSkip bool) {
	c.Client.TLSConfig = &tls.Config{
		InsecureSkipVerify: isSkip,
	}
}
