import { Command, Option } from '@commander-js/extra-typings';
import ora from 'ora';
import type { AddonManager } from '../addon.manager';
import type { BaseCommand } from './cmd';

export class RemoveCommand implements BaseCommand {
  constructor(private addonManager: AddonManager) {}

  buildCommand() {
    const retailOption = new Option(
      '-r, --retail',
      'Remove from the retail version of the game',
    ).conflicts('retail');
    const classicOption = new Option(
      '-c, --classic',
      'Remove from the classic version of the game',
    ).conflicts('retail');

    return new Command('rm')
      .description('Uninstall an addon')
      .argument('<id>')
      .addOption(retailOption)
      .addOption(classicOption)
      .action(async (id, options) => {
        const gameVersion = options.classic ? 'classic' : 'retail';
        const spinner = ora(`Removing ${id} (${gameVersion})`).start();

        try {
          const addon = await this.addonManager.remove(id, gameVersion);
          if (addon !== null) {
            spinner.succeed(
              `Removed ${addon.id} ${addon.version} (${addon.gameVersion}) successfully`,
            );
          } else {
            spinner.warn(`${id} (${gameVersion}) not found`);
          }
        } catch (error) {
          spinner.fail(`Failed to remove ${id} (${gameVersion})`);
        }
      });
  }
}
