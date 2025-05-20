package objects

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// HTTPClient interface for mocking
type HTTPClient interface {
	R() *resty.Request
}

// Request interface for mocking
type Request interface {
	SetHeader(key, value string) *resty.Request
	SetBody(body interface{}) *resty.Request
	Post(url string) (*resty.Response, error)
	Put(url string) (*resty.Response, error)
}

// Response interface for mocking
type Response interface {
	StatusCode() int
	String() string
	Body() []byte
}

// MockHTTPClient is a mock implementation of resty.Client
type MockHTTPClient struct {
	*resty.Client
	transport *mockTransport
}

func NewMockHTTPClient() *MockHTTPClient {
	transport := &mockTransport{}
	client := resty.New()
	client.SetTransport(transport)
	return &MockHTTPClient{
		Client:    client,
		transport: transport,
	}
}

type mockTransport struct {
	statusCode int
	body       []byte
	err        error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}

	// Handle different URL paths
	switch {
	case strings.Contains(req.URL.Path, "/objects/events/types/batch"):
		// This is a batch request
		return &http.Response{
			StatusCode: m.statusCode,
			Body:       io.NopCloser(io.Reader(&mockReadCloser{data: m.body})),
			Header:     make(http.Header),
		}, nil
	case strings.Contains(req.URL.Path, "/objects/events/types/"):
		// This is an individual event type creation request
		// For duplicate entries, return 400
		if strings.Contains(string(m.body), "Duplicate entry") {
			return &http.Response{
				StatusCode: 400,
				Body:       io.NopCloser(io.Reader(&mockReadCloser{data: m.body})),
				Header:     make(http.Header),
			}, nil
		}
		// Otherwise return 200 for successful creation
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(io.Reader(&mockReadCloser{data: []byte{}})),
			Header:     make(http.Header),
		}, nil
	default:
		// Default case, return the configured status code and body
		return &http.Response{
			StatusCode: m.statusCode,
			Body:       io.NopCloser(io.Reader(&mockReadCloser{data: m.body})),
			Header:     make(http.Header),
		}, nil
	}
}

type mockReadCloser struct {
	data []byte
	pos  int
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

// MockRequest is a mock implementation of Request
type MockRequest struct {
	mock.Mock
}

func NewMockRequest() *MockRequest {
	return &MockRequest{}
}

func (m *MockRequest) SetHeader(key, value string) *resty.Request {
	args := m.Called(key, value)
	return args.Get(0).(*resty.Request)
}

func (m *MockRequest) SetBody(body interface{}) *resty.Request {
	args := m.Called(body)
	return args.Get(0).(*resty.Request)
}

func (m *MockRequest) Post(url string) (*resty.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*resty.Response), args.Error(1)
}

func (m *MockRequest) Put(url string) (*resty.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*resty.Response), args.Error(1)
}

// MockResponse is a mock implementation of Response
type MockResponse struct {
	mock.Mock
}

func NewMockResponse() *MockResponse {
	return &MockResponse{}
}

func (m *MockResponse) StatusCode() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockResponse) String() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockResponse) Body() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

