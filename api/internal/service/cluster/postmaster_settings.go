package cluster

import (
	"context"
	"dashboard/api/internal/infra/logger"
	"dashboard/api/internal/model/cluster"

	"dashboard/api/internal/utils"
	"fmt"
)

const (
	ParamConfigFile           = "config_file"
	ParamDataDirectory        = "data_directory"
	ParamSharedBuffers        = "shared_buffers"
	ParamWalBuffers           = "wal_buffers"
	ParamMaxConnections       = "max_connections"
	ParamHbaFile              = "hba_file"
	ParamWalLevel             = "wal_level"
	ParamAutovacuumMaxWorkers = "autovacuum_max_workers"
)

var postmasterParams = []string{
	ParamConfigFile,
	ParamDataDirectory,
	ParamSharedBuffers,
	ParamWalBuffers,
	ParamMaxConnections,
	ParamHbaFile,
	ParamWalLevel,
	ParamAutovacuumMaxWorkers,
}

// More about pg_settings and its context values
// https://www.postgresql.org/docs/current/view-pg-settings.html
func (s *Service) PostmasterSettings(ctx context.Context) (cluster.PostmasterSettings, error) {

	settings, err := s.storage.PostmasterSettings(ctx, postmasterParams)
	if err != nil {
		s.logger.ErrorContext(ctx, "get cluster settings", "error", err)
		return cluster.PostmasterSettings{}, err
	}

	settingsMap := make(map[string]cluster.Setting, len(settings))

	for _, setting := range settings {
		settingsMap[setting.Name] = setting
	}

	sharedBuffers, found := parseSizeSetting(ctx, s.logger, settingsMap, ParamSharedBuffers)
	if !found {
		s.logger.WarnContext(ctx, "shared_buffers not found in pg_settings or corrupt")
	}

	walBuffers, found := parseSizeSetting(ctx, s.logger, settingsMap, ParamWalBuffers)
	if !found {
		s.logger.WarnContext(ctx, "wal_buffers not found in pg_settings or corrupt")
	}

	return cluster.PostmasterSettings{
		ConfigFile:           settingsMap[ParamConfigFile],
		DataDirectory:        settingsMap[ParamDataDirectory],
		SharedBuffers:        sharedBuffers,
		WalBuffers:           walBuffers,
		MaxConnections:       settingsMap[ParamMaxConnections],
		HbaFile:              settingsMap[ParamHbaFile],
		WalLevel:             settingsMap[ParamWalLevel],
		AutovacuumMaxWorkers: settingsMap[ParamAutovacuumMaxWorkers],
	}, nil
}

func parseSizeSetting(ctx context.Context, logger logger.Logger, store map[string]cluster.Setting, key string) (cluster.Setting, bool) {
	setting, ok := store[key]
	if !ok {
		return cluster.Setting{}, false
	}

	if setting.Unit == "8kB" || setting.Unit == "8KB" {
		n, err := utils.ParseInt64(setting.Value)
		if err != nil {
			logger.WarnContext(ctx, fmt.Sprintf("failed to parse %s", key), "value", setting.Value)
			return cluster.Setting{}, false
		}

		setting.Value = utils.PrettyByteSize((8 * utils.KB) * n)
	}

	return setting, true
}
