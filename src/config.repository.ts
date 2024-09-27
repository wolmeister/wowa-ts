import type { KeyValueStore } from './kv-store';

export type Config = 'curse.token' | 'game.dir';

export class ConfigRepository {
  constructor(private kvStore: KeyValueStore) {}

  get(key: Config): Promise<string | null> {
    return this.kvStore.get(['config', key]);
  }

  set(key: Config, value: string): Promise<void> {
    return this.kvStore.set(['config', key], value);
  }
}
