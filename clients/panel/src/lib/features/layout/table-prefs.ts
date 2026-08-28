export function loadColumnPrefs(key: string): Record<string, boolean> {
	if (typeof localStorage === 'undefined') return {};
	try {
		const raw = localStorage.getItem(`ow.table.${key}.cols`);
		return raw ? (JSON.parse(raw) as Record<string, boolean>) : {};
	} catch {
		return {};
	}
}

export function saveColumnPrefs(
	key: string,
	view: { columns: { name?: string; isVisible?: boolean }[] }
): void {
	if (typeof localStorage === 'undefined') return;
	const rec: Record<string, boolean> = {};
	for (const column of view.columns) {
		if (column.name) rec[column.name] = column.isVisible !== false;
	}
	localStorage.setItem(`ow.table.${key}.cols`, JSON.stringify(rec));
}

export function columnVisible(prefs: Record<string, boolean>, name: string): boolean {
	return prefs[name] !== false;
}

export function loadPageSize(key: string, fallback = 20): number {
	if (typeof localStorage === 'undefined') return fallback;
	const n = Number(localStorage.getItem(`ow.table.${key}.rpp`));
	return n > 0 ? n : fallback;
}

export function savePageSize(key: string, size: number): void {
	if (typeof localStorage === 'undefined') return;
	localStorage.setItem(`ow.table.${key}.rpp`, String(size));
}
