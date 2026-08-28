export function asArray<T>(data: unknown, keys: string[]): T[] {
	if (Array.isArray(data)) return data as T[];
	if (data && typeof data === 'object') {
		const rec = data as Record<string, unknown>;
		for (const key of keys) {
			const value = rec[key];
			if (Array.isArray(value)) return value as T[];
		}
	}
	return [];
}

export function pretty(value: unknown): string {
	try {
		return JSON.stringify(value, null, 2);
	} catch {
		return String(value ?? '');
	}
}
