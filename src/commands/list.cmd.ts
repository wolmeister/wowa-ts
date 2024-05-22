import { Command } from '@commander-js/extra-typings';
import type { BaseCommand } from './cmd';
import type { AddonPrinter } from '../addon.printer';
import type { AddonRepository } from '../addon.repository';

export class ListCommand implements BaseCommand {
	constructor(
		private repository: AddonRepository,
		private addonPrinter: AddonPrinter,
	) {}

	buildCommand(): Command {
		// const retailOption = new Option(
		// 	'-r, --retail',
		// 	'Install in the retail version of the game',
		// ).conflicts('retail');
		// const classicOption = new Option(
		// 	'-c, --classic',
		// 	'Install in the classic version of the game',
		// ).conflicts('retail');

		return (
			new Command('ls')
				.description('List all installed addons')
				// .addOption(retailOption)
				// .addOption(classicOption)
				.action(async (url, options) => {
					const addons = await this.repository.getAll();
					this.addonPrinter.print(addons);
				})
		);
	}
}
