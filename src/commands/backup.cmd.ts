import { Command } from '@commander-js/extra-typings';
import type { Ora } from 'ora';
import ora from 'ora';
import type { AddonManager, UpdateEvent } from '../addon.manager';
import type { BackupManager } from '../backup.manager';
import type { BaseCommand } from './cmd';

export class BackupCommand implements BaseCommand {
  constructor(private backupManager: BackupManager) {}

  private getSpinnerKey(event: UpdateEvent): string {
    return `${event.addonId}-${event.gameVersion}`;
  }

  buildCommand(): Command {
    return new Command('backup').description('Backup the WTF folder').action(async (options) => {
      const spinner = ora('creating backup');
      const version = await this.backupManager.backup();
      spinner.succeed(`backup ${version} created successfully`);
    });
  }
}
