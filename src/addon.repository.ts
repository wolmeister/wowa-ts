import type { KeyValueStoreRepository } from './kv-store.repository';

export type GameVersion = 'retail' | 'classic';

export type LocalAddon = {
  id: string;
  name: string;
  slug: string;
  author: string;
  version: string;
  gameVersion: GameVersion;
  directories: string[];
  provider: 'curse' | 'github';
  providerId: string;
  updatedAt: string;
};

export class AddonRepository {
  constructor(private kvStoreRepository: KeyValueStoreRepository<LocalAddon>) {}

  save(addon: LocalAddon): Promise<void> {
    return this.kvStoreRepository.set(['local-addons', addon.gameVersion, addon.id], addon);
  }

  async delete(id: string, gameVersion: GameVersion): Promise<void> {
    return this.kvStoreRepository.set(['local-addons', gameVersion, id], null);
  }

  async get(id: string, gameVersion: GameVersion): Promise<LocalAddon | null> {
    return await this.kvStoreRepository.get(['local-addons', gameVersion, id]);
  }

  async getAll(gameVersion?: GameVersion): Promise<LocalAddon[]> {
    const key = gameVersion ? ['local-addons', gameVersion] : ['local-addons'];
    return await this.kvStoreRepository.getByPrefix(key);
  }
}
