<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	import {
		AdminUnauthorizedError,
		createAppVersion,
		type AdminAppVersionUpsertRequest
	} from '$lib/admin-api';

	const listPath = resolve('/app-versions');

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
	let saving = $state(false);
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

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();

		const payload = buildPayload();
		if (!payload) {
			return;
		}

		saving = true;
		formError = '';

		try {
			await createAppVersion(payload);
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
	<title>新建版本</title>
</svelte:head>

<section class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-semibold text-slate-900">新建版本</h1>
		<a href={listPath} class="text-sm text-slate-600 hover:text-slate-900">取消</a>
	</div>

	<form class="mt-6 space-y-4" onsubmit={handleSubmit}>
		<div class="grid gap-4 sm:grid-cols-2">
			<div class="space-y-2">
				<label class="block text-sm font-medium text-slate-700" for="app-version-platform"
					>平台</label
				>
				<select
					id="app-version-platform"
					bind:value={form.platform}
					class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
					disabled={saving}
				>
					<option value="ios">ios</option>
					<option value="android">android</option>
				</select>
			</div>

			<div class="space-y-2">
				<label class="block text-sm font-medium text-slate-700" for="app-version-code">版本号</label
				>
				<input
					id="app-version-code"
					type="number"
					min="1"
					step="1"
					bind:value={form.versionCode}
					class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
					disabled={saving}
				/>
			</div>
		</div>

		<div class="space-y-2">
			<label class="block text-sm font-medium text-slate-700" for="app-version-name">展示版本</label
			>
			<input
				id="app-version-name"
				type="text"
				bind:value={form.versionName}
				class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
				disabled={saving}
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
				disabled={saving}
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
				disabled={saving}
			/>
		</div>

		<label class="inline-flex items-center gap-2 text-sm text-slate-700">
			<input
				type="checkbox"
				bind:checked={form.isForceUpdate}
				class="rounded border-slate-300 text-slate-900 focus:ring-slate-500"
				disabled={saving}
			/>
			<span>强制更新</span>
		</label>

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
