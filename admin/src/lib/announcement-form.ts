import type { AdminAnnouncement, AdminAnnouncementUpsertRequest } from '$lib/admin-api';

export type AnnouncementFormState = {
	title: string;
	content: string;
	isActive: boolean;
	isBanner: boolean;
};

export function createEmptyAnnouncementForm(): AnnouncementFormState {
	return {
		title: '',
		content: '',
		isActive: true,
		isBanner: false
	};
}

export function fromAdminAnnouncement(item: AdminAnnouncement): AnnouncementFormState {
	return {
		title: item.title,
		content: item.content,
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

	return {
		payload: {
			title,
			content,
			isActive: form.isActive,
			isBanner: form.isBanner
		},
		error: ''
	};
}
