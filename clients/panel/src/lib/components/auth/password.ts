const UPPER = 'ABCDEFGHJKLMNPQRSTUVWXYZ';
const LOWER = 'abcdefghijkmnopqrstuvwxyz';
const DIGIT = '23456789';
const ALL = UPPER + LOWER + DIGIT;

export const PASSWORD_POLICY = 'At least 24 characters, with uppercase, lowercase and a number.';

export function generatePassword(length = 28): string {
	const bytes = crypto.getRandomValues(new Uint8Array(length));
	const chars = Array.from(bytes, (b, i) => {
		if (i === 0) return UPPER[b % UPPER.length];
		if (i === 1) return LOWER[b % LOWER.length];
		if (i === 2) return DIGIT[b % DIGIT.length];
		return ALL[b % ALL.length];
	});
	for (let i = chars.length - 1; i > 0; i--) {
		const j = bytes[i] % (i + 1);
		[chars[i], chars[j]] = [chars[j], chars[i]];
	}
	return chars.join('');
}

export function passwordLooksValid(value: string): boolean {
	return value.length >= 24 && /[A-Z]/.test(value) && /[a-z]/.test(value) && /\d/.test(value);
}
