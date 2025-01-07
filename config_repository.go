package main

type Config string

const (
	CurseToken Config = "curse.token"
	GameDir    Config = "game.dir"
	AuthToken  Config = "auth.token"
)

type ConfigRepository struct {
	kvStore *KeyValueStore
}

func NewConfigRepository(kvStore *KeyValueStore) *ConfigRepository {
	return &ConfigRepository{kvStore: kvStore}
}

func (cr *ConfigRepository) Get(key Config) (string, error) {
	return cr.kvStore.Get([]string{"config", string(key)})
}

func (cr *ConfigRepository) Set(key Config, value *string) error {
	return cr.kvStore.Set([]string{"config", string(key)}, value)
}
