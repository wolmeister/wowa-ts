import type { RepositoryMigration } from './kv-store.repository';

type AddonV2Source = {
  directories: string[];
};

type AddonV2Target = {
  directories: {
    name: string;
    hash: { algorithm: 'none' };
  }[];
};

const toV2: RepositoryMigration<AddonV2Source, AddonV2Target> = (addon) => {
  return {
    ...addon,
    directories: addon.directories.map((d) => ({ name: d, hash: { algorithm: 'none' } })),
  };
};

type AddonV3Source = Record<string, unknown>;

type AddonV3Target = {
  installedAt: string | null;
  updatedAt: string | null;
};

const toV3: RepositoryMigration<AddonV3Source, AddonV3Target> = (addon) => {
  return {
    ...addon,
    installedAt: null,
    updatedAt: null,
  };
};

export const addonRepositoryMigrations: RepositoryMigration[] = [
  toV2 as RepositoryMigration,
  toV3 as RepositoryMigration,
];
