import type { Snippet } from 'svelte';

export type PageSave = {
	onclick: () => void | Promise<void>;
	disabled?: () => boolean;
	pending?: () => boolean;
};

export type PageIconAction = PageSave & {
	label: string;
	icon: 'sync' | 'restart';
};

export type PageSearch = {
	placeholder?: string;
	fields?: Array<PropertyKey | ((row: never) => unknown)>;
	table?: object;
	value?: () => string;
	setValue?: (value: string) => void;
	oninput?: () => void;
};

export type PageTableView = {
	columns: { index: number; name?: string; isVisible?: boolean; toggle?: () => void }[];
};

export type PageFilter = {
	value: () => string;
	options: { value: string; label: string }[];
	onselect: (value: string) => void;
};

export type PageToolbar = {
	table: object;
	view?: PageTableView;
	onviewchange?: () => void;
	filter?: PageFilter;
};

type Chrome = {
	title?: string | null;
	description?: string | null;
	actions?: Snippet | null;
	create?: { label: string; onclick: () => void } | null;
	action?: PageIconAction | null;
	save?: PageSave | null;
	search?: PageSearch | null;
	toolbar?: PageToolbar | null;
};

class PageChrome {
	title = $state<string | null>(null);
	description = $state<string | null>(null);
	actions = $state<Snippet | null>(null);
	create = $state<{ label: string; onclick: () => void } | null>(null);
	action = $state<PageIconAction | null>(null);
	save = $state<PageSave | null>(null);
	search = $state<PageSearch | null>(null);
	toolbar = $state<PageToolbar | null>(null);
	#generation = 0;

	set(next: Chrome = {}): number {
		this.#generation += 1;
		this.title = next.title ?? null;
		this.description = next.description ?? null;
		this.actions = next.actions ?? null;
		this.create = next.create ?? null;
		this.action = next.action ?? null;
		this.save = next.save ?? null;
		this.search = next.search ?? null;
		this.toolbar = next.toolbar ?? null;
		return this.#generation;
	}

	clear(token?: number): void {
		const expected = token ?? this.#generation;
		queueMicrotask(() => {
			if (this.#generation !== expected) return;
			this.title = null;
			this.description = null;
			this.actions = null;
			this.create = null;
			this.action = null;
			this.save = null;
			this.search = null;
			this.toolbar = null;
		});
	}
}

export const pageChrome = new PageChrome();
