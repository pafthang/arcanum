const STORAGE_KEY = 'remnawave.content-width';

export type ContentWidth = 'contained' | 'wide';

function read(): ContentWidth {
	if (typeof localStorage === 'undefined') return 'contained';
	return localStorage.getItem(STORAGE_KEY) === 'wide' ? 'wide' : 'contained';
}

class ContentWidthState {
	value = $state<ContentWidth>(read());

	set(next: ContentWidth) {
		this.value = next;
		if (typeof localStorage !== 'undefined') localStorage.setItem(STORAGE_KEY, next);
	}
}

export const contentWidth = new ContentWidthState();
