package cluster

type PostmasterSettings struct {
	ConfigFile           Setting `json:"configFile"`
	DataDirectory        Setting `json:"dataDirectory"`
	SharedBuffers        Setting `json:"sharedBuffers"`
	WalBuffers           Setting `json:"walBuffers"`
	MaxConnections       Setting `json:"maxConnections"`
	HbaFile              Setting `json:"hbaFile"`
	WalLevel             Setting `json:"walLevel"`
	AutovacuumMaxWorkers Setting `json:"autovacuumMaxWorkers"`
}

type Setting struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Unit        string `json:"unit"`
	Description string `json:"description"`
}
