package globals

import (
	"flag"
	"path"

	"github.com/rs/xid"

	db "core.db"
	"github.com/mitchellh/go-homedir"
)

var (
	// HomeDir .
	HomeDir, _ = homedir.Dir()

	// FlagHTTPAddr .
	FlagHTTPAddr = flag.String("http", ":6030", "the http listen address")

	// FlagAllowedHosts .
	FlagAllowedHosts = flag.String("allowed-hosts", "", "the allowed hosts, empty means `all are allowed`")

	// FlagMaxExecTime .
	FlagMaxExecTime = flag.Int64("max-exec-time", 5, "max execution time of each script in seconds")

	// FlagMaxBodySize .
	FlagMaxBodySize = flag.Int64("max-body-size", 2, "max body size in MB")

	// FlagIndexName .
	FlagIndexName = flag.String("index", path.Join(HomeDir, ".aggrex"), "the database index")

	// FlagAdminToken .
	FlagAdminToken = flag.String("admin-token", xid.New().String(), "the admin secret token")
)

var (
	// DBHandler .
	DBHandler *db.DB
)

// PopulateGlobals .
func PopulateGlobals() {
	flag.Parse()
}
