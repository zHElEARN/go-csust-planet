<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { onMount } from 'svelte';

	import AnnouncementForm from '$lib/AnnouncementForm.svelte';
	import {
		buildAnnouncementPayload,
		createEmptyAnnouncementForm,
		fromAdminAnnouncement
	} from '$lib/announcement-form';
	import {
		AdminUnauthorizedError,
		deleteAnnouncement,
		getAnnouncement,
		updateAnnouncement
	} from '$lib/admin-api';

	const listRoute = '/announcements' as const;
	const listPath = resolve('/announcements');
	const announcementId = $derived(page.url.searchParams.get('id')?.trim() ?? '');

	let form = $state(createEmptyAnnouncementForm());
	let loading = $state(true);
	let saving = $state(false);
	let deleting = $state(false);
	let notFound = $state(false);
	let loadError = $state('');
	let formError = $state('');

	async function loadAnnouncement() {
		if (!announcementId) {
			notFound = true;
			loading = false;
			return;
		}

		loading = true;
		notFound = false;
		loadError = '';

		try {
			const item = await getAnnouncement(announcementId);
			form = fromAdminAnnouncement(item);
		} catch (error) {
			if (error instanceof AdminUnauthorizedError) {
				return;
			}

			const message = error instanceof Error ? error.message : '加载失败';
			if (message === '未找到该公告') {
				notFound = true;
			} else {
				loadError = message;
			}
		} finally {
			loading = false;
		}
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (!announcementId) {
			notFound = true;
			return;
		}

		const result = buildAnnouncementPayload(form);
		if (!result.payload) {
			formError = result.error;
			return;
		}

		if (browser && !window.confirm('确认保存修改？')) {
			return;
		}

		saving = true;
		formError = '';

		try {
			await updateAnnouncement(announcementId, result.payload);
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

	async function handleDelete() {
		if (!announcementId) {
			notFound = true;
			return;
		}

		if (browser && !window.confirm('确认删除？')) {
			return;
		}

		deleting = true;
		formError = '';

		try {
			await deleteAnnouncement(announcementId);
			void goto(listPath);
		} catch (error) {
			if (error instanceof AdminUnauthorizedError) {
				return;
			}

			formError = error instanceof Error ? error.message : '删除失败';
		} finally {
			deleting = false;
		}
	}

	onMount(() => {
		void loadAnnouncement();
	});
</script>

<svelte:head>
	<title>编辑公告</title>
</svelte:head>

<div class="admin-page">
	<div class="admin-page-header">
		<h1 class="admin-page-title">编辑公告</h1>
		<a href={listPath} class="text-sm text-slate-600 hover:text-slate-900">返回</a>
	</div>

	<section class="admin-card">
		{#if loading}
			<p class="text-sm text-slate-500">加载中</p>
		{:else if notFound}
			<div class="space-y-4">
				<p class="text-sm text-slate-500">未找到该公告</p>
				<a
					href={listPath}
					class="inline-flex rounded-md border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
				>
					返回列表
				</a>
			</div>
		{:else if loadError}
			<div class="space-y-4">
				<p class="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600">
					{loadError}
				</p>
				<a
					href={listPath}
					class="inline-flex rounded-md border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
				>
					返回列表
				</a>
			</div>
		{:else}
			<AnnouncementForm
				bind:form
				disabled={saving || deleting}
				{formError}
				cancelRoute={listRoute}
				submitLabel={saving ? '保存中' : '保存'}
				cancelLabel="取消"
				deleteLabel={deleting ? '删除中' : '删除'}
				onSubmit={handleSubmit}
				onDelete={handleDelete}
			/>
		{/if}
	</section>
</div>
