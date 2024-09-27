import { Command } from '@commander-js/extra-typings';
import type { Ora } from 'ora';
import ora from 'ora';
import type { AddonManager, UpdateEvent } from '../addon.manager';
import type { BaseCommand } from './cmd';

export class UpdateCommand implements BaseCommand {
  constructor(private addonManager: AddonManager) {}

  private getSpinnerKey(event: UpdateEvent): string {
    return `${event.addonId}-${event.gameVersion}`;
  }

  buildCommand(): Command {
    return (
      new Command('up')
        .alias('update')
        .description('Update all installed addons')
        // .addOption(retailOption)
        // .addOption(classicOption)
        .action(async (options) => {
          const spinners = new Map<string, Ora>();
          await this.addonManager.updateAll((event) => {
            switch (event.name) {
              case 'start': {
                spinners.set(
                  this.getSpinnerKey(event),
                  ora(`Installing ${event.addonId} (${event.gameVersion})`).start(),
                );
                break;
              }
              case 'updated': {
                spinners
                  .get(this.getSpinnerKey(event))
                  ?.succeed(
                    `${event.addonId} (${event.gameVersion}) updated to ${event.toVersion}`,
                  );
                break;
              }
              case 'already-up-to-date': {
                spinners
                  .get(this.getSpinnerKey(event))
                  ?.info(`${event.addonId} (${event.gameVersion}) is already up to date`);
                break;
              }
              case 'reinstalled': {
                spinners
                  .get(this.getSpinnerKey(event))
                  ?.warn(`${event.addonId} (${event.gameVersion}) reinstalled`);
                break;
              }
              case 'error': {
                spinners
                  .get(this.getSpinnerKey(event))
                  ?.fail(
                    `Failed to update ${event.addonId} (${event.gameVersion}) - ${event.reason}`,
                  );
                break;
              }
            }
          });
        })
    );
  }
}
