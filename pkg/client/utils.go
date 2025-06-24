package client

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type driverHubVersionResponse struct {
	GitCommitSha string `json:"git_commit_sha"`
	Version      string `json:"version"`
}

type rtsp2StreamIdRequest struct {
	ObjectID []string `json:"object_id"`
	Payload  struct {
		RtspSource string `json:"rtsp_source"`
		StreamID   string `json:"stream_id"`
		Record     bool   `json:"record"`
	} `json:"payload"`
}

type RTSPToStreamIDOpts struct {
	Record         bool
	SourceOnDemand bool `json:"source_on_demand,omitempty"`
}

func (n *NetsocsDriverClient) RTSPToStreamID(rtsp string, streamID string, opts ...RTSPToStreamIDOpts) (videoEngine string, err error) {
	videoEngineDefaultId := "netsocs_native.video_engine.default"
	videoEngineDefaultDomain := "netsocs_native.video_engine"
	req := rtsp2StreamIdRequest{}
	req.ObjectID = []string{videoEngineDefaultId}
	req.Payload.RtspSource = rtsp
	req.Payload.StreamID = streamID

	if len(opts) > 0 {
		req.Payload.Record = opts[0].Record
	}

	resp, err := resty.New().R().SetBody(req).Post(fmt.Sprintf("%s/objects/actions/executions/%s/rtsp_to_stream_id", n.driverHubHost, videoEngineDefaultDomain))

	if err != nil {
		return "", err
	}

	if resp.StatusCode() >= 400 {
		return "", fmt.Errorf("error converting rtsp to stream id: %s", resp.String())
	}

	return videoEngineDefaultId, nil

}
