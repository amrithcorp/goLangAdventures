package implementation

import (
	"context"
	"core/core/pkg/grpc/protobuf"
)

func (s *Server) Ping(ctx context.Context, in *protobuf.PingRequest) (*protobuf.PingResponse, error) {
	return &protobuf.PingResponse{
		Acknowledgement: ("Message we got: " + in.Message),
		Response:        "healthy",
	}, nil
}
