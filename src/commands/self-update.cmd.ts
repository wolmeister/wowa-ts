import { exists, rename, rm, writeFile } from 'node:fs/promises';
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
          const downloadRes = await got(latestRelease.assets[0].browserDownloadUrl, {
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

          spinner.succeed(`updated wowa to ${latestVersion}`);
        } catch (e) {
          console.error(e);
          process.exit(1);
        }
      });
  }
}
