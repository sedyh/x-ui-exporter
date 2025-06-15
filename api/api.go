package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"x-ui-exporter/config"
	"x-ui-exporter/metrics"

	"github.com/digilolnet/client3xui"
)

// API logic partially was taken from the client3xui module
// https://github.com/digilolnet/client3xui

var (
	cookieCache struct {
		Cookie    http.Cookie
		ExpiresAt time.Time
		sync.Mutex
	}

	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.CLIConfig.InsecureSkipVerify,
			},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 30 * time.Second,
	}

	// Response object pools
	apiResponsePool = sync.Pool{
		New: func() interface{} {
			return &client3xui.ApiResponse{}
		},
	}
	serverStatusPool = sync.Pool{
		New: func() interface{} {
			return &client3xui.ServerStatusResponse{}
		},
	}
	inboundsResponsePool = sync.Pool{
		New: func() interface{} {
			return &client3xui.GetInboundsResponse{}
		},
	}

	// Buffer pool for request bodies
	bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

func GetAuthToken() (*http.Cookie, error) {
	cookieCache.Lock()
	defer cookieCache.Unlock()

	remainingTime := time.Until(cookieCache.ExpiresAt).Minutes()
	if cookieCache.Cookie.Name != "" && remainingTime > 0 {
		log.Printf("Login cookies will expire in %.2f minutes", remainingTime)
		return &cookieCache.Cookie, nil
	}

	path := config.CLIConfig.BaseURL + "/login"
	data := url.Values{
		"username": {config.CLIConfig.ApiUsername},
		"password": {config.CLIConfig.ApiPassword},
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.WriteString(data.Encode())
	defer bufferPool.Put(buf)

	req, err := http.NewRequest("POST", path, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var loginResp struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
	}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, err
	}

	if !loginResp.Success {
		return nil, errors.New(loginResp.Msg)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("authentication failed")
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "3x-ui" {
			cookieCache.Cookie = *cookie
			cookieCache.ExpiresAt = time.Now().Add(time.Minute * 59)
		}
	}

	if cookieCache.Cookie.Name == "" {
		return nil, errors.New("no cookies found in auth response")
	}

	return &cookieCache.Cookie, nil
}

func FetchOnlineUsersCount(cookie *http.Cookie) {
	body, err := sendRequest("/panel/inbound/onlines", http.MethodPost, cookie)
	if err != nil {
		log.Println("Error making request for inbound onlines:", err)
		return
	}

	response := apiResponsePool.Get().(*client3xui.ApiResponse)
	defer apiResponsePool.Put(response)

	if err := json.Unmarshal(body, response); err != nil {
		log.Println("Error unmarshaling response:", err)
		return
	}

	var arr []json.RawMessage
	if err := json.Unmarshal(response.Obj, &arr); err != nil {
		log.Println("Error converting Obj as array:", err)
		return
	}

	metrics.OnlineUsersCount.Set(float64(len(arr)))
}

func FetchServerStatus(cookie *http.Cookie) {
	// Clear old version metric to avoid accumulating obsolete label values
	metrics.XrayVersion.Reset()

	body, err := sendRequest("/server/status", http.MethodPost, cookie)
	if err != nil {
		log.Println("Error making request for system stats:", err)
		return
	}

	response := serverStatusPool.Get().(*client3xui.ServerStatusResponse)
	defer serverStatusPool.Put(response)

	if err := json.Unmarshal(body, response); err != nil {
		log.Println("Error unmarshaling response:", err)
		return
	}

	// XRay metrics
	xrayVersion := strings.ReplaceAll(response.Obj.Xray.Version, ".", "")
	num, err := strconv.ParseFloat(xrayVersion, 64)
	if err != nil {
		log.Println("Error converting xrayVersion:", err)
		metrics.XrayVersion.WithLabelValues(response.Obj.Xray.Version).Set(0)
	} else {
		metrics.XrayVersion.WithLabelValues(response.Obj.Xray.Version).Set(num)
	}

	// Panel metrics
	metrics.PanelThreads.Set(float64(response.Obj.AppStats.Threads))
	metrics.PanelMemory.Set(float64(response.Obj.AppStats.Mem))
	metrics.PanelUptime.Set(float64(response.Obj.AppStats.Uptime))
}

func FetchInboundsList(cookie *http.Cookie) {
	// Clear old metric values to avoid exposing stale data from previous
	// updates. Resetting ensures obsolete label combinations are removed
	// before setting new values.
	metrics.InboundUp.Reset()
	metrics.InboundDown.Reset()
	metrics.ClientUp.Reset()
	metrics.ClientDown.Reset()

	body, err := sendRequest("/panel/api/inbounds/list", http.MethodGet, cookie)
	if err != nil {
		log.Println("Error making request for inbounds list:", err)
		return
	}

	response := inboundsResponsePool.Get().(*client3xui.GetInboundsResponse)
	defer inboundsResponsePool.Put(response)

	if err := json.Unmarshal(body, response); err != nil {
		log.Println("Error unmarshaling response:", err)
		return
	}

	for _, inbound := range response.Obj {
		iid := strconv.Itoa(inbound.ID)
		metrics.InboundUp.WithLabelValues(
			iid, inbound.Remark,
		).Set(float64(inbound.Up))

		metrics.InboundDown.WithLabelValues(
			iid, inbound.Remark,
		).Set(float64(inbound.Down))

		for _, client := range inbound.ClientStats {
			cid := strconv.Itoa(client.ID)
			metrics.ClientUp.WithLabelValues(
				cid, client.Email,
			).Set(float64(client.Up))

			metrics.ClientDown.WithLabelValues(
				cid, client.Email,
			).Set(float64(client.Down))
		}
	}
}

func createRequest(method, path string, cookie *http.Cookie) (*http.Request, error) {
	requestUrl := fmt.Sprintf("%s%s", config.CLIConfig.BaseURL, path)

	req, err := http.NewRequest(method, requestUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.AddCookie(cookie)
	return req, nil
}

func sendRequest(path, method string, cookie *http.Cookie) ([]byte, error) {
	req, err := createRequest(method, path, cookie)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
