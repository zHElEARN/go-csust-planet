<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { onMount } from 'svelte';

	import AppVersionForm from '$lib/AppVersionForm.svelte';
	import AppVersionScopeNotice from '$lib/AppVersionScopeNotice.svelte';
	import {
		buildAppVersionPayload,
		createEmptyAppVersionForm,
		fromAdminAppVersion
	} from '$lib/app-version-form';
	import {
		AdminUnauthorizedError,
		deleteAppVersion,
		getAppVersion,
		updateAppVersion
	} from '$lib/admin-api';
	import {
		appVersionListPath,
		getAppVersionScopeDetails,
		parseAppVersionScope
	} from '$lib/app-version-scope';

	const scope = $derived(parseAppVersionScope(page.url.searchParams.get('scope')));
	const scopeDetails = $derived(getAppVersionScopeDetails(scope));
	const listPath = $derived(appVersionListPath(scope));
	const versionId = $derived(page.url.searchParams.get('id')?.trim() ?? '');

	let form = $state(createEmptyAppVersionForm());
	let loading = $state(true);
	let saving = $state(false);
	let deleting = $state(false);
	let notFound = $state(false);
	let loadError = $state('');
	let formError = $state('');

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
			const item = await getAppVersion(scope, versionId);
			form = fromAdminAppVersion(item);
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

		const result = buildAppVersionPayload(form);
		if (!result.payload) {
			formError = result.error;
			return;
		}

		if (browser && !window.confirm(`确认保存修改到${scopeDetails.label}？`)) {
			return;
		}

		saving = true;
		formError = '';

		try {
			await updateAppVersion(scope, versionId, result.payload);
			void goto(resolve(listPath));
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

		if (browser && !window.confirm(`确认从${scopeDetails.label}删除该版本？`)) {
			return;
		}

		deleting = true;
		formError = '';

		try {
			await deleteAppVersion(scope, versionId);
			void goto(resolve(listPath));
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
	<title>编辑{scope === 'legacy' ? '旧版迁移' : '新版发布'}版本</title>
</svelte:head>

<div class="admin-page">
	<div class="admin-page-header">
		<div>
			<h1 class="admin-page-title">编辑{scope === 'legacy' ? '旧版迁移' : '新版发布'}版本</h1>
			<p class="mt-1 text-sm text-slate-500">目标：{scopeDetails.label}</p>
		</div>
		<a href={resolve(listPath)} class="text-sm text-slate-600 hover:text-slate-900">返回</a>
	</div>

	<AppVersionScopeNotice {scope} />

	<section class="admin-card">
		{#if loading}
			<p class="text-sm text-slate-500">加载中</p>
		{:else if notFound}
			<div class="space-y-4">
				<p class="text-sm text-slate-500">未找到该版本</p>
				<a
					href={resolve(listPath)}
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
					href={resolve(listPath)}
					class="inline-flex rounded-md border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
				>
					返回列表
				</a>
			</div>
		{:else}
			<AppVersionForm
				bind:form
				disabled={saving || deleting}
				{formError}
				cancelPath={listPath}
				submitLabel={saving ? '保存中' : '保存'}
				cancelLabel="取消"
				deleteLabel={deleting ? '删除中' : '删除'}
				onSubmit={handleSubmit}
				onDelete={handleDelete}
			/>
		{/if}
	</section>
</div>
