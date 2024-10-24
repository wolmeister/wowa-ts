import { Command } from '@commander-js/extra-typings';
import type { AddonRepository } from '../addon.repository';
import type { BaseCommand } from './cmd';

export class ExportCommand implements BaseCommand {
  constructor(private repository: AddonRepository) {}

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
      new Command('export')
        .description('Export all installed addons')
        // .addOption(retailOption)
        // .addOption(classicOption)
        .action(async (url, options) => {
          const addons = await this.repository.getAll();
          const command = addons.map((a) => `wowa add ${a.slug} --${a.gameVersion}`).join(' && ');
          console.log(command);
        })
    );
  }
}
