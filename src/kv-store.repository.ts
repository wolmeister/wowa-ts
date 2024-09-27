import type { KeyValueStore } from './kv-store';

export type RepositoryMigration<T = unknown, R = unknown> = (item: T) => R;

export type RepositoryItem<T> = {
  version: number;
  value: T;
};

export class KeyValueStoreRepository<T> {
  constructor(
    private kvStore: KeyValueStore,
    private version: number,
    private migrations: RepositoryMigration[],
  ) {}

  async get(key: string[]): Promise<T | null> {
    const rawItem = await this.kvStore.get(key);
    if (rawItem == null) {
      return null;
    }

    return this.parseAndMigrateItem(rawItem);
  }

  async getByPrefix(prefix: string[]): Promise<T[]> {
    const rawItems = await this.kvStore.getByPrefix(prefix);
    return Promise.all(rawItems.map((i) => this.parseAndMigrateItem(i)));
  }

  private async parseAndMigrateItem(rawItem: string): Promise<T> {
    const parsedItem = JSON.parse(rawItem);
    let repositoryItem: RepositoryItem<T>;

    // We need this to support pre migrations items.
    if (Object.keys(parsedItem).length === 2 && 'version' in parsedItem && 'value' in parsedItem) {
      repositoryItem = parsedItem;
    } else {
      repositoryItem = {
        value: parsedItem,
        version: 1,
      };
    }

    if (repositoryItem.version > this.version) {
      throw new Error('Cannot downgrade the item version');
    }

    if (repositoryItem.version === this.version) {
      return repositoryItem.value;
    }

    const neededMigrations = this.migrations.slice(repositoryItem.version - 1, this.version);
    return neededMigrations.reduce((item, migrate) => {
      return migrate(item) as T;
    }, repositoryItem.value);
  }

  async set(key: string[], value: T | null): Promise<void> {
    if (value === null) {
      return this.kvStore.set(key, null);
    }

    return this.kvStore.set(
      key,
      JSON.stringify(<RepositoryItem<T>>{
        version: this.version,
        value,
      }),
    );
  }
}
