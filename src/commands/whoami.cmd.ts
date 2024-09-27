import { Command } from '@commander-js/extra-typings';
import chalk from 'chalk';
import type { UserService } from '../user.service';
import type { BaseCommand } from './cmd';

export class WhoamiCommand implements BaseCommand {
  constructor(private userService: UserService) {}

  buildCommand(): Command {
    return new Command('whoami')
      .description('Display the user email currently logged in')
      .action(async () => {
        const user = await this.userService.getUser();
        if (user === null) {
          console.log(chalk.hex('#FFA500')('null'));
          return;
        }
        console.log(chalk.blueBright(user.email));
      });
  }
}
