import fs from 'node:fs/promises';
import path from 'node:path';
import { Mutex } from 'async-mutex';

type KeyValueEntry = {
	key: string[];
	value: string;
};

export class KeyValueStore {
	private storePath: string | null = null;
	private entries: KeyValueEntry[] = [];

	private readonly mutex = new Mutex();

	async init(storePath: string): Promise<void> {
		return this.mutex.runExclusive(async () => {
			if (this.storePath !== null) {
				throw new Error('The KeyValueStore is already initialized');
			}
			this.storePath = storePath;

			if (await fs.exists(storePath)) {
				const entriesString = await fs.readFile(storePath, 'utf-8');
				this.entries = JSON.parse(entriesString);
			}
		});
	}

	private keyEquals(a: string[], b: string[]): boolean {
		if (a.length !== b.length) {
			return false;
		}

		return a.every((v, i) => v === b[i]);
	}

	async get(key: string[]): Promise<string | null> {
		return this.mutex.runExclusive(() => {
			if (this.storePath === null) {
				throw new Error('The KeyValueStore is not initialized yet');
			}

			const entry = this.entries.find((e) => this.keyEquals(e.key, key));
			return entry?.value ?? null;
		});
	}

	async getByPrefix(prefix: string[]): Promise<string[]> {
		return this.mutex.runExclusive(() => {
			if (this.storePath === null) {
				throw new Error('The KeyValueStore is not initialized yet');
			}

			return this.entries
				.filter((e) => this.keyEquals(e.key.slice(0, prefix.length), prefix))
				.map((e) => e.value);
		});
	}

	async set(key: string[], value: string): Promise<void> {
		return this.mutex.runExclusive(async () => {
			if (this.storePath === null) {
				throw new Error('The KeyValueStore is not initialized yet');
			}

			this.entries = this.entries.filter((e) => !this.keyEquals(e.key, key));
			this.entries.push({ key, value });

			const directory = path.dirname(this.storePath);
			const directoryExists = await fs.exists(directory);
			if (directoryExists === false) {
				await fs.mkdir(directory, { recursive: true });
			}

			const jsonString = JSON.stringify(this.entries, null, 2);
			await fs.writeFile(this.storePath, jsonString, 'utf-8');
		});
	}
}
