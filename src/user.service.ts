import got from 'got';
import type { ConfigRepository } from './config.repository';

export class UserService {
  constructor(
    private configRepository: ConfigRepository,
    private apiUrl: string,
  ) {}

  async getUserToken(): Promise<string | null> {
    return this.configRepository.get('auth.token');
  }

  async getUserEmail(): Promise<string | null> {
    const token = await this.getUserToken();
    if (token == null) {
      return null;
    }
    const stringPayload = Buffer.from(token.split('.')[1], 'base64').toString('utf-8');
    const payload = JSON.parse(stringPayload);
    return payload.email;
  }

  async signin(email: string, password: string): Promise<void> {
    const authResponse = await got.post({
      url: `${this.apiUrl}/login`,
      body: JSON.stringify({ email, password }),
      headers: {
        'Content-Type': 'application/json',
      },
    });
    await this.configRepository.set('auth.token', authResponse.body);
  }
}
