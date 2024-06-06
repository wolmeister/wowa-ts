import { Command, Option } from '@commander-js/extra-typings';
import ora from 'ora';
import type { AddonManager } from '../addon.manager';
import type { BaseCommand } from './cmd';

export class InstallCommand implements BaseCommand {
	constructor(private addonManager: AddonManager) {}

	buildCommand() {
		const retailOption = new Option(
			'-r, --retail',
			'Install in the retail version of the game',
		).conflicts('retail');
		const classicOption = new Option(
			'-c, --classic',
			'Install in the classic version of the game',
		).conflicts('retail');

		return new Command('add')
			.description('Install an addon')
			.argument('<url>')
			.addOption(retailOption)
			.addOption(classicOption)
			.action(async (url, options) => {
				const gameVersion = options.classic ? 'classic' : 'retail';
				const spinner = ora(`Installing ${url} (${gameVersion})`).start();

				try {
					const addon = await this.addonManager.installByUrl(url, gameVersion);
					spinner.succeed(
						`Installed ${addon.id} ${addon.version} (${addon.gameVersion}) successfully`,
					);
				} catch (error) {
					spinner.fail(`Failed to install ${url} (${gameVersion})`);
					console.error(error);
				}
			});
	}
}
