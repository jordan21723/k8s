package util

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

func CheckHttpExists(url string) bool {
	resp, err := http.Head(url)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

/**
 * nameServer must end with port (example 8.8.8.8:53)
 */
func CommonRequest(requestUrl, httpMethod, nameServer string, postBody json.RawMessage, header map[string]string, skipTlsCheck, disableKeepAlive bool, timeout time.Duration) ([]byte, int, error) {
	var req *http.Request
	var reqErr error

	req, reqErr = http.NewRequest(httpMethod, requestUrl, bytes.NewReader(postBody))
	if reqErr != nil {
		return []byte{}, http.StatusInternalServerError, reqErr
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	for key, val := range header {
		req.Header.Set(key, val)
	}
	client := &http.Client{
		Timeout: timeout,
	}
	tr := &http.Transport{
		DisableKeepAlives: disableKeepAlive,
	}
	if skipTlsCheck {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if nameServer != "" {
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 1 * time.Second}
				return d.DialContext(ctx, "udp", nameServer)
			},
		}
		tr.DialContext = (&net.Dialer{
			Resolver: r,
		}).DialContext
	}
	client.Transport = tr
	resp, respErr := client.Do(req)
	if respErr != nil {
		return []byte{}, http.StatusInternalServerError, respErr
	}
	defer resp.Body.Close()
	body, readBodyErr := ioutil.ReadAll(resp.Body)
	if readBodyErr != nil {
		return []byte{}, http.StatusInternalServerError, readBodyErr
	}
	return body, resp.StatusCode, nil
}

func Download(downloadURL, saveTo string) error {
	// Get the data
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("The url you looking for not found ")
	}

	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(saveTo)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func HttpGet(url string, timeout time.Duration) (*http.Response, error) {
	var (
		cancel func()
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if timeout > 0 {
		timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), timeout)
		req = req.WithContext(timeoutCtx)
		cancel = cancelFunc
	}

	var c = &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	if cancel == nil {
		return res, nil
	}

	res.Body = newWithFuncReadCloser(res.Body, cancel)
	return res, nil
}

func newWithFuncReadCloser(rc io.ReadCloser, f func()) io.ReadCloser {
	return &withFuncReadCloser{
		f:          f,
		ReadCloser: rc,
	}
}

type withFuncReadCloser struct {
	f func()
	io.ReadCloser
}

func (wrc *withFuncReadCloser) Close() error {
	if wrc.f != nil {
		wrc.f()
	}
	return wrc.ReadCloser.Close()
}
