import os from 'node:os';
import path from 'node:path';
import { program } from '@commander-js/extra-typings';
import { version } from '../package.json';
import { AddonManager } from './addon.manager';
import { AddonPrinter } from './addon.printer';
import { AddonRemoteRepository } from './addon.remote-repository';
import { AddonRepository } from './addon.repository';
import { BackupManager } from './backup.manager';
import { BackupCommand } from './commands/backup.cmd';
import { ConfigCommand } from './commands/config.cmd';
import { ExportCommand } from './commands/export.cmd';
import { InstallCommand } from './commands/install.cmd';
import { ListCommand } from './commands/list.cmd';
import { LoginCommand } from './commands/login.cmd';
import { RemoveCommand } from './commands/remove.cmd';
import { SelfUpdateCommand } from './commands/self-update.cmd';
import { SharedMediaSyncCommand } from './commands/shared-media-sync.cmd';
import { UpdateCommand } from './commands/update.cmd';
import { WhoamiCommand } from './commands/whoami.cmd';
import { ConfigRepository } from './config.repository';
import { CurseClient } from './curse.client';
import { KeyValueStore } from './kv-store';
import { KeyValueStoreRepository } from './kv-store.repository';
import { SharedMediaManager } from './shared-media.manager';
import { UserService } from './user.service';

function getKeyValueStorePath(): string {
  const platform = os.platform();
  const homedir = os.homedir();
  switch (platform) {
    case 'darwin':
      return path.join(homedir, 'Library', 'Application Support', 'wowa', 'wowa.json');
    case 'linux':
      return path.join(homedir, '.config', 'wowa', 'wowa.json');
    case 'win32':
      return path.join(homedir, 'AppData', 'Roaming', 'wowa', 'wowa.json');
    default: {
      throw new Error(`Unsupported operating system: ${platform}`);
    }
  }
}

async function main(): Promise<void> {
  const kvStore = new KeyValueStore();
  await kvStore.init(getKeyValueStorePath());

  const configRepository = new ConfigRepository(kvStore);
  const addonRepository = new AddonRepository(new KeyValueStoreRepository(kvStore, 1, []));

  const curseClient = new CurseClient();
  const curseToken = await configRepository.get('curse.token');
  if (curseToken !== null) {
    curseClient.setToken(curseToken);
  }

  if (process.env.API_URL === undefined) {
    console.error('Env variable API_URL not provided');
    process.exit(1);
  }
  const userService = new UserService(configRepository, process.env.API_URL);
  const remoteAddonRepository = new AddonRemoteRepository(userService, process.env.API_URL);

  const addonManager = new AddonManager(
    curseClient,
    userService,
    addonRepository,
    remoteAddonRepository,
    configRepository,
  );
  const addonPrinter = new AddonPrinter();
  const backupManager = new BackupManager(configRepository);
  const sharedMediaManager = new SharedMediaManager(addonRepository, configRepository);

  const installCommand = new InstallCommand(addonManager);
  const updateCommand = new UpdateCommand(addonManager);
  const removeCommand = new RemoveCommand(addonManager);
  const listCommand = new ListCommand(addonRepository, addonPrinter);
  const configCommand = new ConfigCommand(kvStore);
  const backupCommand = new BackupCommand(backupManager);
  const sharedMediaSyncCommand = new SharedMediaSyncCommand(sharedMediaManager);
  const loginCommand = new LoginCommand(userService);
  const whoamiCommand = new WhoamiCommand(userService);
  const selfUpdateCommand = new SelfUpdateCommand();
  const exportCommand = new ExportCommand(addonRepository);

  program.addCommand(installCommand.buildCommand());
  program.addCommand(updateCommand.buildCommand());
  program.addCommand(removeCommand.buildCommand());
  program.addCommand(listCommand.buildCommand());
  program.addCommand(configCommand.buildCommand());
  program.addCommand(backupCommand.buildCommand());
  program.addCommand(sharedMediaSyncCommand.buildCommand());
  program.addCommand(loginCommand.buildCommand());
  program.addCommand(whoamiCommand.buildCommand());
  program.addCommand(selfUpdateCommand.buildCommand());
  program.addCommand(exportCommand.buildCommand());
  program.version(version);
  program.parse();
}

main();
