import { exists, mkdir, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { setImmediate } from 'node:timers/promises';
import { Command, Option } from '@commander-js/extra-typings';
import ora from 'ora';
import type { AddonRepository } from '../addon.repository';
import type { ConfigRepository } from '../config.repository';
import type { SharedMediaManager } from '../shared-media.manager';
import type { BaseCommand } from './cmd';

export class SharedMediaSyncCommand implements BaseCommand {
	constructor(private sharedMediaManager: SharedMediaManager) {}

	buildCommand() {
		const retailOption = new Option(
			'-r, --retail',
			'Remove from the retail version of the game',
		).conflicts('retail');
		const classicOption = new Option(
			'-c, --classic',
			'Remove from the classic version of the game',
		).conflicts('retail');

		return new Command('shared-media-sync')
			.alias('sms')
			.description('Syncs custom shared media using the SharedMedia addon')
			.addOption(retailOption)
			.addOption(classicOption)
			.action(async (options) => {
				const gameVersion = options.classic ? 'classic' : 'retail';
				const spinner = ora('Syncing custom media').start();

				// We need this because otherwise the ora spinner will break,
				// Don't ask me why.
				await setImmediate();

				try {
					const result = await this.sharedMediaManager.syncSharedMedia(gameVersion);
					if (result === 'missing-sharedmedia') {
						spinner.warn('SharedMedia is not installed. Please run wowa add sharedmedia.');
						return;
					}
					spinner.succeed('Custom shared media synchronized successfully');
				} catch (error) {
					spinner.fail('Failed to sync custom media');
				}
			});
	}
}
