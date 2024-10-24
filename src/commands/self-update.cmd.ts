import { chmod, exists, rename, rm, writeFile } from 'node:fs/promises';
import os from 'node:os';
import { Command } from '@commander-js/extra-typings';
import got from 'got';
import ora from 'ora';
import { lte } from 'semver';
import { version } from '../../package.json';
import { getLatestGithubRelease } from '../github.client';
import type { BaseCommand } from './cmd';

export class SelfUpdateCommand implements BaseCommand {
  buildCommand(): Command {
    return new Command('self-update')
      .alias('su')
      .description('Check for new wowa updates')
      .action(async () => {
        try {
          const spinner = ora('checking for update').start();
          const latestRelease = await getLatestGithubRelease('wolmeister', 'wowa-ts');
          const latestVersion = latestRelease.tagName.substring(1);
          if (lte(latestVersion, version)) {
            spinner.info('wowa is already up to date');
            return;
          }

          spinner.text = 'downloading latest version';
          // switch (platform) {
          //   case 'darwin':
          //     return path.join(homedir, 'Library/Application Support/wowa/wowa.json');
          //   case 'linux':
          //     return path.join(homedir, '.config/wowa/wowa.json');
          //   case 'win32':
          //     return path.join(homedir, 'AppData/Roaming/wowa/wowa.json');
          //   default: {
          //     throw new Error(`Unsupported operating system: ${platform}`);
          //   }
          // }
          const platform = os.platform();
          const asset = latestRelease.assets.find((a) => {
            switch (platform) {
              case 'win32':
                return a.name.includes('win64');
              case 'linux':
                return a.name.includes('linux64');
              default:
                throw new Error(`Unsupported operating system: ${platform}`);
            }
          });
          if (asset === undefined) {
            spinner.fail('release asset not found');
            process.exit(1);
          }
          const downloadRes = await got(asset.browserDownloadUrl, {
            responseType: 'buffer',
          });
          if (!downloadRes.ok) {
            spinner.fail('download failed');
            process.exit(1);
          }
          const buffer = downloadRes.body;
          console.log(`downloaded ${buffer.byteLength} bytes`);

          if (await exists(`${process.execPath}.backup`)) {
            await rm(`${process.execPath}.backup`);
          }

          await rename(process.execPath, `${process.execPath}.backup`);
          await writeFile(process.execPath, buffer);
          await chmod(process.execPath, 0o755);

          spinner.succeed(`updated wowa to ${latestVersion}`);
        } catch (e) {
          console.error(e);
          process.exit(1);
        }
      });
  }
}
