package client

import (
	"bytes"
	"context"
	"crypto/tls"
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

type RestfulClient struct {
	logger     *zap.Logger
	context    context.Context
	cancel     context.CancelFunc
	httpClient *http.Client
	apiConfig  *ApiConfig
}

func NewRestfulClient(ctx context.Context, rootLogger *zap.Logger, apiConfig *ApiConfig) *RestfulClient {
	logger := rootLogger.Named("restful-client")
	ctx, cancel := context.WithCancel(ctx)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return &RestfulClient{
		logger:     logger,
		context:    ctx,
		cancel:     cancel,
		httpClient: client,
		apiConfig:  apiConfig,
	}
}

func (rc *RestfulClient) Get(data []byte) (*http.Response, error) {
	return rc.doRequest(http.MethodGet, data)
}

func (rc *RestfulClient) Post(data []byte) (*http.Response, error) {
	return rc.doRequest(http.MethodPost, data)
}

func (rc *RestfulClient) Delete(data []byte) (*http.Response, error) {
	return rc.doRequest(http.MethodDelete, data)
}

func (rc *RestfulClient) doRequest(method string, data []byte) (*http.Response, error) {
	// todo: Consider to use (carefully!!!) buffer pools for request and response buffers (perhaps use fasthttp).
	// todo: Note that this requires a cautious and well-thought-of design, so we don't end up with a memory leak!
	// todo: Might not be cost-effective considering the relatively small rate of http communication this agent performs.
	requestBody := bytes.NewBuffer(data)

	req, err := http.NewRequestWithContext(rc.context, http.MethodPost, rc.apiConfig.Url, requestBody)
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
