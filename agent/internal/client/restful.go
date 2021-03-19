package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	requestTimeout     = time.Second * 30
	requestContentType = "application/json"
)

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

func (rc *RestfulClient) Get(endpoint string) (*http.Response, error) {
	return rc.sendRequest(http.MethodGet, endpoint, nil)
}

func (rc *RestfulClient) Post(endpoint string, message []byte) (*http.Response, error) {
	return rc.sendRequest(http.MethodPost, endpoint, message)
}

func (rc *RestfulClient) Delete(endpoint string, message []byte) (*http.Response, error) {
	return rc.sendRequest(http.MethodDelete, endpoint, message)
}

func (rc *RestfulClient) Put(endpoint string, message []byte) (*http.Response, error) {
	return rc.sendRequest(http.MethodPut, endpoint, message)
}

func (rc *RestfulClient) sendRequest(method string, endpoint string, message []byte) (*http.Response, error) {
	var err error

	// todo: Consider to use (carefully!!!) buffer pools for request and response buffers (perhaps use fasthttp).
	// todo: Note that this requires a cautious and well-thought-of design, so we don't end up with a memory leak!
	// todo: Might not be cost-effective considering the relatively small rate of http communication this agent performs.
	requestBody := bytes.NewBuffer(message)

	url := fmt.Sprintf("%s/%s/", rc.apiConfig.Url, trimUrlSeparatorSuffix(endpoint))

	requestContext, cancelRequest := context.WithTimeout(rc.context, requestTimeout)
	defer func() {
		if err != nil {
			cancelRequest()
		}
	}()

	request, err := http.NewRequestWithContext(requestContext, method, url, requestBody)
	if err != nil {
		return nil, errors.WithMessage(err, "new request")
	}
	request.Header.Set("Authorization", fmt.Sprintf("Token %s", rc.apiConfig.Token))
	request.Header.Set("Accept-Encoding", "gzip") // Requests aren't gzipped by default.
	request.Header.Set("Content-Type", requestContentType)

	response, err := rc.httpClient.Do(request)
	if err != nil {
		return nil, errors.WithMessage(err, "request failed")
	}

	return response, nil
}

func (rc *RestfulClient) AbortAll() {
	rc.httpClient.CloseIdleConnections()
	rc.cancel()
}
