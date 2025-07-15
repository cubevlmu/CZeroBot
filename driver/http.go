package driver

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/tidwall/gjson"
	log "github.com/wdvxdr1123/ZeroBot/log"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

type HTTP struct {
	URL         string
	AccessToken string
	lst         net.Listener
	caller      *HTTPCaller
}

func (h *HTTP) Connect() {
	log.Infof("[httpcaller] 正在尝试与服务器握手: %s", h.caller.URL)
	rsp, err := h.caller.CallAPI(zero.APIRequest{Action: "get_login_info", Params: nil})
	if err != nil {
		log.Warningf("[httpcaller] 与服务器握手失败: %s\n%v", h.caller.URL, err)
		return
	}
	if rsp.RetCode == 0 {
		h.caller.selfID = rsp.Data.Get("user_id").Int()
		zero.APICallers.Store(h.caller.selfID, h.caller) // 添加Caller到 APICaller list...
		log.Infof("[httpcaller] 与服务器 %s 握手成功, 账号: %d", h.caller.URL, h.caller.selfID)
	} else {
		log.Warningf("[httpcaller] 与服务器握手失败: %s", h.caller.URL)
		log.Warningf("[httpcaller] status:%s, retcode:%d, msg:%s, wording:%s", rsp.Status, rsp.RetCode, rsp.Message, rsp.Wording)
	}
}

type HTTPCaller struct {
	URL         string
	AccessToken string
	selfID      int64
}

func NewHTTPClient(url, accessToken, callerURL, callerToken string) *HTTP {
	return &HTTP{
		URL:         url,
		AccessToken: accessToken,
		caller:      &HTTPCaller{URL: callerURL, AccessToken: callerToken},
	}
}

// listen 启动 HTTP 服务器监听
func (h *HTTP) listen() {
	network, address := resolveURI(h.URL)
	uri, err := url.Parse(address)
	if err == nil && uri.Scheme != "" {
		address = uri.Host
	}

	listener, err := net.Listen(network, address)
	if err != nil {
		log.Warningf("[httpsever] server failed to listen at port: %v", err)
		h.lst = nil
		return
	}

	h.lst = listener
	log.Infof("[httpsever] server listening at port %v", listener.Addr())
}

// any 处理所有 API 请求
func (h *HTTP) any(w http.ResponseWriter, r *http.Request, apiHandler func([]byte, zero.APICaller)) {
	if r.Method != http.MethodPost {
		log.Warningf("[httpserver] refused request from %s : invalid request method %s", r.RemoteAddr, r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		log.Warningf("[httpserver] refused request from %s : invalid Content-Type %s", r.RemoteAddr, r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	content, err := io.ReadAll(r.Body)
	if err != nil {
		log.Warningf("[httpserver] refused request from %s : failed to read request body %s", r.RemoteAddr, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if h.AccessToken != "" {
		signatureHeader := r.Header.Get("X-Signature")
		if signatureHeader == "" {
			log.Warningf("[httpserver] refused request from %s : no signature in request", r.RemoteAddr)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		mac := hmac.New(sha1.New, helper.StringToBytes(h.AccessToken))
		mac.Write(content)
		if signatureHeader != "sha1="+hex.EncodeToString(mac.Sum(nil)) {
			log.Warningf("[httpserver] refused request from %s : invalid signature", r.RemoteAddr)
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	apiHandler(content, h.caller)
}

// Listen 监听 HTTP 请求
func (h *HTTP) Listen(handler func([]byte, zero.APICaller)) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		h.any(w, r, handler)
	})

	server := &http.Server{
		Handler: mux,
	}

	for {
		if h.lst == nil {
			time.Sleep(2 * time.Second)
			h.listen()
			continue
		}
		log.Infof("[httpserver] server start handling at : %v", h.lst.Addr())
		err := server.Serve(h.lst)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Warningf("[httpserver] 服务器在端点 %s 失败: %s", h.lst.Addr(), err)
			h.lst = nil
		} else if errors.Is(err, http.ErrServerClosed) {
			log.Info("[httpserver] server closed")
			return
		}
	}
}

// httpCaller 对 api 进行调用
// 不关闭body会导致资源泄漏!
func (c *HTTPCaller) httpCaller(action string, payload []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, c.URL+"/"+action, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	header := req.Header
	header.Set("X-Client-Role", "Universal")
	header.Set("User-Agent", "ZeroBot/1.6.3")

	if c.AccessToken != "" {
		header.Set("Authorization", "Bearer "+c.AccessToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *HTTPCaller) CallAPI(request zero.APIRequest) (zero.APIResponse, error) {
	p, err := json.Marshal(request.Params)
	if err != nil {
		return nullResponse, err
	}

	resp, err := c.httpCaller(request.Action, p)
	if err != nil {
		return nullResponse, err
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nullResponse, err
	}
	payload := helper.BytesToString(content)
	if resp.StatusCode != http.StatusOK {
		return zero.APIResponse{Status: payload, RetCode: int64(1000 + resp.StatusCode)}, fmt.Errorf("caller returned invalid data : %d", resp.StatusCode)
	}
	rsp := gjson.Parse(payload)
	msg := rsp.Get("message").Str
	if msg != "" {
		msg = rsp.Get("msg").Str
	}
	return zero.APIResponse{
		Status:  rsp.Get("status").Str,
		Data:    rsp.Get("data"),
		Message: msg,
		Wording: rsp.Get("wording").Str,
		RetCode: rsp.Get("retcode").Int(),
		Echo:    rsp.Get("echo").Uint(),
	}, nil
}
