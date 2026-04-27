<script lang="ts">
	import { resolve } from '$app/paths';

	import type { AnnouncementFormState } from '$lib/announcement-form';

	let {
		form = $bindable<AnnouncementFormState>(),
		disabled = false,
		formError = '',
		cancelRoute,
		submitLabel,
		cancelLabel,
		deleteLabel = '删除',
		onSubmit,
		onDelete
	}: {
		form: AnnouncementFormState;
		disabled?: boolean;
		formError?: string;
		cancelRoute: '/announcements';
		submitLabel: string;
		cancelLabel: string;
		deleteLabel?: string;
		onSubmit: (event: SubmitEvent) => void | Promise<void>;
		onDelete?: () => void | Promise<void>;
	} = $props();
</script>

<form class="space-y-4" onsubmit={onSubmit}>
	<div class="space-y-2">
		<label class="block text-sm font-medium text-slate-700" for="announcement-title">标题</label>
		<input
			id="announcement-title"
			type="text"
			bind:value={form.title}
			class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
			{disabled}
		/>
	</div>

	<div class="space-y-2">
		<label class="block text-sm font-medium text-slate-700" for="announcement-content">内容</label>
		<textarea
			id="announcement-content"
			rows="8"
			bind:value={form.content}
			class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
			{disabled}
		></textarea>
	</div>

	<div class="flex flex-wrap gap-4">
		<label class="inline-flex items-center gap-2 text-sm text-slate-700">
			<input
				type="checkbox"
				bind:checked={form.isActive}
				class="rounded border-slate-300 text-slate-900 focus:ring-slate-500"
				{disabled}
			/>
			<span>生效</span>
		</label>

		<label class="inline-flex items-center gap-2 text-sm text-slate-700">
			<input
				type="checkbox"
				bind:checked={form.isBanner}
				class="rounded border-slate-300 text-slate-900 focus:ring-slate-500"
				{disabled}
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
			{disabled}
		>
			{submitLabel}
		</button>

		<a
			href={resolve(cancelRoute)}
			class="rounded-md border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
		>
			{cancelLabel}
		</a>

		{#if onDelete}
			<button
				type="button"
				class="rounded-md border border-red-300 px-4 py-2 text-sm font-medium text-red-600 hover:bg-red-50 disabled:cursor-not-allowed disabled:border-red-200 disabled:text-red-300"
				onclick={onDelete}
				{disabled}
			>
				{deleteLabel}
			</button>
		{/if}
	</div>
</form>
