package repositories

// Service is the stuct representing a vcs µService
import (
	"path/filepath"

	gocache "github.com/patrickmn/go-cache"

	"github.com/ovh/cds/engine/api"
	"github.com/ovh/cds/engine/cache"
	"github.com/ovh/cds/engine/service"
	"github.com/ovh/cds/sdk"
)

// Service is the repostories service
type Service struct {
	service.Common
	Cfg        Configuration
	Router     *api.Router
	Cache      cache.Store
	dao        dao
	cacheSize  int64
	localCache *gocache.Cache
}

// Configuration is the vcs configuration structure
type Configuration struct {
	Name                  string                          `toml:"name" comment:"Name of this CDS Repositories Service\n Enter a name to enable this service" json:"name"`
	Basedir               string                          `toml:"basedir" comment:"Root directory where the service will store all checked-out repositories" json:"basedir"`
	OperationRetention    int                             `toml:"operationRetention" comment:"Operation retention in redis store (in days)" default:"5" json:"operationRetention"`
	RepositoriesRetention int                             `toml:"repositoriesRetention" comment:"Re retention on the filesystem (in days)" default:"10" json:"repositoriesRetention"`
	HTTP                  service.HTTPRouterConfiguration `toml:"http" comment:"######################\n CDS Repositories HTTP Configuration \n######################" json:"http"`
	URL                   string                          `default:"http://localhost:8085" json:"url"`
	API                   service.APIServiceConfiguration `toml:"api" comment:"######################\n CDS API Settings \n######################" json:"api"`
	Cache                 struct {
		TTL   int           `toml:"ttl" default:"60" json:"ttl"`
		Redis sdk.RedisConf `toml:"redis" json:"redis"`
	} `toml:"cache" comment:"######################\n CDS Repositories Cache Settings \n######################" json:"cache"`
	MaxWorkers int `toml:"maxWorkers" comment:"Maximum of operations that can be done in parallel" default:"10" json:"maxWorkers"`
}

// Repo retiens a sdk.OperationRepo from an sdk.Operation
func (s Service) Repo(op sdk.Operation) *sdk.OperationRepo {
	r := new(sdk.OperationRepo)
	r.URL = op.URL
	r.Basedir = filepath.Join(s.Cfg.Basedir, r.ID())
	r.RepositoryStrategy = op.RepositoryStrategy
	return r
}
