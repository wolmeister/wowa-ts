import fs from 'node:fs/promises';
import path from 'node:path';
import AdmZip from 'adm-zip';
import type { Addon, AddonRepository, GameVersion } from './addon.repository';
import type { ConfigRepository } from './config.repository';
import {
	type CurseClient,
	CurseFileReleaseType,
	type CurseMod,
	SearchModsSortField,
	SearchModsSortOrder,
} from './curse.client';

export type UpdateEvent = { addonId: string; gameVersion: GameVersion } & (
	| { name: 'start' }
	| { name: 'already-up-to-date'; version: string }
	| { name: 'updated'; fromVersion: string; toVersion: string }
	| { name: 'error'; reason?: unknown }
);

export type UpdateListener = (event: UpdateEvent) => void;

export class AddonManager {
	constructor(
		private curseClient: CurseClient,
		private repository: AddonRepository,
		private configRepository: ConfigRepository,
	) {}

	public async installByUrl(url: string, gameVersion: GameVersion): Promise<Addon> {
		const slug = url.replace('https://www.curseforge.com/wow/addons/', '');

		const searchModsResponse = await this.curseClient.searchMods({
			gameId: 1,
			gameVersionTypeId: gameVersion === 'retail' ? 517 : 67408,
			slug: slug,
			index: 0,
			sortField: SearchModsSortField.Popularity,
			sortOrder: SearchModsSortOrder.Desc,
		});

		const curseMod = searchModsResponse.data.find((a) => a.slug === slug);
		if (!curseMod) {
			throw new Error('Addon not found');
		}

		return this.install(curseMod, gameVersion);
	}

	public installById(id: number, gameVersion: GameVersion): void {
		throw new Error('method not implemented');
	}

	public async updateAll(listener?: UpdateListener): Promise<void> {
		// TODO - This method should throw if one of the updates fail?
		const addons = await this.repository.getAll();
		await Promise.all(
			addons.map(async (addon) => {
				listener?.({ addonId: addon.id, gameVersion: addon.gameVersion, name: 'start' });

				try {
					const updatedAddon = await this.installByUrl(addon.id, addon.gameVersion);
					if (updatedAddon.version !== addon.version) {
						listener?.({
							addonId: addon.id,
							gameVersion: addon.gameVersion,
							name: 'updated',
							fromVersion: addon.version,
							toVersion: updatedAddon.version,
						});
						return;
					}

					listener?.({
						addonId: addon.id,
						gameVersion: addon.gameVersion,
						name: 'already-up-to-date',
						version: addon.version,
					});
				} catch (error) {
					listener?.({
						addonId: addon.id,
						gameVersion: addon.gameVersion,
						name: 'error',
						reason: error,
					});
				}
			}),
		);
	}

	async remove(id: string, gameVersion: GameVersion): Promise<Addon | null> {
		const gameFolder = await this.configRepository.get('game.dir');
		if (gameFolder === null) {
			throw new Error('Config game.dir not defined');
		}

		const existingAddon = await this.repository.get(id, gameVersion);
		if (existingAddon === null) {
			return null;
		}

		const versionFolder = gameVersion === 'classic' ? '_classic_era_' : '_retail_';
		const addonsFolder = path.join(gameFolder, `${versionFolder}/Interface/AddOns`);

		await Promise.all(
			existingAddon.directories
				.map((d) => path.join(addonsFolder, d))
				.map(async (d) => {
					await fs.rm(d, { recursive: true, force: true });
				}),
		);

		await this.repository.delete(id, gameVersion);

		return existingAddon;
	}

	private async install(curseMod: CurseMod, gameVersion: GameVersion): Promise<Addon> {
		const gameFolder = await this.configRepository.get('game.dir');
		if (gameFolder === null) {
			throw new Error('Config game.dir not defined');
		}

		const gameVersionTypeId = gameVersion === 'retail' ? 517 : 67408;
		const fileIndex = curseMod.latestFilesIndexes.find(
			(fi) =>
				fi.gameVersionTypeId === gameVersionTypeId &&
				fi.releaseType === CurseFileReleaseType.Release,
		);

		if (!fileIndex) {
			throw new Error('File index not found');
		}

		const modFile = (await this.curseClient.getModFile(curseMod.id, fileIndex.fileId)).data;
		const existingAddon = await this.repository.get(curseMod.slug, gameVersion);
		if (existingAddon && existingAddon.version === modFile.displayName) {
			return existingAddon;
		}

		const response = await fetch(modFile.downloadUrl);
		if (response.body === null) {
			throw new Error('Failed to fetch addon file');
		}

		const versionFolder = gameVersion === 'classic' ? '_classic_era_' : '_retail_';
		const addonsFolder = path.join(gameFolder, `${versionFolder}/Interface/AddOns`);

		if (await fs.exists(addonsFolder)) {
			await Promise.all(
				(existingAddon?.directories ?? [])
					.map((d) => path.join(addonsFolder, d))
					.map(async (d) => {
						await fs.rm(d, { recursive: true, force: true });
					}),
			);
		} else {
			await fs.mkdir(addonsFolder, { recursive: true });
		}

		const arrayBuffer = await response.arrayBuffer();
		const buffer = Buffer.from(arrayBuffer);
		const zip = new AdmZip(buffer);
		for (const entry of zip.getEntries()) {
			if (entry.isDirectory) {
				await fs.mkdir(path.join(addonsFolder, entry.entryName), { recursive: true });
				continue;
			}

			await fs.writeFile(path.join(addonsFolder, entry.entryName), entry.getData());
		}

		const installedAddon: Addon = {
			id: curseMod.slug,
			name: curseMod.name,
			version: modFile.displayName,
			author: curseMod.authors[0]?.name ?? 'N/A',
			gameVersion: gameVersion,
			directories: modFile.modules.map((module) => module.name),
		};
		await this.repository.save(installedAddon);

		return installedAddon;
	}
}
