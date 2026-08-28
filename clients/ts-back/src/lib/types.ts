export type Json = unknown;

export type TableQuery = {
	start?: number;
	size?: number;
	search?: string;
	status?: string;
	filters?: { id: string; value: unknown }[];
	filterModes?: Record<string, string>;
	sorting?: { id: string; desc: boolean }[];
};

export type ReorderItem = { uuid: string; viewPosition: number };

export type Credentials = { username: string; password: string };
export type TokenResponse = { accessToken: string };
export type OAuth2Authorize = { provider: string; code?: string; state?: string };

export type AuthStatus = {
	isLoginAllowed: boolean;
	isRegisterAllowed: boolean;
	authentication: {
		passkey: { enabled: boolean };
		oauth2: {
			providers: {
				github: boolean;
				pocketid: boolean;
				yandex: boolean;
				keycloak: boolean;
				generic: boolean;
				telegram: boolean;
			};
		};
		password: { enabled: boolean };
	} | null;
	branding: { title: string | null; logoUrl: string | null };
};

export type UserTraffic = {
	usedTrafficBytes: number;
	lifetimeUsedTrafficBytes: number;
	onlineAt: string | null;
	lastConnectedNodeUuid: string | null;
	firstConnectedAt: string | null;
};

export type User = {
	id: number;
	shortUuid: string;
	username: string;
	status: string;
	trafficLimitBytes: number;
	trafficLimitStrategy: string;
	expireAt: string;
	telegramId: number | null;
	email: string | null;
	description: string | null;
	tag: string | null;
	hwidDeviceLimit: number | null;
	externalSquadUuid: string | null;
	trojanPassword: string;
	vlessUuid: string;
	ssPassword: string;
	lastTriggeredThreshold: number;
	subRevokedAt: string | null;
	lastTrafficResetAt: string | null;
	subscriptionUrl: string;
	activeInternalSquads: { uuid: string; name: string }[];
	userTraffic: UserTraffic;
	createdAt: string;
	updatedAt: string;
};

export type UserList = { users: User[]; total: number };
export type UserStream = { users: User[]; nextCursor: string | null; hasMore: boolean };
export type ResolveUser = { id: number; username: string; shortUuid: string };
export type AccessibleNodes = {
	userId: number;
	activeNodes: {
		uuid: string;
		nodeName: string;
		countryCode: string;
		configProfileUuid: string;
		configProfileName: string;
		activeSquads: { squadName: string; activeInbounds: string[] }[];
	}[];
};
export type UserHistory = {
	total: number;
	records: {
		id: number;
		userId: number;
		requestAt: string;
		srrResponseType: string;
		requestIp: string | null;
		userAgent: string | null;
		srrRuleName: string | null;
	}[];
};

export type Node = {
	uuid: string;
	id: number;
	name: string;
	address: string;
	port: number | null;
	proxyUrl: string | null;
	isConnected: boolean;
	isConnecting: boolean;
	isDisabled: boolean;
	lastStatusChange: string | null;
	lastStatusMessage: string | null;
	isTrafficTrackingActive: boolean;
	trafficResetDay: number | null;
	trafficLimitBytes: number;
	trafficUsedBytes: number;
	notifyPercent: number | null;
	viewPosition: number;
	countryCode: string;
	consumptionMultiplier: number;
	nodeConsumptionMultiplier: number;
	tags: string[];
	integrationUuids: string[];
	ips: Json;
	createdAt: string;
	updatedAt: string;
	configProfile: Json;
	providerUuid: string | null;
	provider: Json;
	activePluginUuid: string | null;
	note: string | null;
	system: Json;
	versions: Json;
	xrayUptime: number;
	usersOnline: number;
};