func TestAddEventTypes(t *testing.T) {
	tests := []struct {
		name           string
		eventTypes     []EventType
		statusCode     int
		responseBody   []byte
		expectedError  bool
		expectedStatus int
	}{
		{
			name: "successful batch creation",
			eventTypes: []EventType{
				{Domain: "test", EventType: "event1"},
				{Domain: "test", EventType: "event2"},
			},
			statusCode: 201,
			responseBody: func() []byte {
				successBatch := EventTypesBatchResponse{
					Successful: []EventTypeResponse{
						{Domain: "test", EventType: "event1"},
						{Domain: "test", EventType: "event2"},
					},
					Failed: []EventTypeResponse{},
				}
				body, _ := json.Marshal(successBatch)
				return body
			}(),
			expectedError:  false,
			expectedStatus: 201,
		},
		{
			name: "mixed batch results",
			eventTypes: []EventType{
				{Domain: "test", EventType: "event1"},
				{Domain: "test", EventType: "event2"},
				{Domain: "test", EventType: "event3"},
			},
			statusCode: 201,
			responseBody: func() []byte {
				mixedBatch := EventTypesBatchResponse{
					Successful: []EventTypeResponse{
						{Domain: "test", EventType: "event1"},
						{Domain: "test", EventType: "event2"},
					},
					Failed: []EventTypeResponse{
						{
							Domain:             "test",
							EventType:          "event3",
							DisplayDescription: "Invalid event type",
						},
					},
				}
				body, _ := json.Marshal(mixedBatch)
				return body
			}(),
			expectedError:  false,
			expectedStatus: 201,
		},
		{
			name: "all events failed in batch",
			eventTypes: []EventType{
				{Domain: "test", EventType: "event1"},
				{Domain: "test", EventType: "event2"},
			},
			statusCode: 201,
			responseBody: func() []byte {
				failedBatch := EventTypesBatchResponse{
					Successful: []EventTypeResponse{},
					Failed: []EventTypeResponse{
						{
							Domain:             "test",
							EventType:          "event1",
							DisplayDescription: "Invalid domain",
						},
						{
							Domain:             "test",
							EventType:          "event2",
							DisplayDescription: "Invalid event type",
						},
					},
				}
				body, _ := json.Marshal(failedBatch)
				return body
			}(),
			expectedError:  false,
			expectedStatus: 201,
		},
		{
			name: "server error all events failed",
			eventTypes: []EventType{
				{Domain: "test", EventType: "event1"},
				{Domain: "test", EventType: "event2"},
			},
			statusCode:    400,
			responseBody:  []byte("all event types failed to create"),
			expectedError: true,
		},
		{
			name: "fallback to individual creation",
			eventTypes: []EventType{
				{Domain: "test", EventType: "event1"},
			},
			statusCode:    404,
			responseBody:  []byte{},
			expectedError: false,
		},
		{
			name:          "empty event types",
			eventTypes:    []EventType{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &mockTransport{
				statusCode: tt.statusCode,
				body:       tt.responseBody,
			}

			client := resty.New()
			client.SetTransport(transport)

			controller := &objectController{
				driverhub_host: "http://test.com",
				driver_key:     "test-key",
				httpClient:     client,
			}

			err := controller.AddEventTypes(tt.eventTypes)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddEventTypesFallback(t *testing.T) {
	tests := []struct {
		name           string
		eventTypes     []EventType
		statusCode     int
		responseBody   []byte
		expectedError  bool
		expectedStatus int
	}{
		{
			name: "successful individual creation",
			eventTypes: []EventType{
				{Domain: "test", EventType: "event1"},
				{Domain: "test", EventType: "event2"},
			},
			statusCode:    200,
			responseBody:  []byte{},
			expectedError: false,
		},
		{
			name: "duplicate entry handling",
			eventTypes: []EventType{
				{Domain: "test", EventType: "event1"},
			},
			statusCode:    400,
			responseBody:  []byte("Duplicate entry"),
			expectedError: false,
		},
		{
			name:          "empty event types",
			eventTypes:    []EventType{},
			expectedError: true,
		},
		{
			name: "empty event type field",
			eventTypes: []EventType{
				{Domain: "test", EventType: ""},
			},
			expectedError: true,
		},
		{
			name: "empty domain field",
			eventTypes: []EventType{
				{Domain: "", EventType: "event1"},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &mockTransport{
				statusCode: tt.statusCode,
				body:       tt.responseBody,
			}

			client := resty.New()
			client.SetTransport(transport)

			controller := &objectController{
				driverhub_host: "http://test.com",
				driver_key:     "test-key",
				httpClient:     client,
			}

			err := controller.AddEventTypesFallback(tt.eventTypes)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
