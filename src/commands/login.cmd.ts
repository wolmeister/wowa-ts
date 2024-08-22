import { Command } from '@commander-js/extra-typings';
import type { BaseCommand } from './cmd';
import type { UserService } from '../user.service';
import chalk from 'chalk';
import { prompt } from 'enquirer';

export class LoginCommand implements BaseCommand {
	constructor(private userService: UserService) {}

	buildCommand(): Command {
		return new Command('login').description('Login to your wowa account').action(async () => {
			const user = await this.userService.getUser();
			if (user !== null) {
				console.log(`You are already signed in as ${chalk.blueBright(user.email)}`);
				return;
			}

			const response = await prompt<{ email: string }>({
				type: 'input',
				name: 'email',
				message: 'What is your email?',
			});

			await this.userService.sendLoginEmail(response.email);

			const otpResponse = await prompt<{ otp: string }>({
				type: 'input',
				name: 'otp',
				message: `Please enter the code sent to ${response.email}`,
			});

			await this.userService.finishEmailLogin(response.email, otpResponse.otp);
		});
	}
}
