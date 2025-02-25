package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

type HTTPClient struct {
	client *http.Client
}

type RequestParams struct {
	URL     string
	Headers map[string]string
	Query   map[string]string
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{},
	}
}

func (c *HTTPClient) doRequest(params RequestParams, method string, body io.Reader) (*http.Response, error) {
	requestURL, err := url.Parse(params.URL)
	if err != nil {
		return nil, err
	}

	// Set query params
	query := requestURL.Query()
	for key, value := range params.Query {
		query.Add(key, value)
	}
	requestURL.RawQuery = query.Encode()

	// Create the request
	req, err := http.NewRequest(method, requestURL.String(), body)
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range params.Headers {
		req.Header.Set(key, value)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	// Check response status code
	if res.StatusCode != http.StatusOK {
		_ = res.Body.Close()
		return nil, errors.New(res.Status)
	}

	return res, nil
}

func (c *HTTPClient) Get(params RequestParams, target interface{}) error {
	res, err := c.doRequest(params, "GET", nil)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	// Decode response into target struct
	err = json.NewDecoder(res.Body).Decode(target)
	if err != nil {
		return err
	}

	return nil
}

func (c *HTTPClient) GetBytes(params RequestParams) ([]byte, error) {
	res, err := c.doRequest(params, "GET", nil)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	// Read response bytes
	bytesRes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return bytesRes, nil
}

func (c *HTTPClient) Post(params RequestParams, rawBody interface{}, target interface{}) error {
	body, err := json.Marshal(rawBody)
	if err != nil {
		return err
	}

	res, err := c.doRequest(params, "POST", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	//Decode response into target struct
	err = json.NewDecoder(res.Body).Decode(target)
	if err != nil {
		return err
	}

	return nil
}
