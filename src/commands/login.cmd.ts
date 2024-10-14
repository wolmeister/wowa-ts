import { Command } from '@commander-js/extra-typings';
import chalk from 'chalk';
import { prompt } from 'enquirer';
import type { UserService } from '../user.service';
import type { BaseCommand } from './cmd';

export class LoginCommand implements BaseCommand {
  constructor(private userService: UserService) {}

  buildCommand(): Command {
    return new Command('login').description('Login to your wowa account').action(async () => {
      const user = await this.userService.getUser();
      if (user !== null) {
        console.log(`You are already signed in as ${chalk.blueBright(user.email)}`);
        return;
      }

      const { type } = await prompt<{ type: 'signin' | 'signup' }>({
        type: 'select',
        name: 'email',
        message: 'Do you want to login to an existing account or create a new one?',
        choices: [
          { name: 'Create a new account', value: 'signin' },
          { name: 'Login to an existing account', value: 'signup' },
        ],
      });

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

      if (type === 'signin') {
        await this.userService.signin(email, password);
      } else {
        await this.userService.signup(email, password);
      }
    });
  }
}
