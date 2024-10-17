package consul

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol"
	httpError "github.com/handysuherman/clean-arch-payment-service/internal/pkg/http_error"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/serializer"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/tracing"
	"github.com/hashicorp/consul/api"
	"github.com/opentracing/opentracing-go"
)

// ConsulHTTPServiceRequest holds the configuration for making service requests.
type ConsulHTTPServiceRequest struct {
	OperationName   string
	Method          string
	Header          http.Header
	TLS             *tls.Config
	QueryOptions    *api.QueryOptions
	QueryParameters url.Values // Query parameters, if any
	RoutePath       string     // Path with route parameters, e.g., "/users/:userID"
}

// ServiceClient defines the interface for interacting with services.
type HTTPServiceClient interface {
	Request(ctx context.Context, config *ConsulHTTPServiceRequest, body io.Reader) ([]byte, error)
	Fetch(ctx context.Context, config *ConsulHTTPServiceRequest, body []byte) ([]byte, error)
}

// ConsulHTTPService implements ServiceClient using ConsulHTTP for service discovery.
type ConsulHTTPService struct {
	ServiceName string
	balancer    *LeastConnectionsBalancer
	fetchClient *client.Client
}

// NewConsulService creates a new ConsulService instance.
func NewConsulHTTPService(cfg *ConsulServiceConfig, opts ...config.ClientOption) (HTTPServiceClient, error) {
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("service name not specified")
	}

	consulClient, err := New(cfg.ConsulAddr, true, cfg.ConsulCerts)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve client from consul: %v", err)
	}

	if cfg.QueryOptions == nil {
		cfg.QueryOptions = DefaultQueryOptions()
	}

	services, _, err := consulClient.Health().Service(cfg.ServiceName, "", true, cfg.QueryOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to retrive services from consul: %v", err)
	}

	c, err := client.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to set new client: %v", err)
	}

	return &ConsulHTTPService{
		ServiceName: cfg.ServiceName,
		balancer:    NewLeastConnectionsBalancer(services),
		fetchClient: c,
	}, nil
}

// Request makes a request to the service based on the provided configuration.
func (r *ConsulHTTPService) Request(ctx context.Context, config *ConsulHTTPServiceRequest, body io.Reader) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, config.OperationName)
	defer span.Finish()

	// Attempt the request multiple times
	maxRetries := 3 // You can adjust the number of retries
	for attempt := 0; attempt < maxRetries; attempt++ {
		instance := r.balancer.GetNextInstance()
		// defer r.balancer.ReleaseInstance(instance)
		log.Println(instance.Service.ID)

		var serviceURL string
		u := &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", instance.Service.Address, instance.Service.Port),
			Path:   config.RoutePath, // Use RoutePath as the base path
		}

		if config.TLS != nil {
			u.Scheme = "https"
		}

		// Append query parameters to the URL
		if len(config.QueryParameters) > 0 {
			separator := "?"
			if strings.Contains(u.RawQuery, "?") {
				separator = "&"
			}
			u.RawQuery += separator + config.QueryParameters.Encode()
		}

		serviceURL = u.String()

		req, err := http.NewRequestWithContext(ctx, config.Method, serviceURL, body)
		if err != nil {
			return nil, tracing.TraceWithError(span, fmt.Errorf("failed to create HTTP request: %v", err))
		}

		req.Header = config.Header

		client := &http.Client{}

		if config.TLS != nil {
			client.Transport = &http.Transport{
				TLSClientConfig: config.TLS,
			}
		}

		r.balancer.ClearErrorNode(instance.Service.ID)

		resp, err := client.Do(req)
		if err != nil {
			r.balancer.MarkNodeAsError(instance.Service.ID)
			continue // Retry with the next instance
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, tracing.TraceWithError(span, fmt.Errorf("unable to read response body: %v", err))
		}

		// Process the response
		if resp.StatusCode != http.StatusOK {
			var errResponse httpError.RestError
			if err := serializer.Unmarshal(body, &errResponse); err != nil {
				return nil, tracing.TraceWithError(span, fmt.Errorf("unable to read error response body, response status: %v", resp.StatusCode))
			}
			return nil, tracing.TraceWithError(span, fmt.Errorf("response was not successful: %v", errResponse.ErrMessage))
		}

		r.balancer.ReleaseInstance(instance)
		return body, nil // Successful response
	}

	// If all attempts fail, return an error
	return nil, fmt.Errorf("all attempts failed please try again later")

}

func (r *ConsulHTTPService) Fetch(ctx context.Context, config *ConsulHTTPServiceRequest, body []byte) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, config.OperationName)
	defer span.Finish()

	// Attempt the request multiple times
	maxRetries := 3 // You can adjust the number of retries
	for attempt := 0; attempt < maxRetries; attempt++ {
		instance := r.balancer.GetNextInstance()
		// defer r.balancer.ReleaseInstance(instance)
		log.Println(instance.Service.ID)

		var serviceURL string
		u := &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", instance.Service.Address, instance.Service.Port),
			Path:   config.RoutePath, // Use RoutePath as the base path
		}

		if config.TLS != nil {
			u.Scheme = "https"
		}

		// Append query parameters to the URL
		if len(config.QueryParameters) > 0 {
			separator := "?"
			if strings.Contains(u.RawQuery, "?") {
				separator = "&"
			}
			u.RawQuery += separator + config.QueryParameters.Encode()
		}

		serviceURL = u.String()

		req, resp := protocol.AcquireRequest(), protocol.AcquireResponse()
		req.SetRequestURI(serviceURL)
		req.SetMethod(config.Method)

		for key, values := range config.Header {
			for _, value := range values {
				req.SetHeader(key, value)
			}
		}

		r.balancer.ClearErrorNode(instance.Service.ID)

		err := r.fetchClient.Do(ctx, req, resp)
		if err != nil {
			r.balancer.MarkNodeAsError(instance.Service.ID)
			continue // Retry with the next instance
		}

		// Process the response
		if resp.StatusCode() != http.StatusOK {
			var errResponse httpError.RestError
			if err := serializer.Unmarshal(body, &errResponse); err != nil {
				return nil, tracing.TraceWithError(span, fmt.Errorf("unable to read error response body, response status: %v", resp.StatusCode()))
			}
			return nil, tracing.TraceWithError(span, fmt.Errorf("response was not successful: %v", errResponse.ErrMessage))
		}

		r.balancer.ReleaseInstance(instance)
		return resp.Body(), nil
	}

	// If all attempts fail, return an error
	return nil, fmt.Errorf("all attempts failed please try again later")
}
