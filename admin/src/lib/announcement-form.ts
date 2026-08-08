import type {
	AdminAnnouncement,
	AdminAnnouncementUpsertRequest,
	AnnouncementPlatform
} from '$lib/admin-api';

export const announcementPlatformOptions: ReadonlyArray<{
	value: AnnouncementPlatform;
	label: string;
}> = [
	{ value: 'ios', label: '仅 iOS' },
	{ value: 'android', label: '仅安卓' },
	{ value: 'all', label: '全部平台' }
];

export type AnnouncementFormState = {
	title: string;
	content: string;
	platform: AnnouncementPlatform | '';
	isActive: boolean;
	isBanner: boolean;
};

export function createEmptyAnnouncementForm(): AnnouncementFormState {
	return {
		title: '',
		content: '',
		platform: '',
		isActive: true,
		isBanner: false
	};
}

export function fromAdminAnnouncement(item: AdminAnnouncement): AnnouncementFormState {
	return {
		title: item.title,
		content: item.content,
		platform: item.platform,
		isActive: item.isActive,
		isBanner: item.isBanner
	};
}

export function buildAnnouncementPayload(
	form: AnnouncementFormState
): { payload: AdminAnnouncementUpsertRequest; error: '' } | { payload: null; error: string } {
	const title = form.title.trim();
	const content = form.content.trim();
	if (!title || !content) {
		return { payload: null, error: '请填写完整内容' };
	}
	if (!form.platform) {
		return { payload: null, error: '请选择发布平台' };
	}

	return {
		payload: {
			title,
			content,
			platform: form.platform,
			isActive: form.isActive,
			isBanner: form.isBanner
		},
		error: ''
	};
}
