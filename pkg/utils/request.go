package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Request struct {
	Client *http.Client
	Agreement string
	Domain string
	Headers map[string]string
}

func NewRequest(c http.Client, agreement string, domain string, headers map[string]string) *Request {
	r := &Request{
		Client: &c,
		Agreement: agreement,
		Domain: domain,
		Headers: headers,
	}

	return r
}

func (r *Request) Get(url string, data map[string]interface{}) ([]byte, error) {

	buffer := &bytes.Buffer{}
	reqUrl := fmt.Sprintf("%s://%s%s?", r.Agreement, r.Domain, url)
	buffer.WriteString(reqUrl)

	if len(data) > 0{
		for _, value := range data{
			buffer.WriteString(value.(string))
			buffer.WriteString("&")
		}
	}

	req, err := http.NewRequest("GET", buffer.String(), nil)

	if err != nil {
		return nil, err
	}

	for key, val := range r.Headers {
		req.Header.Set(key, val)
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (r *Request) Post(url string, data map[string]interface{}) ([]byte, error) {
	bytesData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	reqUrl := fmt.Sprintf("%s://%s%s", r.Agreement, r.Domain, url)

	req, err := http.NewRequest("POST", reqUrl, bytes.NewReader(bytesData))

	if err != nil {
		return nil, err
	}

	for key, val := range r.Headers {
		req.Header.Set(key, val)
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}