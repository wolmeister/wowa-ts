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
	| { name: 'reinstalled'; version: string }
	| { name: 'updated'; fromVersion: string; toVersion: string }
	| { name: 'error'; reason?: unknown }
);

export type UpdateListener = (event: UpdateEvent) => void;

export type InstalLResult = {
	addon: Addon;
	status: 'already-installed' | 'installed' | 'reinstalled' | 'updated';
};

export class AddonManager {
	constructor(
		private curseClient: CurseClient,
		private repository: AddonRepository,
		private configRepository: ConfigRepository,
	) {}

	public async installByUrl(url: string, gameVersion: GameVersion): Promise<InstalLResult> {
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

		const fireEvent = (event: UpdateEvent): void => {
			if (listener !== undefined) {
				setImmediate(() => {
					listener(event);
				});
			}
		};

		await Promise.all(
			addons.map(async (addon) => {
				fireEvent({ addonId: addon.id, gameVersion: addon.gameVersion, name: 'start' });

				try {
					const installResult = await this.installByUrl(addon.id, addon.gameVersion);
					if (installResult.status === 'updated') {
						fireEvent({
							addonId: addon.id,
							gameVersion: addon.gameVersion,
							name: 'updated',
							fromVersion: addon.version,
							toVersion: installResult.addon.version,
						});
						return;
					}
					if (installResult.status === 'reinstalled') {
						fireEvent({
							addonId: addon.id,
							gameVersion: addon.gameVersion,
							name: 'reinstalled',
							version: installResult.addon.version,
						});
						return;
					}

					fireEvent({
						addonId: addon.id,
						gameVersion: addon.gameVersion,
						name: 'already-up-to-date',
						version: addon.version,
					});
				} catch (error) {
					fireEvent({
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
				.map((d) => path.join(addonsFolder, d.name))
				.map(async (d) => {
					await fs.rm(d, { recursive: true, force: true });
				}),
		);

		await this.repository.delete(id, gameVersion);

		return existingAddon;
	}

	private async install(curseMod: CurseMod, gameVersion: GameVersion): Promise<InstalLResult> {
		const gameVersionTypeId = gameVersion === 'retail' ? 517 : 67408;
		const fileIndex = curseMod.latestFilesIndexes.find(
			(fi) =>
				fi.gameVersionTypeId === gameVersionTypeId &&
				fi.releaseType === CurseFileReleaseType.Release,
		);

		if (!fileIndex) {
			throw new Error('File index not found');
		}

		const addonsFolder = await this.getAddonsFolder(gameVersion);

		const modFile = (await this.curseClient.getModFile(curseMod.id, fileIndex.fileId)).data;
		const existingAddon = await this.repository.get(curseMod.slug, gameVersion);
		if (existingAddon && existingAddon.version === modFile.displayName) {
			if (await this.isAddonInstallationValid(existingAddon)) {
				return {
					addon: existingAddon,
					status: 'already-installed',
				};
			}
		}

		const response = await fetch(modFile.downloadUrl);
		if (response.body === null) {
			throw new Error('Failed to fetch addon file');
		}

		if (await fs.exists(addonsFolder)) {
			await Promise.all(
				(existingAddon?.directories ?? [])
					.map((d) => path.join(addonsFolder, d.name))
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
		const zipEntries = zip.getEntries();

		// Some zip files will not contain all directory entries.
		// So we need to create these folders manually.
		// We don't care about empty folders, so we can safely skip them.
		const directories = new Set<string>();
		for (const entry of zipEntries) {
			if (entry.isDirectory) {
				continue;
			}
			directories.add(path.dirname(path.join(addonsFolder, entry.entryName)));
		}

		// Create all directories first.
		// Then later we can write all files at once.
		await Promise.all(
			Array.from(directories).map(async (dir) => {
				await fs.mkdir(dir, { recursive: true });
			}),
		);

		// Write all files.
		await Promise.all(
			zipEntries
				.filter((entry) => !entry.isDirectory)
				.map(async (entry) => {
					await fs.writeFile(path.join(addonsFolder, entry.entryName), entry.getData());
				}),
		);

		const installedAddon: Addon = {
			id: curseMod.slug,
			name: curseMod.name,
			version: modFile.displayName,
			author: curseMod.authors[0]?.name ?? 'N/A',
			gameVersion: gameVersion,
			directories: await Promise.all(
				modFile.modules.map(async (module) => ({
					name: module.name,
					hash: null,
				})),
			),
			provider: {
				name: 'curse',
				curseModId: curseMod.id,
			},
			installedAt: existingAddon?.installedAt ?? new Date().toISOString(),
			updatedAt: existingAddon === null ? null : new Date().toISOString(),
		};
		await this.repository.save(installedAddon);

		return {
			addon: installedAddon,
			status:
				existingAddon !== null && existingAddon.version === modFile.displayName
					? 'reinstalled'
					: existingAddon !== null
						? 'updated'
						: 'installed',
		};
	}

	private async isAddonInstallationValid(addon: Addon): Promise<boolean> {
		const addonsFolder = await this.getAddonsFolder(addon.gameVersion);
		const validDirectories = await Promise.all(
			addon.directories.map(async (addonDirectory) => {
				const modulePath = path.join(addonsFolder, addonDirectory.name);
				return fs.exists(modulePath);
			}),
		);
		return validDirectories.every((valid) => valid === true);
	}

	private async getAddonsFolder(gameVersion: GameVersion): Promise<string> {
		const gameFolder = await this.configRepository.get('game.dir');
		if (gameFolder === null) {
			throw new Error('Config game.dir not defined');
		}

		const versionFolder = gameVersion === 'classic' ? '_classic_era_' : '_retail_';
		const addonsFolder = path.join(gameFolder, `${versionFolder}/Interface/AddOns`);

		return addonsFolder;
	}
}
