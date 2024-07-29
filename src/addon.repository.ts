import type { KeyValueStoreRepository } from './kv-store.repository';

export type GameVersion = 'retail' | 'classic';

export type AddonDirectory = {
	name: string;
	hash: string | null;
};

export type Addon = {
	id: string;
	name: string;
	author: string;
	version: string;
	gameVersion: GameVersion;
	directories: AddonDirectory[];
	provider: {
		name: 'curse';
		curseModId: number;
	};
	installedAt: string | null;
	updatedAt: string | null;
};

export class AddonRepository {
	constructor(private kvStoreRepository: KeyValueStoreRepository<Addon>) {}

	save(addon: Addon): Promise<void> {
		return this.kvStoreRepository.set(['addons', addon.gameVersion, addon.id], addon);
	}

	async delete(id: string, gameVersion: GameVersion): Promise<void> {
		return this.kvStoreRepository.set(['addons', gameVersion, id], null);
	}

	async get(id: string, gameVersion: GameVersion): Promise<Addon | null> {
		return await this.kvStoreRepository.get(['addons', gameVersion, id]);
	}

	async getAll(gameVersion?: GameVersion): Promise<Addon[]> {
		const key = gameVersion ? ['addons', gameVersion] : ['addons'];
		return await this.kvStoreRepository.getByPrefix(key);
	}
}
