import { mkdir, readdir, writeFile } from 'node:fs/promises';
import path from 'node:path';
import type { AddonRepository, GameVersion } from './addon.repository';
import type { ConfigRepository } from './config.repository';

const mediaTypes = {
  background: null,
  border: null,
  font: ['.ttf'],
  sound: null,
  statusbar: null,
} as const;

export class SharedMediaManager {
  constructor(
    private addonRepository: AddonRepository,
    private configRepository: ConfigRepository,
  ) {}

  async syncSharedMedia(gameVersion: GameVersion): Promise<'missing-sharedmedia' | 'synchronized'> {
    const sharedMediaAddon = await this.addonRepository.get('sharedmedia', gameVersion);
    if (sharedMediaAddon === null) {
      return 'missing-sharedmedia';
    }

    const addonsFolder = await this.getAddonsFolder(gameVersion);
    const myMediaFolder = path.join(addonsFolder, 'SharedMedia_MyMedia');

    // Make sure the myMedia folder exists
    await mkdir(myMediaFolder, { recursive: true });

    // Create the toc file
    const tocContent = [
      '## Interface: 11503, 40400, 110000, 110002',
      '## Title: SharedMedia_MyMedia',
      '## Dependencies: SharedMedia',
      'MyMedia.lua',
    ].join('\n');
    await writeFile(path.join(myMediaFolder, 'SharedMedia_MyMedia.toc'), tocContent, 'utf-8');

    // Create the MyMedia.lua
    const luaContent: string[] = ['local LSM = LibStub("LibSharedMedia-3.0")'];

    for (const [mediaType, extensions] of Object.entries(mediaTypes)) {
      luaContent.push(...(await this.generateMediaLuaCode(myMediaFolder, mediaType, extensions)));
    }

    await writeFile(path.join(myMediaFolder, 'MyMedia.lua'), luaContent.join('\n'), 'utf-8');

    return 'synchronized';
  }

  private async generateMediaLuaCode(
    myMediaFolder: string,
    mediaType: string,
    extensions: ReadonlyArray<string> | null,
  ): Promise<string[]> {
    const luaContent: string[] = [];

    luaContent.push('\n-- -----');
    luaContent.push(`-- ${mediaType.toUpperCase()}`);
    luaContent.push('-- -----');

    const mediaTypeFolder = path.join(myMediaFolder, mediaType);
    await mkdir(mediaTypeFolder, { recursive: true });

    const files = await readdir(mediaTypeFolder);
    for (const file of files) {
      if (extensions !== null) {
        if (!extensions.some((ext) => file.endsWith(ext))) {
          continue;
        }
      }

      const fileName = path.parse(file).name;
      luaContent.push(
        `LSM:Register("${mediaType}", "${fileName}", [[Interface\\Addons\\SharedMedia_MyMedia\\${mediaType}\\${file}]])`,
      );
    }

    return luaContent;
  }

  private async getAddonsFolder(gameVersion: GameVersion): Promise<string> {
    const gameFolder = await this.configRepository.get('game.dir');
    if (gameFolder === null) {
      throw new Error('Config game.dir not defined');
    }

    const versionFolder = gameVersion === 'classic' ? '_classic_era_' : '_retail_';
    const addonsFolder = path.join(gameFolder, versionFolder, 'Interface', 'AddOns');

    return addonsFolder;
  }
}
