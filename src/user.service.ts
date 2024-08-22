import type { SupabaseClient, User } from '@supabase/supabase-js';

export class UserService {
	constructor(private supabaseClient: SupabaseClient) {}

	async getUser(): Promise<User | null> {
		const userResponse = await this.supabaseClient.auth.getUser();
		return userResponse.data.user;
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
