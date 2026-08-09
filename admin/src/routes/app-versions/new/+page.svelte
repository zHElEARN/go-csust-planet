<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	import AppVersionForm from '$lib/AppVersionForm.svelte';
	import AppVersionScopeNotice from '$lib/AppVersionScopeNotice.svelte';
	import { buildAppVersionPayload, createEmptyAppVersionForm } from '$lib/app-version-form';
	import { AdminUnauthorizedError, createAppVersion } from '$lib/admin-api';
	import {
		appVersionListPath,
		getAppVersionScopeDetails,
		parseAppVersionScope
	} from '$lib/app-version-scope';

	const scope = $derived(parseAppVersionScope(page.url.searchParams.get('scope')));
	const scopeDetails = $derived(getAppVersionScopeDetails(scope));
	const listPath = $derived(appVersionListPath(scope));

	let form = $state(createEmptyAppVersionForm());
	let saving = $state(false);
	let formError = $state('');

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();

		const result = buildAppVersionPayload(form);
		if (!result.payload) {
			formError = result.error;
			return;
		}

		if (scope === 'legacy' && browser && !window.confirm('确认保存到旧版迁移库？')) {
			return;
		}

		saving = true;
		formError = '';

		try {
			await createAppVersion(scope, result.payload);
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
</script>

<svelte:head>
	<title>新建{scope === 'legacy' ? '旧版迁移' : '新版发布'}版本</title>
</svelte:head>

<div class="admin-page">
	<div class="admin-page-header">
		<div>
			<h1 class="admin-page-title">新建{scope === 'legacy' ? '旧版迁移' : '新版发布'}版本</h1>
			<p class="mt-1 text-sm text-slate-500">目标：{scopeDetails.label}</p>
		</div>
		<a href={resolve(listPath)} class="text-sm text-slate-600 hover:text-slate-900">取消</a>
	</div>

	<AppVersionScopeNotice {scope} />

	<section class="admin-card">
		<AppVersionForm
			bind:form
			disabled={saving}
			{formError}
			cancelPath={listPath}
			submitLabel={saving ? '保存中' : '保存'}
			cancelLabel="取消"
			onSubmit={handleSubmit}
		/>
	</section>
</div>
