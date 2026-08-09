import { clearStoredAdminToken, getStoredAdminToken } from '$lib/admin-auth';
import type { AppVersionScope } from '$lib/app-version-scope';

export type AnnouncementPlatform = 'ios' | 'android' | 'all';

export type AdminAnnouncement = {
	id: string;
	title: string;
	content: string;
	platform: AnnouncementPlatform;
	isActive: boolean;
	isBanner: boolean;
	createdAt: string;
};

export type AdminAnnouncementUpsertRequest = {
	title: string;
	content: string;
	platform: AnnouncementPlatform;
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

export type CalendarNote = {
	row: number;
	content: string;
	needNumber?: boolean;
};

export type CustomWeekRange = {
	startRow: number;
	endRow: number;
	content: string;
};

export type AdminSemesterCalendar = {
	semesterCode: string;
	title: string;
	subtitle: string;
	calendarStart: string;
	calendarEnd: string;
	semesterStart: string;
	semesterEnd: string;
	notes: CalendarNote[];
	customWeekRanges: CustomWeekRange[];
	createdAt: string;
};

export type AdminSemesterCalendarUpsertRequest = {
	semesterCode: string;
	title: string;
	subtitle: string;
	calendarStart: string;
	calendarEnd: string;
	semesterStart: string;
	semesterEnd: string;
	notes: CalendarNote[];
	customWeekRanges: CustomWeekRange[];
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

function appVersionAdminPath(scope: AppVersionScope): string {
	return scope === 'legacy' ? '/v1/admin/app-versions' : '/v2/admin/app-versions';
}

export function listAppVersions(scope: AppVersionScope): Promise<AdminAppVersion[]> {
	return adminRequest<AdminAppVersion[]>(appVersionAdminPath(scope));
}

export function getAppVersion(scope: AppVersionScope, id: string): Promise<AdminAppVersion> {
	return adminRequest<AdminAppVersion>(`${appVersionAdminPath(scope)}/${id}`);
}

export function createAppVersion(
	scope: AppVersionScope,
	payload: AdminAppVersionUpsertRequest
): Promise<AdminAppVersion> {
	return adminRequest<AdminAppVersion>(appVersionAdminPath(scope), {
		method: 'POST',
		body: JSON.stringify(payload)
	});
}

export function updateAppVersion(
	scope: AppVersionScope,
	id: string,
	payload: AdminAppVersionUpsertRequest
): Promise<AdminAppVersion> {
	return adminRequest<AdminAppVersion>(`${appVersionAdminPath(scope)}/${id}`, {
		method: 'PUT',
		body: JSON.stringify(payload)
	});
}

export function deleteAppVersion(scope: AppVersionScope, id: string): Promise<void> {
	return adminRequest<void>(`${appVersionAdminPath(scope)}/${id}`, {
		method: 'DELETE'
	});
}

export function listSemesterCalendars(): Promise<AdminSemesterCalendar[]> {
	return adminRequest<AdminSemesterCalendar[]>('/v1/admin/semester-calendars');
}

export function getSemesterCalendar(semesterCode: string): Promise<AdminSemesterCalendar> {
	return adminRequest<AdminSemesterCalendar>(`/v1/admin/semester-calendars/${semesterCode}`);
}

export function createSemesterCalendar(
	payload: AdminSemesterCalendarUpsertRequest
): Promise<AdminSemesterCalendar> {
	return adminRequest<AdminSemesterCalendar>('/v1/admin/semester-calendars', {
		method: 'POST',
		body: JSON.stringify(payload)
	});
}

export function updateSemesterCalendar(
	semesterCode: string,
	payload: AdminSemesterCalendarUpsertRequest
): Promise<AdminSemesterCalendar> {
	return adminRequest<AdminSemesterCalendar>(`/v1/admin/semester-calendars/${semesterCode}`, {
		method: 'PUT',
		body: JSON.stringify(payload)
	});
}

export function deleteSemesterCalendar(semesterCode: string): Promise<void> {
	return adminRequest<void>(`/v1/admin/semester-calendars/${semesterCode}`, {
		method: 'DELETE'
	});
}
