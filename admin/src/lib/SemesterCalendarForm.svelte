<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		createEmptyCustomWeekRange,
		createEmptyNote,
		type SemesterCalendarFormState
	} from '$lib/semester-calendar-form';

	let {
		form = $bindable<SemesterCalendarFormState>(),
		disabled = false,
		formError = '',
		cancelRoute,
		submitLabel,
		cancelLabel,
		deleteLabel = '删除',
		onSubmit,
		onDelete
	}: {
		form: SemesterCalendarFormState;
		disabled?: boolean;
		formError?: string;
		cancelRoute: '/semester-calendars';
		submitLabel: string;
		cancelLabel: string;
		deleteLabel?: string;
		onSubmit: (event: SubmitEvent) => void | Promise<void>;
		onDelete?: () => void | Promise<void>;
	} = $props();

	function addNote() {
		form.notes = [...form.notes, createEmptyNote()];
	}

	function removeNote(index: number) {
		form.notes = form.notes.filter((_, currentIndex) => currentIndex !== index);
	}

	function addCustomWeekRange() {
		form.customWeekRanges = [...form.customWeekRanges, createEmptyCustomWeekRange()];
	}

	function removeCustomWeekRange(index: number) {
		form.customWeekRanges = form.customWeekRanges.filter(
			(_, currentIndex) => currentIndex !== index
		);
	}
</script>

<form class="space-y-6" onsubmit={onSubmit}>
	<div class="grid gap-4 sm:grid-cols-2">
		<div class="space-y-2">
			<label class="block text-sm font-medium text-slate-700" for="semester-calendar-code"
				>学期代码</label
			>
			<input
				id="semester-calendar-code"
				type="text"
				bind:value={form.semesterCode}
				class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
				{disabled}
			/>
		</div>

		<div class="space-y-2">
			<label class="block text-sm font-medium text-slate-700" for="semester-calendar-title"
				>标题</label
			>
			<input
				id="semester-calendar-title"
				type="text"
				bind:value={form.title}
				class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
				{disabled}
			/>
		</div>
	</div>

	<div class="space-y-2">
		<label class="block text-sm font-medium text-slate-700" for="semester-calendar-subtitle"
			>副标题</label
		>
		<input
			id="semester-calendar-subtitle"
			type="text"
			bind:value={form.subtitle}
			class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
			{disabled}
		/>
	</div>

	<div class="grid gap-4 sm:grid-cols-2">
		<div class="space-y-2">
			<label class="block text-sm font-medium text-slate-700" for="semester-calendar-start"
				>校历开始日期</label
			>
			<input
				id="semester-calendar-start"
				type="date"
				bind:value={form.calendarStart}
				class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
				{disabled}
			/>
		</div>

		<div class="space-y-2">
			<label class="block text-sm font-medium text-slate-700" for="semester-calendar-end"
				>校历结束日期</label
			>
			<input
				id="semester-calendar-end"
				type="date"
				bind:value={form.calendarEnd}
				class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
				{disabled}
			/>
		</div>

		<div class="space-y-2">
			<label
				class="block text-sm font-medium text-slate-700"
				for="semester-calendar-semester-start"
			>
				学期开始日期
			</label>
			<input
				id="semester-calendar-semester-start"
				type="date"
				bind:value={form.semesterStart}
				class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
				{disabled}
			/>
		</div>

		<div class="space-y-2">
			<label class="block text-sm font-medium text-slate-700" for="semester-calendar-semester-end">
				学期结束日期
			</label>
			<input
				id="semester-calendar-semester-end"
				type="date"
				bind:value={form.semesterEnd}
				class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
				{disabled}
			/>
		</div>
	</div>

	<div class="space-y-4">
		<div class="flex items-center justify-between gap-3">
			<h2 class="text-base font-medium text-slate-900">备注</h2>
			<button
				type="button"
				class="rounded-md border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
				onclick={addNote}
				{disabled}
			>
				新增备注
			</button>
		</div>

		<div class="space-y-3">
			{#each form.notes as item, index (index)}
				<div class="rounded-lg border border-slate-200 p-4">
					<div class="grid gap-4 sm:grid-cols-[8rem_minmax(0,1fr)_auto]">
						<div class="space-y-2">
							<label class="block text-sm font-medium text-slate-700" for={`note-row-${index}`}
								>行号</label
							>
							<input
								id={`note-row-${index}`}
								type="number"
								min="1"
								step="1"
								bind:value={item.row}
								class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
								{disabled}
							/>
						</div>

						<div class="space-y-2">
							<label class="block text-sm font-medium text-slate-700" for={`note-content-${index}`}>
								内容
							</label>
							<input
								id={`note-content-${index}`}
								type="text"
								bind:value={item.content}
								class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
								{disabled}
							/>
						</div>

						<div class="flex items-end gap-3">
							<label class="inline-flex items-center gap-2 text-sm text-slate-700">
								<input
									type="checkbox"
									bind:checked={item.needNumber}
									class="rounded border-slate-300 text-slate-900 focus:ring-slate-500"
									{disabled}
								/>
								<span>显示序号</span>
							</label>

							<button
								type="button"
								class="text-sm text-red-600 hover:text-red-700 disabled:text-red-300"
								onclick={() => removeNote(index)}
								{disabled}
							>
								删除
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	</div>

	<div class="space-y-4">
		<div class="flex items-center justify-between gap-3">
			<h2 class="text-base font-medium text-slate-900">自定义周次范围</h2>
			<button
				type="button"
				class="rounded-md border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
				onclick={addCustomWeekRange}
				{disabled}
			>
				新增范围
			</button>
		</div>

		<div class="space-y-3">
			{#each form.customWeekRanges as item, index (index)}
				<div class="rounded-lg border border-slate-200 p-4">
					<div class="grid gap-4 sm:grid-cols-[8rem_8rem_minmax(0,1fr)_auto]">
						<div class="space-y-2">
							<label
								class="block text-sm font-medium text-slate-700"
								for={`range-start-row-${index}`}
							>
								开始行
							</label>
							<input
								id={`range-start-row-${index}`}
								type="number"
								min="1"
								step="1"
								bind:value={item.startRow}
								class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
								{disabled}
							/>
						</div>

						<div class="space-y-2">
							<label
								class="block text-sm font-medium text-slate-700"
								for={`range-end-row-${index}`}
							>
								结束行
							</label>
							<input
								id={`range-end-row-${index}`}
								type="number"
								min="1"
								step="1"
								bind:value={item.endRow}
								class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
								{disabled}
							/>
						</div>

						<div class="space-y-2">
							<label
								class="block text-sm font-medium text-slate-700"
								for={`range-content-${index}`}
							>
								内容
							</label>
							<input
								id={`range-content-${index}`}
								type="text"
								bind:value={item.content}
								class="block w-full rounded-md border-slate-300 text-sm text-slate-900 focus:border-slate-500 focus:ring-slate-500"
								{disabled}
							/>
						</div>

						<div class="flex items-end">
							<button
								type="button"
								class="text-sm text-red-600 hover:text-red-700 disabled:text-red-300"
								onclick={() => removeCustomWeekRange(index)}
								{disabled}
							>
								删除
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
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

		<button
			type="button"
			onclick={() => goto(resolve(cancelRoute))}
			class="rounded-md border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
		>
			{cancelLabel}
		</button>

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
