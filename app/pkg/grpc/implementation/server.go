package implementation

import (
	"core/core/pkg/grpc/protobuf"

	"github.com/vmihailenco/taskq/v3"
	"gorm.io/gorm"
)

type Server struct {
	protobuf.UnimplementedCoreServer
	Database  gorm.DB
	TaskQueue taskq.Queue
}
