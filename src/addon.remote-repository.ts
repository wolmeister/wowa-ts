import got, { HTTPError } from 'got';
import type { GameVersion } from './addon.repository';
import type { UserService } from './user.service';

export type RemoteAddon = {
  id: string;
  user_id: string;
  game_version: GameVersion;
  slug: string;
  name: string;
  author: string;
  provider: 'curse';
  external_id: string;
  url: string;
  created_at: string;
  updated_at: string;
};

export type CreateAddonRequest = {
  game_version: GameVersion;
  slug: string;
  name: string;
  author: string;
  provider: 'curse';
  external_id: string;
  url: string;
};

export class AddonRemoteRepository {
  constructor(
    private userService: UserService,
    private apiUrl: string,
  ) {}

  async createAddon(addon: CreateAddonRequest): Promise<RemoteAddon> {
    const token = await this.userService.getUserToken();
    if (token === null) {
      throw new Error('No user signed in');
    }

    const response = await got.post({
      url: `${this.apiUrl}/addons`,
      body: JSON.stringify(addon),
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
    });

    return JSON.parse(response.body);
  }

  async deleteAddon(slug: string, gameVersion: GameVersion): Promise<void> {
    const token = await this.userService.getUserToken();
    if (token === null) {
      throw new Error('No user signed in');
    }

    await got.delete({
      url: `${this.apiUrl}/addons/${gameVersion}/${slug}`,
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
  }

  async getAddons(): Promise<RemoteAddon[]> {
    const token = await this.userService.getUserToken();
    if (token === null) {
      throw new Error('No user signed in');
    }

    const response = await got.get({
      url: `${this.apiUrl}/addons`,
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    return JSON.parse(response.body);
  }

  async getAddon(slug: string, gameVersion: GameVersion): Promise<RemoteAddon | null> {
    const token = await this.userService.getUserToken();
    if (token === null) {
      throw new Error('No user signed in');
    }

    try {
      const response = await got.get({
        url: `${this.apiUrl}/addons/${gameVersion}/${slug}`,
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      return JSON.parse(response.body);
    } catch (error) {
      if (error instanceof HTTPError) {
        if (error.response.statusCode === 404) {
          return null;
        }
      }
      throw error;
    }
  }
}