export type Host = {
	uuid: string;
	viewPosition: number;
	remark: string;
	address: string;
	port: number;
	path: string | null;
	sni: string | null;
	host: string | null;
	alpn: string | null;
	fingerprint: string | null;
	isDisabled: boolean;
	securityLayer: string;
	tags: string[];
	isHidden: boolean;
	nodes: string[];
	inbound: { configProfileUuid: string | null; configProfileInboundUuid: string | null };
	xrayJsonTemplateUuid: string | null;
	excludedInternalSquads: string[];
	excludeFromSubscriptionTypes: string[];
};

export type Tags = { tags: string[] };

export type ConfigProfile = {
	uuid: string;
	viewPosition: number;
	name: string;
	config: Json;
	inbounds: ConfigInbound[];
	nodes: { uuid: string; name: string; countryCode: string }[];
	createdAt: string;
	updatedAt: string;
};

export type ConfigInbound = {
	uuid: string;
	profileUuid: string;
	tag: string;
	type: string;
	network: string | null;
	security: string | null;
	port: number | null;
	rawInbound: Json;
	activeSquads?: string[];
};

export type ConfigProfileList = { total: number; configProfiles: ConfigProfile[] };

export type Snippet = { name: string; snippet: Json };
export type SnippetList = { total: number; snippets: Snippet[] };

export type InternalSquad = {
	uuid: string;
	viewPosition: number;
	name: string;
	info: { membersCount: number; inboundsCount: number };
	inbounds: ConfigInbound[];
	createdAt: string;
	updatedAt: string;
};

export type ExternalSquad = {
	uuid: string;
	viewPosition: number;
	name: string;
	info: { membersCount: number };
	templates: { templateUuid: string; templateType: string }[];
	subscriptionSettings: Json;
	hostOverrides: Json;
	createdAt: string;
	updatedAt: string;
};

export type HwidDevice = {
	hwid: string;
	userId: number;
	platform: string | null;
	osVersion: string | null;
	deviceModel: string | null;
	userAgent: string | null;
	requestIp: string | null;
	createdAt: string;
	updatedAt: string;
};

export type HwidList = { total: number; devices: HwidDevice[] };

export type ApiToken = {
	uuid: string;
	name: string;
	expireAt: string;
	scopes: string[];
	createdAt: string;
	updatedAt: string;
	token?: string;
};

export type ApiTokenList = { tokens: ApiToken[] };
export type Ott = { ott: string };

export type NodePlugin = { uuid: string; viewPosition: number; name: string; pluginConfig: Json };
export type NodePluginList = { total: number; nodePlugins: NodePlugin[] };
export type SharedList = { name: string; config: Json };
export type SharedListPreview = { name: string; type: string; itemsCount: number };
export type SharedListList = { total: number; sharedLists: SharedListPreview[] };

export type Keygen = { secretKey: string; pubKey: string };

export type SystemStats = {
	cpu: { cores: number };
	memory: { total: number; free: number; used: number };
	uptime: number;
	timestamp: number;
	users: { statusCounts: Record<string, number>; totalUsers: number };
	onlineStats: { lastDay: number; lastWeek: number; neverOnline: number; onlineNow: number };
	nodes: { totalOnline: number; totalBytesLifetime: string };
};

export type SystemMetadata = {
	version: string;
	build: { time: string; number: string };
	git: {
		backend: { commitSha: string; branch: string; commitUrl: string };
		frontend: { commitSha: string; commitUrl: string };
	};
};

export type BandwidthStats = {
	bandwidthLastTwoDays: { current: string; previous: string; difference: string };
	bandwidthLastSevenDays: { current: string; previous: string; difference: string };
	bandwidthLast30Days: { current: string; previous: string; difference: string };
	bandwidthCalendarMonth: { current: string; previous: string; difference: string };
	bandwidthCurrentYear: { current: string; previous: string; difference: string };
};

export type NodesUsage = {
	categories: string[];
	sparklineData: number[];
	topNodes: { uuid: string; color: string; name: string; countryCode: string; total: number }[];
	series: {
		uuid: string;
		name: string;
		color: string;
		countryCode: string;
		total: number;
		data: number[];
	}[];
};
