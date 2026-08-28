import { clearQueryCache } from './query-cache.js';

const STORAGE_KEY = 'remnawave.session';

export class Session {
	token = $state('');
	initialized = $state(false);

	get isAuthenticated(): boolean {
		return this.token.length > 0;
	}

	init(): void {
		if (typeof localStorage === 'undefined') {
			this.initialized = true;
			return;
		}
		this.token = localStorage.getItem(STORAGE_KEY) ?? '';
		this.initialized = true;
	}

	setToken(token: string): void {
		this.token = token;
		if (typeof localStorage !== 'undefined') {
			if (token) localStorage.setItem(STORAGE_KEY, token);
			else localStorage.removeItem(STORAGE_KEY);
		}
	}

	clear(): void {
		this.setToken('');
		clearQueryCache();
	}
}

export const session = new Session();

if (typeof window !== 'undefined') {
	session.init();
}
