import { randomUUID } from 'node:crypto';
import fs from 'node:fs/promises';
import path from 'node:path';
import type { SupabaseClient } from '@supabase/supabase-js';
import AdmZip from 'adm-zip';
import got from 'got';
import type { AddonRepository, GameVersion, LocalAddon } from './addon.repository';
import type { ConfigRepository } from './config.repository';
import {
  type CurseClient,
  CurseFileReleaseType,
  type CurseMod,
  SearchModsSortField,
  SearchModsSortOrder,
} from './curse.client';
import type { Database } from './supabase.db.types';
import type { UserService } from './user.service';

export type UpdateEvent = { addonId: string; gameVersion: GameVersion } & (
  | { name: 'start' }
  | { name: 'already-up-to-date'; version: string }
  | { name: 'reinstalled'; version: string }
  | { name: 'updated'; fromVersion: string; toVersion: string }
  | { name: 'error'; reason?: unknown }
);

export type UpdateListener = (event: UpdateEvent) => void;

export type InstalLResult = {
  addon: LocalAddon;
  status: 'already-installed' | 'installed' | 'reinstalled' | 'updated';
};

export class AddonManager {
  constructor(
    private curseClient: CurseClient,
    private supabase: SupabaseClient<Database>,
    private userService: UserService,
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
    const addons = await this.supabase.from('addons').select();
    if (addons.error !== null) {
      throw addons.error;
    }

    const fireEvent = (event: UpdateEvent): void => {
      if (listener !== undefined) {
        setImmediate(() => {
          listener(event);
        });
      }
    };

    await Promise.all(
      addons.data.map(async (addon) => {
        fireEvent({ addonId: addon.slug, gameVersion: addon.game_version, name: 'start' });

        try {
          const installResult = await this.installByUrl(addon.slug, addon.game_version);
          switch (installResult.status) {
            case 'installed': {
              fireEvent({
                addonId: addon.slug,
                gameVersion: addon.game_version,
                name: 'updated',
                fromVersion: 'TODO',
                toVersion: installResult.addon.version,
              });
              break;
            }
            case 'updated': {
              fireEvent({
                addonId: addon.slug,
                gameVersion: addon.game_version,
                name: 'updated',
                fromVersion: 'TODO',
                toVersion: installResult.addon.version,
              });
              break;
            }
            case 'reinstalled': {
              fireEvent({
                addonId: addon.slug,
                gameVersion: addon.game_version,
                name: 'reinstalled',
                version: installResult.addon.version,
              });
              break;
            }
            case 'already-installed': {
              fireEvent({
                addonId: addon.slug,
                gameVersion: addon.game_version,
                name: 'already-up-to-date',
                version: installResult.addon.version,
              });
              break;
            }
          }
        } catch (error) {
          fireEvent({
            addonId: addon.slug,
            gameVersion: addon.game_version,
            name: 'error',
            reason: error,
          });
        }
      }),
    );
  }

  async remove(id: string, gameVersion: GameVersion): Promise<LocalAddon | null> {
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

  private async install(curseMod: CurseMod, gameVersion: GameVersion): Promise<InstalLResult> {
    const user = await this.userService.getUser();
    if (user == null) {
      throw new Error('User not signed in');
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

    const response = await got(modFile.downloadUrl, {
      responseType: 'buffer',
    });

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

    const buffer = response.body;
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

    const remoteAddon = await this.supabase
      .from('addons')
      .select()
      .eq('game_version', gameVersion)
      .eq('slug', curseMod.slug);
    if (remoteAddon.error !== null) {
      throw remoteAddon.error;
    }
    if (remoteAddon.data.length > 1) {
      throw new Error('More than one addon found?');
    }

    if (remoteAddon.data.length === 0) {
      const result = await this.supabase.from('addons').insert({
        id: randomUUID(),
        user_id: user.id,
        slug: curseMod.slug,
        game_version: gameVersion,
        author: curseMod.authors[0]?.name ?? 'N/A',
        name: curseMod.name,
        provider: 'curse',
        provider_id: String(curseMod.id),
        url: `https://www.curseforge.com/wow/addons/${curseMod.slug}`,
      });
      if (result.error) {
        console.log(result.error);
        throw result.error;
      }
    }

    const installedAddon: LocalAddon = {
      id: curseMod.slug,
      slug: curseMod.slug,
      name: curseMod.name,
      version: modFile.displayName,
      author: curseMod.authors[0]?.name ?? 'N/A',
      gameVersion: gameVersion,
      directories: modFile.modules.map((module) => module.name),
      provider: 'curse',
      providerId: String(curseMod.id),
      updatedAt: new Date().toISOString(),
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

  private async isAddonInstallationValid(addon: LocalAddon): Promise<boolean> {
    const addonsFolder = await this.getAddonsFolder(addon.gameVersion);
    const validDirectories = await Promise.all(
      addon.directories.map(async (addonDirectory) => {
        const modulePath = path.join(addonsFolder, addonDirectory);
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
