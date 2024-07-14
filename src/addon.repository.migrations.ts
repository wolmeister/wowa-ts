import type { RepositoryMigration } from './kv-store.repository';

type AddonV1 = {
	directories: string[];
};

type AddonV2 = {
	directories: {
		name: string;
		hash: { algorithm: 'none' };
	}[];
};

const toV2: RepositoryMigration<AddonV1, AddonV2> = (addon) => {
	return {
		...addon,
		directories: addon.directories.map((d) => ({ name: d, hash: { algorithm: 'none' } })),
	};
};

export const addonRepositoryMigrations: RepositoryMigration[] = [toV2 as RepositoryMigration];
