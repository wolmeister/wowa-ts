import type { SupabaseClient, User } from '@supabase/supabase-js';
import { Mutex } from 'async-mutex';

export class UserService {
	private readonly getUserMutex = new Mutex();
	private cachedUser: User | null = null;

	constructor(private supabaseClient: SupabaseClient) {}

	async getUser(): Promise<User | null> {
		return this.getUserMutex.runExclusive(async () => {
			if (this.cachedUser !== null) {
				return this.cachedUser;
			}
			const userResponse = await this.supabaseClient.auth.getUser();
			this.cachedUser = userResponse.data.user;
			return this.cachedUser;
		});
	}

	async sendLoginEmail(email: string): Promise<void> {
		await this.supabaseClient.auth.signInWithOtp({
			email,
		});
	}

	async finishEmailLogin(email: string, token: string): Promise<User> {
		const authResponse = await this.supabaseClient.auth.verifyOtp({
			type: 'email',
			email,
			token,
		});
		if (authResponse.error !== null) {
			throw authResponse.error;
		}
		if (authResponse.data.user == null) {
			throw new Error('User not found');
		}
		return authResponse.data.user;
	}
}
