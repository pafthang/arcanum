import type { TableHandler, Row } from '@vincjo/datatables';

const lastRows = new WeakMap<object, unknown[]>();
const lastSig = new WeakMap<object, string>();

function signature(rows: { id?: unknown; uuid?: unknown; updatedAt?: unknown; status?: unknown; name?: unknown }[]): string {
	return rows
		.map((row) => `${String(row.id ?? row.uuid)}:${String(row.updatedAt ?? row.status ?? row.name ?? '')}`)
		.join('|');
}

function toPlain(value: unknown, seen = new WeakMap<object, unknown>()): unknown {
	if (value === null || typeof value !== 'object') return value;
	if (seen.has(value)) return seen.get(value);
	if (value instanceof Date) return value.toISOString();
	if (Array.isArray(value)) {
		const copy: unknown[] = [];
		seen.set(value, copy);
		for (let i = 0; i < value.length; i += 1) copy[i] = toPlain(value[i], seen);
		return copy;
	}
	const copy: Record<string, unknown> = {};
	seen.set(value, copy);
	for (const key of Object.keys(value)) {
		const item = (value as Record<string, unknown>)[key];
		if (typeof item === 'function') continue;
		copy[key] = toPlain(item, seen);
	}
	return copy;
}

/** Push rows into a Vincjo handler only when the visible set actually changed. */
export function syncTableRows<T extends Row>(
	handler: TableHandler<T>,
	rows: T[],
	opts?: { ignoreEmpty?: boolean }
): void {
	if (opts?.ignoreEmpty && rows.length === 0) return;
	if (lastRows.get(handler) === rows) return;
	const sig = signature(rows);
	if (lastSig.get(handler) === sig) {
		lastRows.set(handler, rows);
		return;
	}
	lastRows.set(handler, rows);
	lastSig.set(handler, sig);
	// Vincjo $state.snapshot() uses structuredClone. Copy into POJOs first.
	handler.setRows(toPlain(rows) as T[]);
}
