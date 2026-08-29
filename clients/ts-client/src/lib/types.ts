export interface User {
	id: string;
	email: string;
	name: string;
	role: string;
	createdAt?: string;
	updatedAt?: string;
}

export interface Space {
	id: string;
	name: string;
	slug?: string;
	role?: string;
	createdAt?: string;
	updatedAt?: string;
}

export interface SpaceMember {
	id: string;
	spaceId: string;
	userId: string;
	role: string;
	user?: User;
	createdAt?: string;
}

export interface ApiKey {
	id: string;
	spaceId: string;
	name: string;
	keyHash?: string;
	actorType: 'user' | 'agent';
	createdAt: string;
}

export interface Team {
	id: string;
	spaceId: string;
	name: string;
	description?: string;
	createdAt?: string;
	updatedAt?: string;
}

export interface Issue {
	id: string;
	spaceId: string;
	title: string;
	body?: string;
	status: 'open' | 'started' | 'done';
	assigneeId?: string;
	assigneeIds?: string[];
	priority?: string;
	dueAt?: string;
	parentId?: string;
	relations?: IssueRelation[];
	labels?: Label[];
	createdAt: string;
	updatedAt: string;
}

export interface IssueRelation {
	id: string;
	spaceId: string;
	fromId: string;
	toId: string;
	kind: 'blocks' | 'blocked_by' | 'duplicate' | 'related';
	createdAt: string;
}

export interface IssueActivity {
	id: string;
	issueId: string;
	actorId: string;
	type: string;
	payload?: Record<string, any>;
	createdAt: string;
}

export interface Label {
	id: string;
	spaceId: string;
	name: string;
	color?: string;
	createdAt?: string;
}

export interface Comment {
	id: string;
	issueId: string;
	actorId: string;
	body: string;
	blobId?: string;
	createdAt: string;
}

export interface Cycle {
	id: string;
	spaceId: string;
	name: string;
	description?: string;
	status: string;
	startDate?: string;
	endDate?: string;
	createdAt: string;
	updatedAt: string;
}

export interface Project {
	id: string;
	spaceId: string;
	name: string;
	key?: string;
	description?: string;
	status: string;
	leadId?: string;
	createdAt: string;
	updatedAt: string;
}

export interface View {
	id: string;
	spaceId: string;
	name: string;
	description?: string;
	query?: string;
	icon?: string;
	createdBy?: string;
	createdAt: string;
}

export interface WorkOverview {
	issues: number;
	byStatus: Record<string, number>;
	assigned: number;
	unassigned: number;
	comments: number;
}

export interface AgentRun {
	id: string;
	spaceId: string;
	agentId: string;
	issueId?: string;
	input?: string;
	status: 'queued' | 'running' | 'succeeded' | 'failed' | 'cancelling' | 'cancelled';
	output?: string;
	error?: string;
	startedAt?: string;
	finishedAt?: string;
	createdAt: string;
	updatedAt: string;
}

export interface AgentSession {
	id: string;
	runId: string;
	spaceId: string;
	stage: string;
	payload: string;
	createdAt: string;
	updatedAt: string;
}

export interface Memory {
	id: string;
	spaceId: string;
	agentId: string;
	tier: 'working' | 'episodic' | 'semantic';
	key: string;
	value: string;
	updatedAt: string;
}

export interface Skill {
	id: string;
	spaceId: string;
	name: string;
	body: string;
	createdAt: string;
}

export interface Channel {
	id: string;
	spaceId: string;
	teamId?: string;
	name: string;
	kind: 'space' | 'team' | 'dm';
	createdAt: string;
	updatedAt: string;
}

export interface Message {
	id: string;
	channelId: string;
	spaceId: string;
	parentId?: string;
	actorId: string;
	body: string;
	blobId?: string;
	source: 'user' | 'agent' | 'integ';
	externalRef?: string;
	createdAt: string;
}

export interface Reaction {
	id: string;
	messageId: string;
	spaceId: string;
	actorId: string;
	emoji: string;
	createdAt: string;
}

export interface Connector {
	id: string;
	spaceId: string;
	kind: string;
	name: string;
	status: string;
	config?: Record<string, any>;
	hasSecret: boolean;
	createdAt: string;
	updatedAt: string;
}

export interface GitHubRepo {
	id: string;
	connectorId: string;
	spaceId: string;
	owner: string;
	name: string;
	installationId?: string;
	createdAt: string;
}

export interface Webhook {
	id: string;
	spaceId: string;
	url: string;
	events: string[];
	active: boolean;
	hasSecret: boolean;
	createdAt: string;
}

export interface WebhookDelivery {
	id: string;
	spaceId: string;
	targetType: string;
	targetId: string;
	eventType: string;
	status: string;
	attempts: number;
	lastError?: string;
	createdAt: string;
	deliveredAt?: string;
}

export interface BlobInfo {
	id: string;
	spaceId: string;
	filename: string;
	contentType: string;
	sizeBytes: number;
	createdAt: string;
}

export interface Machine {
	id: string;
	spaceId: string;
	name: string;
	image: string;
	agentId?: string;
	status: 'recorded' | 'running' | 'stopped' | 'failed';
	dockerId?: string;
	error?: string;
	createdAt: string;
	updatedAt: string;
}

/** Matches Go `logg/models.Activity`. Note: Go field is `created`, not `createdAt`. */
export interface Activity {
	id: string;
	spaceId: string;
	targetType?: string;
	targetId?: string;
	actorId?: string;
	type: string;
	summary: string;
	payload?: Record<string, any>;
	created: string;
}

/**
 * Paginated activity list response from `GET /api/spaces/:spaceId/activity`.
 */
export interface ActivityPage {
	page: number;
	perPage: number;
	totalItems: number;
	items: Activity[];
}

/**
 * @deprecated Use {@link Activity} instead. Kept for backwards compatibility.
 */
export type LogActivity = Activity;
