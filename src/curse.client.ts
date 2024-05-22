// https://docs.curseforge.com/#tocS_Pagination
export type Pagination = {
	index: number;
	pageSize: number;
	resultCount: number;
	totalCount: number;
};

// https://docs.curseforge.com/#tocS_Category
export type CurseCategory = {
	id: number;
	gameId: number;
	name: string;
	slug: string;
	url: string;
	iconUrl: string;
	dateModified: string;
	isClass?: boolean;
	classId?: number;
	parentCategoryId?: number;
	displayIndex?: number;
};

// https://docs.curseforge.com/#tocS_ModAuthor
export type CurseModAuthor = {
	id: number;
	name: string;
	url: string;
};

// https://docs.curseforge.com/#tocS_ModAsset
export type CurseModAsset = {
	id: number;
	modId: number;
	title: string;
	description: string;
	thumbnailUrl: string;
	url: string;
};

// https://docs.curseforge.com/#tocS_FileRelationType
export enum CurseFileReleaseType {
	Release = 1,
	Beta = 2,
	Alpha = 3,
}

// https://docs.curseforge.com/#tocS_FileStatus
export enum CurseFileStatus {
	Processing = 1,
	ChangesRequired = 2,
	UnderReview = 3,
	Approved = 4,
	Rejected = 5,
	MalwareDetected = 6,
	Deleted = 7,
	Archived = 8,
	Testing = 9,
	Released = 10,
	ReadyForReview = 11,
	Deprecated = 12,
	Baking = 13,
	AwaitingPublishing = 14,
	FailedPublishing = 15,
}

// https://docs.curseforge.com/#tocS_HashAlgo
export enum CurseHashAlgorithm {
	Sha1 = 1,
	Md5 = 2,
}

// https://docs.curseforge.com/#tocS_FileHash
export type CurseFileHash = {
	value: string;
	algo: CurseHashAlgorithm;
};

// https://docs.curseforge.com/#tocS_SortableGameVersion
export type CurseSortableGameVersion = {
	gameVersionName: string;
	gameVersionPadded: string;
	gameVersion: string;
	gameVersionReleaseDate: string;
	gameVersionTypeId?: number;
};

// https://docs.curseforge.com/#tocS_FileRelationType
export enum CurseFileRelationType {
	EmbeddedLibrary = 1,
	OptionalDependency = 2,
	RequiredDependency = 3,
	Tool = 4,
	Incompatible = 5,
	Include = 6,
}

// https://docs.curseforge.com/#tocS_FileDependency
export type CurseFileDependency = {
	modId: number;
	relationType: CurseFileRelationType;
};

// https://docs.curseforge.com/#tocS_FileModule
export type CurseFileModule = {
	name: string;
	fingerprint: number;
};

// https://docs.curseforge.com/#tocS_File
export type CurseFile = {
	id: number;
	gameId: number;
	modId: number;
	isAvailable: boolean;
	displayName: string;
	fileName: string;
	releaseType: CurseFileReleaseType;
	fileStatus: CurseFileStatus;
	hashes: CurseFileHash[];
	fileDate: string;
	fileLength: number;
	downloadCount: number;
	fileSizeOnDisk?: number;
	downloadUrl: string;
	gameVersions: string[];
	sortableGameVersions: CurseSortableGameVersion[];
	dependencies: CurseFileDependency[];
	exposeAsAlternative?: boolean;
	parentProjectFileId?: number;
	alternateFileId?: number;
	isServerPack?: boolean;
	serverPackFileId?: number;
	isEarlyAccessContent?: boolean;
	earlyAccessEndDate?: string;
	fileFingerprint: number;
	modules: CurseFileModule[];
};

// https://docs.curseforge.com/#tocS_ModLoaderType
export enum CurseModLoaderType {
	Any = 0,
	Forge = 1,
	Cauldron = 2,
	LiteLoader = 3,
	Fabric = 4,
	Quilt = 5,
	NeoForge = 6,
}

// https://docs.curseforge.com/#tocS_FileIndex
export type CurseFileIndex = {
	gameVersion: string;
	fileId: number;
	filename: string;
	releaseType: CurseFileReleaseType;
	gameVersionTypeId?: number;
	modLoader?: CurseModLoaderType;
};

