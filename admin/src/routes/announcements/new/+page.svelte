<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	import {
		AdminUnauthorizedError,
		createAnnouncement,
		type AdminAnnouncementUpsertRequest
	} from '$lib/admin-api';

	const listPath = resolve('/announcements');

	let form = $state<AdminAnnouncementUpsertRequest>({
		title: '',
		content: '',
		isActive: true,
		isBanner: false
	});
	let saving = $state(false);
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

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();

		const payload = buildPayload();
		if (!payload) {
			return;
		}

		saving = true;
		formError = '';

		try {
			await createAnnouncement(payload);
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

<section class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-semibold text-slate-900">新建公告</h1>
		<a href={listPath} class="text-sm text-slate-600 hover:text-slate-900">取消</a>
	</div>

	<form class="mt-6 space-y-4" onsubmit={handleSubmit}>
		<div class="space-y-2">
			<label class="block text-sm font-medium text-slate-700" for="announcement-title">标题</label>
			<input
				id="announcement-title"
				type="text"
				bind:value={form.title}
				class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
				disabled={saving}
			/>
		</div>

		<div class="space-y-2">
			<label class="block text-sm font-medium text-slate-700" for="announcement-content">内容</label
			>
			<textarea
				id="announcement-content"
				rows="8"
				bind:value={form.content}
				class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
				disabled={saving}
			></textarea>
		</div>

		<div class="flex flex-wrap gap-4">
			<label class="inline-flex items-center gap-2 text-sm text-slate-700">
				<input
					type="checkbox"
					bind:checked={form.isActive}
					class="rounded border-slate-300 text-slate-900 focus:ring-slate-500"
					disabled={saving}
				/>
				<span>生效</span>
			</label>

			<label class="inline-flex items-center gap-2 text-sm text-slate-700">
				<input
					type="checkbox"
					bind:checked={form.isBanner}
					class="rounded border-slate-300 text-slate-900 focus:ring-slate-500"
					disabled={saving}
				/>
				<span>Banner</span>
			</label>
		</div>

		{#if formError}
			<p class="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600">
				{formError}
			</p>
		{/if}

		<div class="flex gap-3">
			<button
				type="submit"
				class="rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
				disabled={saving}
			>
				{saving ? '保存中' : '保存'}
			</button>

			<a
				href={listPath}
				class="rounded-md border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
			>
				取消
			</a>
		</div>
	</form>
</section>
