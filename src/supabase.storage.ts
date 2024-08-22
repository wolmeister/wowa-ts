import type { SupportedStorage } from '@supabase/supabase-js';
import type { ConfigRepository } from './config.repository';
import type { KeyValueStore } from './kv-store';

export class SupabaseStorage implements SupportedStorage {
	constructor(private kvStore: KeyValueStore) {}

	getItem(key: string): Promise<string | null> {
		return this.kvStore.get(['supabase', 'auth', key]);
	}

	setItem(key: string, value: string): Promise<void> {
		return this.kvStore.set(['supabase', 'auth', key], value);
	}

	removeItem(key: string): Promise<void> {
		return this.kvStore.set(['supabase', 'auth', key], null);
	}
}
