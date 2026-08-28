export function moveByUuid<T extends { uuid: string }>(items: T[], fromUuid: string, toUuid: string): T[] {
	if (!fromUuid || fromUuid === toUuid) return items;
	const next = [...items];
	const from = next.findIndex((item) => item.uuid === fromUuid);
	const to = next.findIndex((item) => item.uuid === toUuid);
	if (from < 0 || to < 0) return items;
	const [row] = next.splice(from, 1);
	next.splice(to, 0, row);
	return next;
}

export function sortByPosition<T extends { viewPosition?: number }>(items: T[]): T[] {
	return [...items].sort((a, b) => (a.viewPosition ?? 0) - (b.viewPosition ?? 0));
}
