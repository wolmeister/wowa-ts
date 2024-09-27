import fs from 'node:fs/promises';
import path from 'node:path';
import AdmZip from 'adm-zip';
import type { GameVersion } from './addon.repository';
import type { ConfigRepository } from './config.repository';

export class BackupManager {
  constructor(private configRepository: ConfigRepository) {}

  private async getWtfFolder(gameVersion: GameVersion): Promise<string> {
    const gameFolder = await this.configRepository.get('game.dir');
    if (gameFolder === null) {
      throw new Error('Config game.dir not defined');
    }

    const versionFolder = gameVersion === 'classic' ? '_classic_era_' : '_retail_';
    const wtfFolder = path.join(gameFolder, `${versionFolder}/WTF`);

    return wtfFolder;
  }

  private async getWtfBackupFolder(gameVersion: GameVersion): Promise<string> {
    const gameFolder = await this.configRepository.get('game.dir');
    if (gameFolder === null) {
      throw new Error('Config game.dir not defined');
    }

    const versionFolder = gameVersion === 'classic' ? '_classic_era_' : '_retail_';
    const wtfFolder = path.join(gameFolder, `${versionFolder}/WTF_backup`);

    return wtfFolder;
  }

  async backup(): Promise<string> {
    const wtfFolder = await this.getWtfFolder('retail');
    const wtfBackupFolder = await this.getWtfBackupFolder('retail');

    if ((await fs.exists(wtfBackupFolder)) === false) {
      await fs.mkdir(wtfBackupFolder, { recursive: true });
    }

    const now = new Date();
    const backupName = `${now.getFullYear()}-${now.getMonth() + 1}-${now.getDate()}.zip`;
    const backupPath = path.join(wtfBackupFolder, backupName);

    const zip = new AdmZip();
    zip.addLocalFolder(wtfFolder, path.basename(wtfFolder));
    await zip.writeZipPromise(backupPath);

    return backupName;
  }
}
