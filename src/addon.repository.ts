import type { KeyValueStoreRepository } from './kv-store.repository';

// [document version (1 byte)]
// [id length (1 byte)] [id bytes]
// [name length (1 byte)] [name bytes]
// [slug length (1 byte)] [slug bytes]
// [author length (1 byte)] [author bytes]
// [version length (1 byte)] [version bytes]
// [gameVersion (1 byte)]
// [directories length (1 byte)]
//   [directory 1 length (1 byte)] [directory 1 bytes]
//   [directory 2 length (1 byte)] [directory 2 bytes]
//   ...
// [provider type (1 byte)]
// [providerId length (1 byte)] [providerId bytes]
// [updatedAt length (1 byte)] [updatedAt bytes]

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
