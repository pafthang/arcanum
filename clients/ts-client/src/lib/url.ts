export function createUrl(
	base: string,
	queryParams?: Record<string, unknown>,
	routeParams?: Record<string, unknown>
): string {
	let url = base;
	for (const [key, value] of Object.entries(routeParams ?? {})) {
		if (value === undefined || value === null || String(value).trim() === '') {
			throw new Error(`missing route param "${key}"`);
		}
		url = url.replaceAll(`:${key}`, encodeURIComponent(String(value)));
	}
	if (!queryParams) return url;
	const query = new URLSearchParams();
	for (const [key, value] of Object.entries(queryParams)) {
		if (value === undefined || value === null || value === '') continue;
		if (value === 0) {
			query.append(key, '0');
			continue;
		}
		query.append(key, typeof value === 'object' ? JSON.stringify(value) : String(value));
	}
	const qs = query.toString();
	return qs ? `${url}?${qs}` : url;
}
