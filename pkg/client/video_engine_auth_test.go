package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// A gateway that authenticates /api/netsocs/dh rejects anonymous requests, so
// the video engine calls must carry the driver credentials like every other
// DriverHub call.
func TestVideoEngineCallsSendDriverCredentials(t *testing.T) {
	calls := map[string]func(c *NetsocsDriverClient) error{
		"rtsp_to_stream_id": func(c *NetsocsDriverClient) error {
			_, err := c.RTSPToStreamID("rtsp://cam/1", "stream1")
			return err
		},
		"http_to_stream_id": func(c *NetsocsDriverClient) error {
			_, err := c.HTTPToStreamID("http://cam/1", "stream1")
			return err
		},
		"publish_to_stream_id": func(c *NetsocsDriverClient) error {
			_, err := c.PublishToStreamID("stream1")
			return err
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			var gotAuth, gotToken, gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotAuth, gotToken, gotPath = r.Header.Get("Authorization"), r.Header.Get("X-Auth-Token"), r.URL.Path
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			c := &NetsocsDriverClient{driverKey: "the-driver-key", token: "the-site-token", driverHubHost: srv.URL}
			if err := call(c); err != nil {
				t.Fatalf("call failed: %v", err)
			}
			if gotPath != "/objects/actions/executions/netsocs_native.video_engine/"+name {
				t.Fatalf("unexpected path %q", gotPath)
			}
			if gotAuth != "the-driver-key" {
				t.Errorf("Authorization = %q, want the driver key", gotAuth)
			}
			if gotToken != "the-site-token" {
				t.Errorf("X-Auth-Token = %q, want the site token", gotToken)
			}
		})
	}
}
