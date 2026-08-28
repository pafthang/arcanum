export type Keybind = {
	mod?: boolean;
	shift?: boolean;
	alt?: boolean;
	key: string;
};

export type ShortcutEntry = {
	id: string;
	title: string;
	keys: string;
};

export type ShortcutGroup = {
	name: string;
	shortcuts: ShortcutEntry[];
};

class ShortcutsUi {
	open = $state(false);

	show(): void {
		this.open = true;
	}

	toggle(): void {
		this.open = !this.open;
	}
}

export const shortcutsUi = new ShortcutsUi();

export function isApple(): boolean {
	return typeof navigator !== 'undefined' && /Mac|iPhone|iPad/.test(navigator.platform);
}

export function formatKeybind(bind: Keybind): string {
	const mac = isApple();
	const parts: string[] = [];
	if (bind.mod) parts.push(mac ? '⌘' : 'Ctrl');
	if (bind.shift) parts.push(mac ? '⇧' : 'Shift');
	if (bind.alt) parts.push(mac ? '⌥' : 'Alt');
	const key = bind.key.length === 1 ? bind.key.toUpperCase() : bind.key;
	parts.push(key);
	return mac ? parts.join('') : parts.join('+');
}

export function searchChord(): string {
	return formatKeybind({ mod: true, key: 'P' });
}

export function commandChord(): string {
	return formatKeybind({ mod: true, shift: true, key: 'P' });
}

export function shortcutsChord(): string {
	return formatKeybind({ mod: true, key: '/' });
}

export function shortcutCatalog(): ShortcutGroup[] {
	return [
		{
			name: 'Search',
			shortcuts: [
				{ id: 'search.open', title: 'Open search', keys: searchChord() },
				{ id: 'search.commands', title: 'Run command', keys: commandChord() },
				{ id: 'search.focus', title: 'Focus search', keys: '/' },
				{ id: 'search.users', title: 'Find user', keys: '@' },
				{ id: 'search.sessions', title: 'Inspect sessions', keys: '#' },
				{ id: 'search.prefix', title: 'Command prefix', keys: '>' }
			]
		},
		{
			name: 'Navigation',
			shortcuts: [
				{ id: 'nav.sidebar', title: 'Toggle sidebar', keys: formatKeybind({ mod: true, key: 'B' }) },
				{ id: 'nav.shortcuts', title: 'View shortcuts', keys: shortcutsChord() }
			]
		},
		{
			name: 'Lists',
			shortcuts: [{ id: 'list.create', title: 'Create item', keys: 'N' }]
		}
	];
}

export function filterShortcutGroups(query: string): ShortcutGroup[] {
	const q = query.trim().toLowerCase();
	if (!q) return shortcutCatalog();
	return shortcutCatalog()
		.map((group) => ({
			...group,
			shortcuts: group.shortcuts.filter(
				(item) =>
					item.title.toLowerCase().includes(q) ||
					item.keys.toLowerCase().includes(q) ||
					group.name.toLowerCase().includes(q)
			)
		}))
		.filter((group) => group.shortcuts.length > 0);
}

export function bindShortcutSheet(): () => void {
	function onkeydown(event: KeyboardEvent) {
		if (!(event.metaKey || event.ctrlKey) || event.altKey) return;
		if (event.key !== '/' && event.code !== 'Slash') return;
		event.preventDefault();
		shortcutsUi.toggle();
	}
	window.addEventListener('keydown', onkeydown);
	return () => window.removeEventListener('keydown', onkeydown);
}
