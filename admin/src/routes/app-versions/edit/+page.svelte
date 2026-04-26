<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { onMount } from 'svelte';

	import {
		AdminUnauthorizedError,
		deleteAppVersion,
		getAppVersion,
		updateAppVersion,
		type AdminAppVersionUpsertRequest
	} from '$lib/admin-api';

	const listPath = resolve('/app-versions');
	const versionId = $derived(page.url.searchParams.get('id')?.trim() ?? '');

	let form = $state<{
		platform: 'ios' | 'android';
		versionCode: string;
		versionName: string;
		isForceUpdate: boolean;
		releaseNotes: string;
		downloadUrl: string;
	}>({
		platform: 'ios',
		versionCode: '',
		versionName: '',
		isForceUpdate: false,
		releaseNotes: '',
		downloadUrl: ''
	});
	let loading = $state(true);
	let saving = $state(false);
	let deleting = $state(false);
	let notFound = $state(false);
	let loadError = $state('');
	let formError = $state('');

	function buildPayload(): AdminAppVersionUpsertRequest | null {
		const versionCode = Number(form.versionCode.trim());
		const versionName = form.versionName.trim();
		const releaseNotes = form.releaseNotes.trim();
		const downloadUrl = form.downloadUrl.trim();

		if (
			!form.versionCode.trim() ||
			!Number.isInteger(versionCode) ||
			versionCode <= 0 ||
			!versionName ||
			!releaseNotes ||
			!downloadUrl
		) {
			formError = '请填写完整内容';
			return null;
		}

		return {
			platform: form.platform,
			versionCode,
			versionName,
			isForceUpdate: form.isForceUpdate,
			releaseNotes,
			downloadUrl
		};
	}

	async function loadVersion() {
		if (!versionId) {
			notFound = true;
			loading = false;
			return;
		}

		loading = true;
		notFound = false;
		loadError = '';

		try {
			const item = await getAppVersion(versionId);
			form = {
				platform: item.platform,
				versionCode: String(item.versionCode),
				versionName: item.versionName,
				isForceUpdate: item.isForceUpdate,
				releaseNotes: item.releaseNotes,
				downloadUrl: item.downloadUrl
			};
		} catch (error) {
			if (error instanceof AdminUnauthorizedError) {
				return;
			}

			const message = error instanceof Error ? error.message : '加载失败';
			if (message === '未找到该版本') {
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
		if (!versionId) {
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
			await updateAppVersion(versionId, payload);
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
		if (!versionId) {
			notFound = true;
			return;
		}

		if (browser && !window.confirm('确认删除？')) {
			return;
		}

		deleting = true;
		formError = '';

		try {
			await deleteAppVersion(versionId);
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
		void loadVersion();
	});
</script>

<svelte:head>
	<title>编辑版本</title>
</svelte:head>

<div class="admin-page">
	<div class="admin-page-header">
		<h1 class="admin-page-title">编辑版本</h1>
		<a href={listPath} class="text-sm text-slate-600 hover:text-slate-900">返回</a>
	</div>

	<section class="admin-card">
		{#if loading}
			<p class="text-sm text-slate-500">加载中</p>
		{:else if notFound}
			<div class="space-y-4">
				<p class="text-sm text-slate-500">未找到该版本</p>
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
				<div class="grid gap-4 sm:grid-cols-2">
					<div class="space-y-2">
						<label class="block text-sm font-medium text-slate-700" for="app-version-platform"
							>平台</label
						>
						<select
							id="app-version-platform"
							bind:value={form.platform}
							class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
							disabled={saving || deleting}
						>
							<option value="ios">ios</option>
							<option value="android">android</option>
						</select>
					</div>

					<div class="space-y-2">
						<label class="block text-sm font-medium text-slate-700" for="app-version-code"
							>版本号</label
						>
						<input
							id="app-version-code"
							type="number"
							min="1"
							step="1"
							bind:value={form.versionCode}
							class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
							disabled={saving || deleting}
						/>
					</div>
				</div>

				<div class="space-y-2">
					<label class="block text-sm font-medium text-slate-700" for="app-version-name"
						>展示版本</label
					>
					<input
						id="app-version-name"
						type="text"
						bind:value={form.versionName}
						class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
						disabled={saving || deleting}
					/>
				</div>

				<div class="space-y-2">
					<label class="block text-sm font-medium text-slate-700" for="app-version-release-notes"
						>更新日志</label
					>
					<textarea
						id="app-version-release-notes"
						rows="6"
						bind:value={form.releaseNotes}
						class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
						disabled={saving || deleting}
					></textarea>
				</div>

				<div class="space-y-2">
					<label class="block text-sm font-medium text-slate-700" for="app-version-download-url"
						>下载地址</label
					>
					<input
						id="app-version-download-url"
						type="url"
						bind:value={form.downloadUrl}
						class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
						disabled={saving || deleting}
					/>
				</div>

				<label class="inline-flex items-center gap-2 text-sm text-slate-700">
					<input
						type="checkbox"
						bind:checked={form.isForceUpdate}
						class="rounded border-slate-300 text-slate-900 focus:ring-slate-500"
						disabled={saving || deleting}
					/>
					<span>强制更新</span>
				</label>

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
