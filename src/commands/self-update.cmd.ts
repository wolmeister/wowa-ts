import type { Command } from '@commander-js/extra-typings';
import type { BaseCommand } from './cmd';

export class SelfUpdateCommand implements BaseCommand {
	buildCommand(): Command {
		throw new Error('Method not implemented.');
	}
}
