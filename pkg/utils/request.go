package utils

import (
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

var (
	Getenv  = os.Getenv
	BaseUrl = Getenv("LIVE_URL")
)

var defaultHeaders = map[string]string{
	"Host":            BaseUrl,
	"Referer":         BaseUrl,
	"User-Agent":      "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
	"Accept":          "application/json, text/plain, */*",
	"Accept-Language": "en-US,en;q=0.5",
	"Accept-Encoding": "gzip, deflate, br",
	"Connection":      "keep-alive",
	"Pragma":          "no-cache",
	"Cache-Control":   "no-cache",
	"TE":              "Trailers",
}

var insecureHTTPClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Disable TLS certificate verification
		},
	},
}

func Request[T any](url, method string, body io.Reader, headers ...map[string]string) (T, error) {
	req, _ := http.NewRequest(method, url, body)
	for k, v := range defaultHeaders {
		req.Header.Add(k, v)
	}
	if len(headers) > 0 {
		for k, v := range headers[0] {
			req.Header.Add(k, v)
		}
	}
	var authenticateResponse T
	res, err := insecureHTTPClient.Do(req)
	if err != nil {
		return authenticateResponse, err
	}
	defer res.Body.Close()
	reader, err := gzip.NewReader(res.Body)
	if err != nil {
		return authenticateResponse, err
	}
	bodyByte, err := io.ReadAll(reader)
	if err != nil {
		return authenticateResponse, err
	}
	err = json.Unmarshal(bodyByte, &authenticateResponse)
	return authenticateResponse, err
}

func GetIndex(val, defVal, comp, assign int32) int32 {
	var index = defVal
	if val%10 >= comp {
		index = assign
	}
	return index
}

func GetSize(size int) int {
	if size == 0 {
		size = 500
	}
	return size
}
