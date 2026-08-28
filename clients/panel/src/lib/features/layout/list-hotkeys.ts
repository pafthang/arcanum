export function listHotkeys(opts: { oncreate?: () => void; searchSelector?: string }): () => void {
	function onkeydown(event: KeyboardEvent) {
		const target = event.target as HTMLElement | null;
		const typing = Boolean(target?.closest('input, textarea, select, [contenteditable="true"]'));
		if ((event.key === 'n' || event.key === 'N') && !typing && !event.metaKey && !event.ctrlKey && !event.altKey) {
			event.preventDefault();
			opts.oncreate?.();
		}
	}
	window.addEventListener('keydown', onkeydown);
	return () => window.removeEventListener('keydown', onkeydown);
}
