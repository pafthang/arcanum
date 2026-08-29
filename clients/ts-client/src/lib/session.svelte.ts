import { clearQueryCache } from './query-cache.js';
import type { User } from './types.js';

const STORAGE_TOKEN_KEY = 'arcanum.token';
const STORAGE_SPACE_KEY = 'arcanum.space_id';

export class Session {
	token = $state('');
	spaceId = $state('');
	user = $state<User | null>(null);
	initialized = $state(false);

	get isAuthenticated(): boolean {
		return this.token.length > 0;
	}

	init(): void {
		if (typeof localStorage === 'undefined') {
			this.initialized = true;
			return;
		}
		this.token = localStorage.getItem(STORAGE_TOKEN_KEY) ?? '';
		this.spaceId = localStorage.getItem(STORAGE_SPACE_KEY) ?? '';
		this.initialized = true;
	}

	setToken(token: string): void {
		this.token = token;
		if (typeof localStorage !== 'undefined') {
			if (token) localStorage.setItem(STORAGE_TOKEN_KEY, token);
			else localStorage.removeItem(STORAGE_TOKEN_KEY);
		}
	}

	setSpace(spaceId: string): void {
		this.spaceId = spaceId;
		if (typeof localStorage !== 'undefined') {
			if (spaceId) localStorage.setItem(STORAGE_SPACE_KEY, spaceId);
			else localStorage.removeItem(STORAGE_SPACE_KEY);
		}
	}

	setUser(user: User | null): void {
		this.user = user;
	}

	clear(): void {
		this.setToken('');
		this.setSpace('');
		this.setUser(null);
		clearQueryCache();
	}
}

export const session = new Session();

if (typeof window !== 'undefined') {
	session.init();
}