// https://docs.curseforge.com/#tocS_ModLinks
export type CurseModLinks = {
	websiteUrl: string;
	wikiUrl: string;
	issuesUrl: string;
	sourceUrl: string;
};

// https://docs.curseforge.com/#tocS_ModStatus
export enum CurseModStatus {
	New = 1,
	ChangesRequired = 2,
	UnderSoftReview = 3,
	Approved = 4,
	Rejected = 5,
	ChangesMade = 6,
	Inactive = 7,
	Abandoned = 8,
	Deleted = 9,
	UnderReview = 10,
}

// https://docs.curseforge.com/#tocS_Mod
export type CurseMod = {
	id: number;
	gameId: number;
	name: string;
	slug: string;
	links: CurseModLinks;
	summary: string;
	status: CurseModStatus;
	downloadCount: number;
	isFeatured: boolean;
	primaryCategoryId: number;
	categories: CurseCategory[];
	classId?: number;
	authors: CurseModAuthor[];
	logo: CurseModAsset;
	screenshots: CurseModAsset[];
	mainFileId: number;
	latestFiles: CurseFile[];
	latestFilesIndexes: CurseFileIndex[];
	latestEarlyAccessFilesIndexes: CurseFileIndex[];
	dateCreated: string;
	dateModified: string;
	dateReleased: string;
	allowModDistribution?: boolean;
	gamePopularityRank: number;
	isAvailable: boolean;
	thumbsUpCount: number;
	rating?: number;
};

/** API RESPONSES / FILTERS **/
export enum SearchModsSortField {
	Featured = 1,
	Popularity = 2,
	LastUpdated = 3,
	Name = 4,
	Author = 5,
	TotalDownloads = 6,
	Category = 7,
	GameVersion = 8,
	EarlyAccess = 9,
	FeaturedReleased = 10,
	ReleasedDate = 11,
	Rating = 12,
}

export enum SearchModsSortOrder {
	Asc = 0,
	Desc = 1,
}

// TODO - This is incomplete. Implement all supported filters.
export type SearchModsParams = {
	gameId: number;
	gameVersionTypeId?: number;
	slug?: string;
	index?: number;
	sortField?: SearchModsSortField;
	sortOrder?: SearchModsSortOrder;
};

export type SearchModsResponse = {
	data: CurseMod[];
	pagination: Pagination;
};

export type ModFileResponse = {
	data: CurseFile;
};

export class CurseClient {
	private token: string | null = null;

	setToken(token: string): void {
		this.token = token;
	}

	async searchMods(searchParams: SearchModsParams): Promise<SearchModsResponse> {
		if (this.token === null) {
			throw new Error('Curse token not defined');
		}

		const url = new URL('https://api.curseforge.com/v1/mods/search');
		url.searchParams.set('gameId', String(searchParams.gameId));

		if (searchParams.gameVersionTypeId != null) {
			url.searchParams.set('gameVersionTypeId', String(searchParams.gameVersionTypeId));
		}
		if (searchParams.slug != null) {
			url.searchParams.set('slug', searchParams.slug);
		}
		if (searchParams.index != null) {
			url.searchParams.set('index', String(searchParams.index));
		}
		if (searchParams.sortField != null) {
			url.searchParams.set('sortField', String(searchParams.sortField));
		}
		if (searchParams.sortOrder != null) {
			url.searchParams.set('sortOrder', SearchModsSortOrder.Asc ? 'asc' : 'desc');
		}

		const response = await fetch(url, {
			headers: {
				'x-api-key': this.token,
			},
		});
		if (response.ok === false) {
			throw new Error(await response.text());
		}
		const json = await response.json();
		return json;
	}

	async getModFile(modId: number, fileId: number): Promise<ModFileResponse> {
		if (this.token === null) {
			throw new Error('Curse token not defined');
		}

		const response = await fetch(`https://api.curseforge.com/v1/mods/${modId}/files/${fileId}`, {
			headers: {
				'x-api-key': this.token,
			},
		});
		if (response.ok === false) {
			throw new Error(await response.text());
		}
		const json = await response.json();
		return json;
	}
}
