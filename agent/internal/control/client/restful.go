package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

var responseContentType = "application/json"

type ApiConfig struct {
	Url   string
	Token string
}

func (ac *ApiConfig) Valid() (bool, error) {
	if ac.Url == "" {
		return false, errors.New("empty url")
	} else if ac.Token == "" {
		return false, errors.New("empty token")
	}

	return true, nil
}

type RestfulClient struct {
	logger     *zap.Logger
	context    context.Context
	cancel     context.CancelFunc
	httpClient *http.Client
	apiConfig  *ApiConfig
}

func trimUrlSeparatorSuffix(urlPart string) string {
	return strings.TrimSuffix(urlPart, "/")
}

func NewRestfulClient(ctx context.Context, rootLogger *zap.Logger, apiConfig *ApiConfig) (*RestfulClient, error) {
	if valid, err := apiConfig.Valid(); !valid {
		return nil, errors.WithMessage(err, "validate api config")
	}

	logger := rootLogger.Named("restful-client")
	ctx, cancel := context.WithCancel(ctx)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	apiConfig.Url = trimUrlSeparatorSuffix(apiConfig.Url)

	return &RestfulClient{
		logger:     logger,
		context:    ctx,
		cancel:     cancel,
		httpClient: client,
		apiConfig:  apiConfig,
	}, nil
}

func (rc *RestfulClient) Get(endpoint string, message interface{}) (*http.Response, error) {
	return rc.doRequest(http.MethodGet, endpoint, message)
}

func (rc *RestfulClient) Post(endpoint string, message interface{}) (*http.Response, error) {
	return rc.doRequest(http.MethodPost, endpoint, message)
}

func (rc *RestfulClient) Delete(endpoint string, message interface{}) (*http.Response, error) {
	return rc.doRequest(http.MethodDelete, endpoint, message)
}

func (rc *RestfulClient) Put(endpoint string, message interface{}) (*http.Response, error) {
	return rc.doRequest(http.MethodPut, endpoint, message)
}

func (rc *RestfulClient) doRequest(method string, endpoint string, message interface{}) (*http.Response, error) {
	data, err := json.Marshal(message)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal message")
	}

	// todo: Consider to use (carefully!!!) buffer pools for request and response buffers (perhaps use fasthttp).
	// todo: Note that this requires a cautious and well-thought-of design, so we don't end up with a memory leak!
	// todo: Might not be cost-effective considering the relatively small rate of http communication this agent performs.
	requestBody := bytes.NewBuffer(data)

	url := fmt.Sprintf("%s/%s", rc.apiConfig.Url, trimUrlSeparatorSuffix(endpoint))
	req, err := http.NewRequestWithContext(rc.context, method, url, requestBody)
	if err != nil {
		return nil, errors.WithMessage(err, "new request")
	}
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", rc.apiConfig.Token))
	req.Header.Set("Accept-Encoding", "gzip") // Requests aren't gzipped by default.

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return nil, errors.WithMessage(err, "request failed")
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, responseContentType) {
		return nil, errors.Errorf("invalid content type '%s', expected '%s'", contentType, responseContentType)
	}

	return resp, nil
}

func (rc *RestfulClient) AbortAll() {
	rc.httpClient.CloseIdleConnections()
	rc.cancel()
}
