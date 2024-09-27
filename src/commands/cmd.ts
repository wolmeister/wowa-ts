import type { Command } from '@commander-js/extra-typings';

export type BaseCommand = {
  buildCommand(): Command<unknown[], Record<string, unknown>>;
};
