import got from 'got';

export type GithubUser = {
	login: string;
	id: number;
	nodeId: string;
	avatarUrl: string;
	gravatarId: string;
	url: string;
	htmlUrl: string;
	followersUrl: string;
	followingUrl: string;
	gistsUrl: string;
	starredUrl: string;
	subscriptionsUrl: string;
	organizationsUrl: string;
	reposUrl: string;
	eventsUrl: string;
	receivedEventsUrl: string;
	type: string;
	siteAdmin: boolean;
};

export type GithubReleaseAsset = {
	url: string;
	id: number;
	nodeId: string;
	name: string;
	label: string;
	uploader: GithubUser;
	contentType: string;
	state: string;
	size: number;
	downloadCount: number;
	createdAt: Date;
	updatedAt: Date;
	browserDownloadUrl: string;
};

export type GithubRelease = {
	url: string;
	assetsUrl: string;
	uploadUrl: string;
	htmlUrl: string;
	id: number;
	author: GithubUser;
	nodeId: string;
	tagName: string;
	targetCommitish: string;
	name: string;
	draft: boolean;
	prerelease: boolean;
	createdAt: Date;
	publishedAt: Date;
	assets: GithubReleaseAsset[];
	tarballUrl: string;
	zipballUrl: string;
};

function toCamelCase(str: string): string {
	let newStr = '';
	let capitalize = false;
	for (let i = 0; i < str.length; i += 1) {
		if (capitalize) {
			newStr += str[i].toUpperCase();
			capitalize = false;
			continue;
		}
		if (str[i] === '_') {
			capitalize = true;
			continue;
		}
		newStr += str[i];
	}
	return newStr;
}

function objectToCamelCase<T extends object>(obj: object): T {
	if (typeof obj !== 'object') {
		throw new Error('invalid-object');
	}
	const newObj: Record<string, unknown> = {};
	for (const [key, value] of Object.entries(obj)) {
		let parsedValue = value;
		if (Array.isArray(value)) {
			parsedValue = value.map((arrayValue) => {
				if (typeof arrayValue === 'object') {
					return objectToCamelCase(arrayValue);
				}
				return arrayValue;
			});
		} else if (typeof parsedValue === 'object') {
			parsedValue = objectToCamelCase(parsedValue);
		}
		newObj[toCamelCase(key)] = parsedValue;
	}
	return newObj as T;
}

export async function getLatestGithubRelease(
	organization: string,
	repository: string,
): Promise<GithubRelease> {
	const url = `https://api.github.com/repos/${organization}/${repository}/releases/latest`;
	const res = await got(url, {
		responseType: 'json',
		headers: { 'User-Agent': 'wowa-app' },
	});
	if (!res.ok) {
		throw new Error('request-failed');
	}
	return objectToCamelCase(res.body as object);
}
