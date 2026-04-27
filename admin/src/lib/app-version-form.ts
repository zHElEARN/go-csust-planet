import type { AdminAppVersion, AdminAppVersionUpsertRequest } from '$lib/admin-api';

export type AppVersionFormState = {
	platform: 'ios' | 'android';
	versionCode: number;
	versionName: string;
	isForceUpdate: boolean;
	releaseNotes: string;
	downloadUrl: string;
};

export function createEmptyAppVersionForm(): AppVersionFormState {
	return {
		platform: 'ios',
		versionCode: NaN,
		versionName: '',
		isForceUpdate: false,
		releaseNotes: '',
		downloadUrl: ''
	};
}

export function fromAdminAppVersion(item: AdminAppVersion): AppVersionFormState {
	return {
		platform: item.platform,
		versionCode: item.versionCode,
		versionName: item.versionName,
		isForceUpdate: item.isForceUpdate,
		releaseNotes: item.releaseNotes,
		downloadUrl: item.downloadUrl
	};
}

export function buildAppVersionPayload(
	form: AppVersionFormState
): { payload: AdminAppVersionUpsertRequest; error: '' } | { payload: null; error: string } {
	const versionCode = form.versionCode;
	const versionName = form.versionName.trim();
	const releaseNotes = form.releaseNotes.trim();
	const downloadUrl = form.downloadUrl.trim();

	if (
		isNaN(versionCode) ||
		!Number.isInteger(versionCode) ||
		versionCode <= 0 ||
		!versionName ||
		!releaseNotes ||
		!downloadUrl
	) {
		return { payload: null, error: '请填写完整内容' };
	}

	return {
		payload: {
			platform: form.platform,
			versionCode,
			versionName,
			isForceUpdate: form.isForceUpdate,
			releaseNotes,
			downloadUrl
		},
		error: ''
	};
}
