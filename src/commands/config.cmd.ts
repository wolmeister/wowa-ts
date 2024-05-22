import { Command } from '@commander-js/extra-typings';
import type { KeyValueStore } from '../kv-store';
import type { BaseCommand } from './cmd';

export class ConfigCommand implements BaseCommand {
	constructor(private kvStore: KeyValueStore) {}

	buildCommand() {
		return new Command('config')
			.description('Manage configuration')
			.argument('<key>', 'Configuration key')
			.argument('[value]', 'Configuration value')
			.action(async (key, value) => {
				if (value === undefined) {
					const currentValue = await this.kvStore.get(['config', key]);
					console.log(currentValue ?? 'null');
					return;
				}

				await this.kvStore.set(['config', key], value);
			});
	}
}
