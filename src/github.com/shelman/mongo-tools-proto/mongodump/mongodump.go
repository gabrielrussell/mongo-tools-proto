package mongodump

import (
	commonopts "github.com/shelman/mongo-tools-proto/common/options"
	"github.com/shelman/mongo-tools-proto/common/db"
)

type MongoDump struct {
	// basic mongo tool options
	ToolOptions *commonopts.ToolOptions

	SessionProvider *db.SessionProvider
}
