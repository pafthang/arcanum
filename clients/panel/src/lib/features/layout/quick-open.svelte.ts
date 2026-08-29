import type { ComponentType } from 'svelte';
import { Radio, Zap } from 'lucide-svelte';
import { SETTINGS_NAV } from '$lib/features/settings/nav';
import { shortcutsChord, shortcutsUi } from '$lib/features/layout/shortcuts.svelte';
import { NAV } from '$lib/nav';
import { rw } from '$lib/rw';
import type { User, UserList } from '@arcanum/ts-client';

export type QuickOpenPage = {
	label: string;
	href: string;
	hint: string;
	icon: ComponentType;
};

export type QuickOpenCommand = {
	id: string;
	label: string;
	hint: string;
	prefix?: string;
	href?: string;
	icon: ComponentType;
	onRun?: () => void;
};

export type SessionTarget = { kind: 'user' | 'node'; id: string };

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

class QuickOpen {
	nonce = $state(0);
	prefix = '';

	open(prefix = ''): void {
		this.prefix = prefix;
		this.nonce += 1;
	}
}

export const quickOpen = new QuickOpen();

export function pageCatalog(): QuickOpenPage[] {
	const pages: QuickOpenPage[] = [];
	for (const section of NAV) {
		for (const item of section.items) {
			pages.push({ label: item.label, href: item.href, hint: section.header, icon: item.icon });
		}
	}
	for (const group of SETTINGS_NAV) {
		for (const item of group.items) {
			pages.push({
				label: item.label,
				href: item.href,
				hint: `Settings · ${group.label}`,
				icon: item.icon
			});
		}
	}
	return pages;
}

export function filterPages(query: string, limit = 8): QuickOpenPage[] {
	const pages = pageCatalog();
	const q = query.trim().toLowerCase();
	if (!q) return pages.slice(0, limit);
	return pages
		.filter(
			(page) =>
				page.label.toLowerCase().includes(q) ||
				page.hint.toLowerCase().includes(q) ||
				page.href.toLowerCase().includes(q)
		)
		.slice(0, limit);
}

export function resolveUserBody(value: string): { id: number } | { shortUuid: string } | { username: string } {
	const id = Number(value);
	if (Number.isFinite(id) && String(id) === value) return { id };
	if (value.length <= 16 && !value.includes(' ')) return { shortUuid: value };
	return { username: value };
}

export async function searchUsers(query: string): Promise<User[]> {
	const q = query.trim();
	if (q.length < 2) return [];
	const data = await rw.http.get<UserList>('/api/users', { start: 0, size: 8, search: q });
	return data.users ?? [];
}

export function parseSessionTarget(raw: string): SessionTarget | null {
	const q = raw.trim();
	if (!q) return null;
	const node = q.match(/^(?:node|n)\s+(\S+)$/i);
	if (node) return { kind: 'node', id: node[1] };
	if (UUID.test(q)) return { kind: 'node', id: q };
	return { kind: 'user', id: q };
}

export function sessionHref(target: SessionTarget): string {
	const params = new URLSearchParams({ kind: target.kind, id: target.id });
	return `/dashboard/system/logs?tab=sessions&${params}`;
}

export function commandCatalog(): QuickOpenCommand[] {
	return [
		{
			id: 'sessions.user',
			label: 'Sessions: Inspect by user',
			hint: '#',
			prefix: '#',
			icon: Radio
		},
		{
			id: 'sessions.node',
			label: 'Sessions: Inspect by node',
			hint: '#node',
			prefix: '#node ',
			icon: Radio
		},
		{
			id: 'sessions.drop',
			label: 'Sessions: Drop connections',
			hint: 'Tools',
			href: '/dashboard/system/logs?tab=sessions&drop=1',
			icon: Radio
		},
		{
			id: 'app.shortcuts',
			label: 'View shortcuts',
			hint: shortcutsChord(),
			icon: Zap,
			onRun: () => shortcutsUi.show()
		}
	];
}

export function filterCommands(query: string, limit = 8): QuickOpenCommand[] {
	const commands = commandCatalog();
	const q = query.trim().toLowerCase();
	if (!q) return commands.slice(0, limit);
	return commands
		.filter(
			(command) =>
				command.label.toLowerCase().includes(q) ||
				command.hint.toLowerCase().includes(q) ||
				command.id.toLowerCase().includes(q)
		)
		.slice(0, limit);
}
