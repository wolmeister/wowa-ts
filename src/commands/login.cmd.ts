import { Command } from '@commander-js/extra-typings';
import chalk from 'chalk';
import { prompt } from 'enquirer';
import type { UserService } from '../user.service';
import type { BaseCommand } from './cmd';

export class LoginCommand implements BaseCommand {
  constructor(private userService: UserService) {}

  buildCommand(): Command {
    return new Command('login').description('Login to your wowa account').action(async () => {
      const currentEmail = await this.userService.getUserEmail();
      if (currentEmail !== null) {
        console.log(`You are already signed in as ${chalk.blueBright(currentEmail)}`);
        return;
      }

      const { email } = await prompt<{ email: string }>({
        type: 'input',
        name: 'email',
        message: 'What is your email?',
      });
      const { password } = await prompt<{ password: string }>({
        type: 'password',
        name: 'password',
        message: 'What is your password?',
      });

      await this.userService.signin(email, password);
    });
  }
}
