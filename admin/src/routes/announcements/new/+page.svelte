<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	import AnnouncementForm from '$lib/AnnouncementForm.svelte';
	import { buildAnnouncementPayload, createEmptyAnnouncementForm } from '$lib/announcement-form';
	import { AdminUnauthorizedError, createAnnouncement } from '$lib/admin-api';

	const listRoute = '/announcements' as const;
	const listPath = resolve('/announcements');

	let form = $state(createEmptyAnnouncementForm());
	let saving = $state(false);
	let formError = $state('');

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();

		const result = buildAnnouncementPayload(form);
		if (!result.payload) {
			formError = result.error;
			return;
		}

		saving = true;
		formError = '';

		try {
			await createAnnouncement(result.payload);
			void goto(listPath);
		} catch (error) {
			if (error instanceof AdminUnauthorizedError) {
				return;
			}

			formError = error instanceof Error ? error.message : '保存失败';
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>新建公告</title>
</svelte:head>

<div class="admin-page">
	<div class="admin-page-header">
		<h1 class="admin-page-title">新建公告</h1>
		<a href={listPath} class="text-sm text-slate-600 hover:text-slate-900">取消</a>
	</div>

	<section class="admin-card">
		<AnnouncementForm
			bind:form
			disabled={saving}
			{formError}
			cancelRoute={listRoute}
			submitLabel={saving ? '保存中' : '保存'}
			cancelLabel="取消"
			onSubmit={handleSubmit}
		/>
	</section>
</div>
