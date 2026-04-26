import { clearStoredAdminToken, getStoredAdminToken } from '$lib/admin-auth';

export type AdminAnnouncement = {
	id: string;
	title: string;
	content: string;
	isActive: boolean;
	isBanner: boolean;
	createdAt: string;
};

export type AdminAnnouncementUpsertRequest = {
	title: string;
	content: string;
	isActive: boolean;
	isBanner: boolean;
};

export type AdminAppVersion = {
	id: string;
	platform: 'ios' | 'android';
	versionCode: number;
	versionName: string;
	isForceUpdate: boolean;
	releaseNotes: string;
	downloadUrl: string;
	createdAt: string;
};

export type AdminAppVersionUpsertRequest = {
	platform: 'ios' | 'android';
	versionCode: number;
	versionName: string;
	isForceUpdate: boolean;
	releaseNotes: string;
	downloadUrl: string;
};

export class AdminUnauthorizedError extends Error {
	constructor() {
		super('unauthorized');
	}
}

type ErrorResponse = {
	error?: string;
};

async function readErrorMessage(response: Response): Promise<string> {
	try {
		const data = (await response.json()) as ErrorResponse;
		if (typeof data.error === 'string' && data.error.trim()) {
			return data.error;
		}
	} catch {
		return `请求失败 (${response.status})`;
	}

	return `请求失败 (${response.status})`;
}

async function adminRequest<T>(input: string, init?: RequestInit): Promise<T> {
	const token = getStoredAdminToken();
	if (!token) {
		clearStoredAdminToken();
		throw new AdminUnauthorizedError();
	}

	const headers = new Headers(init?.headers);
	headers.set('Authorization', `Bearer ${token}`);

	if (init?.body !== undefined && !headers.has('Content-Type')) {
		headers.set('Content-Type', 'application/json');
	}

	const response = await fetch(input, {
		...init,
		headers
	});

	if (response.status === 401) {
		clearStoredAdminToken();
		throw new AdminUnauthorizedError();
	}

	if (!response.ok) {
		throw new Error(await readErrorMessage(response));
	}

	if (response.status === 204) {
		return undefined as T;
	}

	return (await response.json()) as T;
}

export function listAnnouncements(): Promise<AdminAnnouncement[]> {
	return adminRequest<AdminAnnouncement[]>('/v1/admin/announcements');
}

export function getAnnouncement(id: string): Promise<AdminAnnouncement> {
	return adminRequest<AdminAnnouncement>(`/v1/admin/announcements/${id}`);
}

export function createAnnouncement(
	payload: AdminAnnouncementUpsertRequest
): Promise<AdminAnnouncement> {
	return adminRequest<AdminAnnouncement>('/v1/admin/announcements', {
		method: 'POST',
		body: JSON.stringify(payload)
	});
}

export function updateAnnouncement(
	id: string,
	payload: AdminAnnouncementUpsertRequest
): Promise<AdminAnnouncement> {
	return adminRequest<AdminAnnouncement>(`/v1/admin/announcements/${id}`, {
		method: 'PUT',
		body: JSON.stringify(payload)
	});
}

export function deleteAnnouncement(id: string): Promise<void> {
	return adminRequest<void>(`/v1/admin/announcements/${id}`, {
		method: 'DELETE'
	});
}

export function listAppVersions(): Promise<AdminAppVersion[]> {
	return adminRequest<AdminAppVersion[]>('/v1/admin/app-versions');
}

export function getAppVersion(id: string): Promise<AdminAppVersion> {
	return adminRequest<AdminAppVersion>(`/v1/admin/app-versions/${id}`);
}

export function createAppVersion(payload: AdminAppVersionUpsertRequest): Promise<AdminAppVersion> {
	return adminRequest<AdminAppVersion>('/v1/admin/app-versions', {
		method: 'POST',
		body: JSON.stringify(payload)
	});
}

export function updateAppVersion(
	id: string,
	payload: AdminAppVersionUpsertRequest
): Promise<AdminAppVersion> {
	return adminRequest<AdminAppVersion>(`/v1/admin/app-versions/${id}`, {
		method: 'PUT',
		body: JSON.stringify(payload)
	});
}

export function deleteAppVersion(id: string): Promise<void> {
	return adminRequest<void>(`/v1/admin/app-versions/${id}`, {
		method: 'DELETE'
	});
}
