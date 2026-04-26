<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { onMount } from 'svelte';

	import {
		AdminUnauthorizedError,
		deleteAnnouncement,
		getAnnouncement,
		updateAnnouncement,
		type AdminAnnouncementUpsertRequest
	} from '$lib/admin-api';

	const listPath = resolve('/announcements');
	const announcementId = $derived(page.url.searchParams.get('id')?.trim() ?? '');

	let form = $state<AdminAnnouncementUpsertRequest>({
		title: '',
		content: '',
		isActive: true,
		isBanner: false
	});
	let loading = $state(true);
	let saving = $state(false);
	let deleting = $state(false);
	let notFound = $state(false);
	let loadError = $state('');
	let formError = $state('');

	function buildPayload(): AdminAnnouncementUpsertRequest | null {
		const title = form.title.trim();
		const content = form.content.trim();
		if (!title || !content) {
			formError = '请填写完整内容';
			return null;
		}

		return {
			title,
			content,
			isActive: form.isActive,
			isBanner: form.isBanner
		};
	}

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
			form = {
				title: item.title,
				content: item.content,
				isActive: item.isActive,
				isBanner: item.isBanner
			};
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

		const payload = buildPayload();
		if (!payload) {
			return;
		}

		if (browser && !window.confirm('确认保存修改？')) {
			return;
		}

		saving = true;
		formError = '';

		try {
			await updateAnnouncement(announcementId, payload);
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
			<form class="space-y-4" onsubmit={handleSubmit}>
				<div class="space-y-2">
					<label class="block text-sm font-medium text-slate-700" for="announcement-title"
						>标题</label
					>
					<input
						id="announcement-title"
						type="text"
						bind:value={form.title}
						class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
						disabled={saving || deleting}
					/>
				</div>

				<div class="space-y-2">
					<label class="block text-sm font-medium text-slate-700" for="announcement-content"
						>内容</label
					>
					<textarea
						id="announcement-content"
						rows="8"
						bind:value={form.content}
						class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
						disabled={saving || deleting}
					></textarea>
				</div>

				<div class="flex flex-wrap gap-4">
					<label class="inline-flex items-center gap-2 text-sm text-slate-700">
						<input
							type="checkbox"
							bind:checked={form.isActive}
							class="rounded border-slate-300 text-slate-900 focus:ring-slate-500"
							disabled={saving || deleting}
						/>
						<span>生效</span>
					</label>

					<label class="inline-flex items-center gap-2 text-sm text-slate-700">
						<input
							type="checkbox"
							bind:checked={form.isBanner}
							class="rounded border-slate-300 text-slate-900 focus:ring-slate-500"
							disabled={saving || deleting}
						/>
						<span>Banner</span>
					</label>
				</div>

				{#if formError}
					<p class="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600">
						{formError}
					</p>
				{/if}

				<div class="flex flex-wrap gap-3">
					<button
						type="submit"
						class="rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
						disabled={saving || deleting}
					>
						{saving ? '保存中' : '保存'}
					</button>

					<a
						href={listPath}
						class="rounded-md border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
					>
						取消
					</a>

					<button
						type="button"
						class="rounded-md border border-red-300 px-4 py-2 text-sm font-medium text-red-600 hover:bg-red-50 disabled:cursor-not-allowed disabled:border-red-200 disabled:text-red-300"
						onclick={handleDelete}
						disabled={saving || deleting}
					>
						{deleting ? '删除中' : '删除'}
					</button>
				</div>
			</form>
		{/if}
	</section>
</div>
