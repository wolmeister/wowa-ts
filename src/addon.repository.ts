import type { KeyValueStore } from './kv-store';

export type GameVersion = 'retail' | 'classic';

export type Addon = {
	id: string;
	name: string;
	author: string;
	version: string;
	gameVersion: GameVersion;
	directories: string[];
};

export class AddonRepository {
	constructor(private kvStore: KeyValueStore) {}

	save(addon: Addon): Promise<void> {
		return this.kvStore.set(['addons', addon.gameVersion, addon.id], JSON.stringify(addon));
	}

	async delete(id: string, gameVersion: GameVersion): Promise<void> {
		return this.kvStore.set(['addons', gameVersion, id], null);
	}

	async get(id: string, gameVersion: GameVersion): Promise<Addon | null> {
		const addon = await this.kvStore.get(['addons', gameVersion, id]);
		if (addon !== null) {
			return JSON.parse(addon);
		}
		return null;
	}

	async getAll(gameVersion?: GameVersion): Promise<Addon[]> {
		const key = gameVersion ? ['addons', gameVersion] : ['addons'];
		const addons = await this.kvStore.getByPrefix(key);
		return addons.map((a) => JSON.parse(a));
	}
}
