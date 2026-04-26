<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	import SemesterCalendarForm from '$lib/SemesterCalendarForm.svelte';
	import { AdminUnauthorizedError, createSemesterCalendar } from '$lib/admin-api';
	import {
		buildSemesterCalendarPayload,
		createEmptySemesterCalendarForm
	} from '$lib/semester-calendar-form';

	const listRoute = '/semester-calendars' as const;
	const listPath = resolve('/semester-calendars');

	let form = $state(createEmptySemesterCalendarForm());
	let saving = $state(false);
	let formError = $state('');

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();

		const result = buildSemesterCalendarPayload(form);
		if (!result.payload) {
			formError = result.error;
			return;
		}

		saving = true;
		formError = '';

		try {
			await createSemesterCalendar(result.payload);
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
	<title>新建校历</title>
</svelte:head>

<div class="admin-page">
	<div class="admin-page-header">
		<h1 class="admin-page-title">新建校历</h1>
		<a href={listPath} class="text-sm text-slate-600 hover:text-slate-900">取消</a>
	</div>

	<section class="admin-card">
		<SemesterCalendarForm
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
