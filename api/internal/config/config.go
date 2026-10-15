package config

import (
	"errors"
	"log"
	"os"

	"github.com/caarlos0/env/v11"
)

type RuntimeVariant string

const (
	LiveRuntime RuntimeVariant = "live"
	DevRuntime  RuntimeVariant = "dev"
	TestRuntime RuntimeVariant = "test"
)

type AppConfig struct {
	Env              EnvironmentVars
	PersistentConfig *PersistentConfigV1
}

type EnvironmentVars struct {
	IsDev      bool
	Port       int            `env:"SERVER_PORT" envDefault:"7999"`
	Runtime    RuntimeVariant `env:"RUNTIME" envDefault:"live"`
	PgHost     string         `env:"PG_HOST"`
	PgPort     int            `env:"PG_PORT"`
	PgUser     string         `env:"PG_USER"`
	PgPassword string         `env:"PG_PASSWORD"`
	PgDatabase string         `env:"PG_DATABASE"`
	PgSSLMode  string         `env:"PG_SSLMODE"`
	Metrics    MetricsVars    `envPrefix:"METRICS_"`
}

func New() AppConfig {

	var envCfg EnvironmentVars

	err := env.Parse(&envCfg)
	if err != nil {
		log.Fatalln("environment variables parse error", err)
	}

	envCfg.IsDev = envCfg.Runtime == DevRuntime

	envCfg.Metrics = envCfg.Metrics.Sanitized()

	persistentCfg, err := configFromEnvs(envCfg)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if persistentCfg != nil {
		return AppConfig{
			Env:              envCfg,
			PersistentConfig: persistentCfg,
		}
	}

	persistentCfg, err = LoadPersistentConfig()
	if err != nil {
		log.Println("volume data for connection not found, await for manual connection")
		_ = RemovePersistentConfig()

		return AppConfig{
			Env:              envCfg,
			PersistentConfig: nil,
		}
	}

	err = ValidatePersistentConfig(persistentCfg)
	if err != nil {
		log.Println("volume /data/config.json is damaged")
		_ = RemovePersistentConfig()
		return AppConfig{
			Env:              envCfg,
			PersistentConfig: nil,
		}
	}

	return AppConfig{
		Env:              envCfg,
		PersistentConfig: persistentCfg,
	}
}

func configFromEnvs(envs EnvironmentVars) (*PersistentConfigV1, error) {

	var foundCount int

	acceptedCount := 5

	persistentCfg := &PersistentConfigV1{
		Version: CurrentVersion,
	}

	if v := envs.PgHost; v != "" {
		persistentCfg.Connection.Host = v
		foundCount++
	}

	if v := envs.PgPort; v > 0 {
		persistentCfg.Connection.Port = v
		foundCount++
	}

	if v := envs.PgUser; v != "" {
		persistentCfg.Connection.User = v
		foundCount++
	}

	if v := envs.PgPassword; v != "" {
		persistentCfg.Connection.Password = v
		foundCount++
	}

	if v := envs.PgDatabase; v != "" {
		persistentCfg.Connection.Database = v
		foundCount++
	}

	// if v := envs.PgSSLMode; v != "" {
	// 	persistentCfg.Connection.SSLMode = v
	// 	foundCount++
	// }

	if foundCount > 0 && foundCount != acceptedCount {
		return nil, errors.New("specify all necessary PG_* envs or remove all for manual authentication")
	}

	if foundCount == acceptedCount {
		return persistentCfg, nil
	}

	return nil, nil
}
