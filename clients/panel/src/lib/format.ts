const BYTE_UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'] as const;

export function formatBytes(n: number | string | null | undefined): string {
	const value = typeof n === 'string' ? Number(n) : (n ?? 0);
	if (!Number.isFinite(value) || value === 0) return '0 B';
	let v = Math.abs(value);
	let i = 0;
	while (v >= 1024 && i < BYTE_UNITS.length - 1) {
		v /= 1024;
		i += 1;
	}
	const digits = v < 10 && i > 0 ? 2 : 1;
	return `${v.toFixed(digits)} ${BYTE_UNITS[i]}`;
}

export function formatTrafficLimit(n: number | string | null | undefined): string {
	const value = typeof n === 'string' ? Number(n) : (n ?? 0);
	if (!Number.isFinite(value) || value <= 0) return 'Unlimited';
	return formatBytes(value);
}

export function parsePrettyBytes(value: string | undefined): number {
	if (!value) return 0;
	const match = value.trim().match(/^([+-]?\d+(?:\.\d+)?)\s*([KMGTP]?B)?$/i);
	if (!match) return Number(value) || 0;
	const amount = Number(match[1]);
	const unit = (match[2] || 'B').toUpperCase();
	const mul: Record<string, number> = {
		B: 1,
		KB: 1024,
		MB: 1024 ** 2,
		GB: 1024 ** 3,
		TB: 1024 ** 4,
		PB: 1024 ** 5
	};
	return amount * (mul[unit] ?? 1);
}

export function trafficPercent(used: number | null | undefined, limit: number | null | undefined): number | null {
	const cap = Number(limit) || 0;
	if (cap <= 0) return null;
	return Math.min(100, Math.round(((Number(used) || 0) / cap) * 100));
}

export type ExpireTone = 'ok' | 'soon' | 'expired' | 'none';

export function expireMeta(s: string | null | undefined): { label: string; detail: string; tone: ExpireTone } {
	if (!s) return { label: '—', detail: '', tone: 'none' };
	const d = new Date(s);
	if (Number.isNaN(d.getTime())) return { label: s, detail: '', tone: 'none' };
	const diff = d.getTime() - Date.now();
	const days = Math.round(diff / 86_400_000);
	const detail = d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
	if (diff < 0) {
		const ago = Math.abs(days);
		return { label: ago <= 1 ? 'Expired' : `Expired ${ago}d ago`, detail, tone: 'expired' };
	}
	if (days <= 3) return { label: days <= 0 ? 'Expires today' : `In ${days}d`, detail, tone: 'soon' };
	if (days < 60) return { label: `In ${days}d`, detail, tone: 'ok' };
	return { label: detail, detail: '', tone: 'ok' };
}

export function toDateInput(iso: string | null | undefined): string {
	if (!iso) return '';
	const d = new Date(iso);
	if (Number.isNaN(d.getTime())) return '';
	return d.toISOString().slice(0, 10);
}

export function dateInputToExpireIso(date: string): string {
	if (!date) return new Date().toISOString();
	return new Date(`${date}T23:59:59.000Z`).toISOString();
}

export type TrafficUnit = 'MB' | 'GB' | 'TB';

export function splitTraffic(bytes: number): { amount: string; unit: TrafficUnit } {
	if (!bytes) return { amount: '0', unit: 'GB' };
	if (bytes >= 1024 ** 4 && bytes % 1024 ** 4 === 0) return { amount: String(bytes / 1024 ** 4), unit: 'TB' };
	if (bytes >= 1024 ** 3 && bytes % (1024 ** 2) === 0) {
		const gb = bytes / 1024 ** 3;
		if (Number.isInteger(gb) || gb >= 1) return { amount: String(Number(gb.toFixed(gb >= 10 ? 0 : 1))), unit: 'GB' };
	}
	const mb = bytes / 1024 ** 2;
	return { amount: String(Number(mb.toFixed(mb >= 10 ? 0 : 1))), unit: 'MB' };
}

export function trafficToBytes(amount: string, unit: TrafficUnit): number {
	const n = Number(amount);
	if (!Number.isFinite(n) || n <= 0) return 0;
	const mul = { MB: 1024 ** 2, GB: 1024 ** 3, TB: 1024 ** 4 };
	return Math.round(n * mul[unit]);
}

export function bandwidthDelta(current?: string, previous?: string): { text: string; tone: 'up' | 'down' | 'neutral' } | undefined {
	if (!previous) return undefined;
	const now = parsePrettyBytes(current);
	const prev = parsePrettyBytes(previous);
	if (prev <= 0) return undefined;
	const pct = ((now - prev) / prev) * 100;
	const rounded = Math.abs(pct) >= 10 ? pct.toFixed(0) : pct.toFixed(1);
	if (Math.abs(pct) < 0.5) return { text: 'flat vs prev', tone: 'neutral' };
	return {
		text: `${pct > 0 ? '+' : ''}${rounded}% vs prev`,
		tone: pct > 0 ? 'up' : 'down'
	};
}

export function formatDate(s: string | null | undefined): string {
	if (!s) return '—';
	const d = new Date(s);
	if (Number.isNaN(d.getTime())) return s;
	return d.toLocaleString();
}

export function formatNumber(n: number | null | undefined): string {
	return new Intl.NumberFormat().format(n ?? 0);
}

export function formatUptime(seconds: number | null | undefined): string {
	if (seconds == null || !Number.isFinite(seconds) || seconds < 0) return '—';
	const days = Math.floor(seconds / 86400);
	const hours = Math.floor((seconds % 86400) / 3600);
	const minutes = Math.floor((seconds % 3600) / 60);
	if (days > 0) return hours > 0 ? `${days}d ${hours}h` : `${days}d`;
	if (hours > 0) return `${hours}h ${minutes}m`;
	return `${minutes}m`;
}

export function plusDaysIso(days: number): string {
	const d = new Date();
	d.setUTCDate(d.getUTCDate() + days);
	return d.toISOString();
}

export function dateRange(days = 7): { start: string; end: string } {
	const end = new Date();
	const start = new Date();
	start.setUTCDate(end.getUTCDate() - days);
	const iso = (d: Date) => d.toISOString().slice(0, 10);
	return { start: iso(start), end: iso(end) };
}
