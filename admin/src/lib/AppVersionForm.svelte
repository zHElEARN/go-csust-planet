<script lang="ts">
	import { resolve } from '$app/paths';

	import type { AppVersionFormState } from '$lib/app-version-form';

	let {
		form = $bindable<AppVersionFormState>(),
		disabled = false,
		formError = '',
		cancelRoute,
		submitLabel,
		cancelLabel,
		deleteLabel = '删除',
		onSubmit,
		onDelete
	}: {
		form: AppVersionFormState;
		disabled?: boolean;
		formError?: string;
		cancelRoute: '/app-versions';
		submitLabel: string;
		cancelLabel: string;
		deleteLabel?: string;
		onSubmit: (event: SubmitEvent) => void | Promise<void>;
		onDelete?: () => void | Promise<void>;
	} = $props();
</script>

<form class="space-y-4" onsubmit={onSubmit}>
	<div class="grid gap-4 sm:grid-cols-2">
		<div class="space-y-2">
			<label class="block text-sm font-medium text-slate-700" for="app-version-platform">平台</label
			>
			<select
				id="app-version-platform"
				bind:value={form.platform}
				class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
				{disabled}
			>
				<option value="ios">ios</option>
				<option value="android">android</option>
			</select>
		</div>

		<div class="space-y-2">
			<label class="block text-sm font-medium text-slate-700" for="app-version-code">版本号</label>
			<input
				id="app-version-code"
				type="number"
				min="1"
				step="1"
				bind:value={form.versionCode}
				class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
				{disabled}
			/>
		</div>
	</div>

	<div class="space-y-2">
		<label class="block text-sm font-medium text-slate-700" for="app-version-name">展示版本</label>
		<input
			id="app-version-name"
			type="text"
			bind:value={form.versionName}
			class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
			{disabled}
		/>
	</div>

	<div class="space-y-2">
		<label class="block text-sm font-medium text-slate-700" for="app-version-release-notes">
			更新日志
		</label>
		<textarea
			id="app-version-release-notes"
			rows="6"
			bind:value={form.releaseNotes}
			class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
			{disabled}
		></textarea>
	</div>

	<div class="space-y-2">
		<label class="block text-sm font-medium text-slate-700" for="app-version-download-url">
			下载地址
		</label>
		<input
			id="app-version-download-url"
			type="url"
			bind:value={form.downloadUrl}
			class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
			{disabled}
		/>
	</div>

	<label class="inline-flex items-center gap-2 text-sm text-slate-700">
		<input
			type="checkbox"
			bind:checked={form.isForceUpdate}
			class="rounded border-slate-300 text-slate-900 focus:ring-slate-500"
			{disabled}
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
