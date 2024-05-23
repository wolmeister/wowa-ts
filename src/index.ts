import os from 'node:os';
import path from 'node:path';
import { program } from '@commander-js/extra-typings';
import { AddonManager } from './addon.manager';
import { AddonRepository } from './addon.repository';
import { ConfigCommand } from './commands/config.cmd';
import { InstallCommand } from './commands/install.cmd';
import { ConfigRepository } from './config.repository';
import { CurseClient } from './curse.client';
import { KeyValueStore } from './kv-store';
import { AddonPrinter } from './addon.printer';
import { ListCommand } from './commands/list.cmd';
import { UpdateCommand } from './commands/update.cmd';
import { RemoveCommand } from './commands/remove.cmd';

function getKeyValueStorePath(): string {
	const platform = os.platform();
	const homedir = os.homedir();
	switch (platform) {
		case 'darwin':
			return path.join(homedir, 'Library/Application Support/wowa/wowa.json');
		case 'linux':
			return path.join(homedir, '.config/wowa/wowa.json');
		case 'win32':
			return path.join(homedir, 'AppData/Roaming/wowa/wowa.json');
		default: {
			throw new Error(`Unsupported operating system: ${platform}`);
		}
	}
}

const kvStore = new KeyValueStore();
await kvStore.init(getKeyValueStorePath());

const configRepository = new ConfigRepository(kvStore);
const addonRepository = new AddonRepository(kvStore);

const curseClient = new CurseClient();
const curseToken = await configRepository.get('curse.token');
if (curseToken !== null) {
	curseClient.setToken(curseToken);
}

const addonManager = new AddonManager(curseClient, addonRepository, configRepository);
const addonPrinter = new AddonPrinter();

const installCommand = new InstallCommand(addonManager);
const updateCommand = new UpdateCommand(addonManager);
const removeCommand = new RemoveCommand(addonManager);
const listCommand = new ListCommand(addonRepository, addonPrinter);
const configCommand = new ConfigCommand(kvStore);

program.addCommand(installCommand.buildCommand());
program.addCommand(updateCommand.buildCommand());
program.addCommand(removeCommand.buildCommand());
program.addCommand(listCommand.buildCommand());
program.addCommand(configCommand.buildCommand());
program.parse();
