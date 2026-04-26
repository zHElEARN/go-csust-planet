import { browser } from '$app/environment';

const ADMIN_TOKEN_KEY = 'admin_bearer_token';
const ADMIN_VALIDATE_ENDPOINT = '/v1/admin/announcements';

export function getStoredAdminToken(): string {
	if (!browser) {
		return '';
	}

	return sessionStorage.getItem(ADMIN_TOKEN_KEY)?.trim() ?? '';
}

export function setStoredAdminToken(token: string): void {
	if (!browser) {
		return;
	}

	sessionStorage.setItem(ADMIN_TOKEN_KEY, token.trim());
}

export function clearStoredAdminToken(): void {
	if (!browser) {
		return;
	}

	sessionStorage.removeItem(ADMIN_TOKEN_KEY);
}

export async function validateAdminToken(token: string): Promise<boolean> {
	const normalizedToken = token.trim();
	if (!normalizedToken) {
		return false;
	}

	const response = await fetch(ADMIN_VALIDATE_ENDPOINT, {
		headers: {
			Authorization: `Bearer ${normalizedToken}`
		}
	});

	if (response.status === 401) {
		return false;
	}

	if (!response.ok) {
		throw new Error(`admin token validation failed with status ${response.status}`);
	}

	return true;
}
