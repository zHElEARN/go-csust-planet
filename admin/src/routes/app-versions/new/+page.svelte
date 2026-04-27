<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	import AppVersionForm from '$lib/AppVersionForm.svelte';
	import { buildAppVersionPayload, createEmptyAppVersionForm } from '$lib/app-version-form';
	import { AdminUnauthorizedError, createAppVersion } from '$lib/admin-api';

	const listRoute = '/app-versions' as const;
	const listPath = resolve('/app-versions');

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

		saving = true;
		formError = '';

		try {
			await createAppVersion(result.payload);
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

<div class="admin-page">
	<div class="admin-page-header">
		<h1 class="admin-page-title">新建版本</h1>
		<a href={listPath} class="text-sm text-slate-600 hover:text-slate-900">取消</a>
	</div>

	<section class="admin-card">
		<AppVersionForm
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
