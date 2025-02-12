package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type driverHubVersionResponse struct {
	GitCommitSha string `json:"git_commit_sha"`
	Version      string `json:"version"`
}

func (n *NetsocsDriverClient) checkVersion() error {
	resp, err := http.Get(n.buildURL("/api/v1/version"))

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	data := driverHubVersionResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	if data.Version == "" {
		return fmt.Errorf("driver hub version must be greater than 2.0.0")
	}

	versionMajor := data.Version[0]

	if versionMajor < '2' {
		return fmt.Errorf("driver hub version must be greater than 2.0.0")
	}

	return nil
}

type rtsp2StreamIdRequest struct {
	ObjectID []string `json:"object_id"`
	Payload  struct {
		RtspSource string `json:"rtsp_source"`
		StreamID   string `json:"stream_id"`
	} `json:"payload"`
}

func (n *NetsocsDriverClient) RTSPToStreamID(rtsp string, streamID string) (videoEngine string, err error) {
	videoEngineDefaultId := "netsocs_native.video_engine.default"
	videoEngineDefaultDomain := "netsocs_native.video_engine"
	req := rtsp2StreamIdRequest{}
	req.ObjectID = []string{videoEngineDefaultId}
	req.Payload.RtspSource = rtsp
	req.Payload.StreamID = streamID

	resp, err := resty.New().R().SetBody(req).Post(fmt.Sprintf("%s/objects/actions/executions/%s/rtsp_to_stream_id", n.driverHubHost, videoEngineDefaultDomain))

	if err != nil {
		return "", err
	}

	if resp.StatusCode() >= 400 {
		return "", fmt.Errorf("error converting rtsp to stream id: %s", resp.String())
	}

	return videoEngineDefaultId, nil

}
