export function errorMessage(err: unknown, fallback: string): string {
	if (err instanceof Error && err.message) return err.message;
	return fallback;
}

export function parseJsonObject(
	text: string
): { ok: true; value: Record<string, unknown> } | { ok: false; error: string } {
	try {
		const value = JSON.parse(text) as unknown;
		if (!value || typeof value !== 'object' || Array.isArray(value)) {
			return { ok: false, error: 'Config must be a JSON object.' };
		}
		return { ok: true, value: value as Record<string, unknown> };
	} catch {
		return { ok: false, error: 'Config must be valid JSON.' };
	}
}

export function parseSnippetArray(
	text: string
): { ok: true; value: unknown[] } | { ok: false; error: string } {
	try {
		const value = JSON.parse(text) as unknown;
		if (!Array.isArray(value) || value.length === 0) {
			return { ok: false, error: 'Snippet must be a non-empty JSON array of objects.' };
		}
		for (const item of value) {
			if (!item || typeof item !== 'object' || Array.isArray(item) || Object.keys(item as object).length === 0) {
				return { ok: false, error: 'Each snippet item must be a non-empty object. Empty {} is not allowed.' };
			}
		}
		return { ok: true, value };
	} catch {
		return { ok: false, error: 'Snippet must be valid JSON.' };
	}
}

export function profileHasInbound(config: Record<string, unknown>): boolean {
	return Array.isArray(config.inbounds) && config.inbounds.length > 0;
}

export const DEFAULT_PROFILE_CONFIG = `{
  "log": { "loglevel": "warning" },
  "inbounds": [
    {
      "tag": "vless-in",
      "protocol": "vless",
      "port": 443,
      "settings": { "clients": [] },
      "streamSettings": { "network": "tcp", "security": "tls" }
    }
  ],
  "outbounds": [{ "protocol": "freedom", "tag": "direct" }]
}`;

export const DEFAULT_SNIPPET = `[
  {
    "tag": "example-in",
    "protocol": "vless",
    "port": 443,
    "settings": { "clients": [] }
  }
]`;
