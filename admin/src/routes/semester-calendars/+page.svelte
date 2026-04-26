<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';

	import {
		AdminUnauthorizedError,
		deleteSemesterCalendar,
		listSemesterCalendars,
		type AdminSemesterCalendar
	} from '$lib/admin-api';

	const newPath = resolve('/semester-calendars/new');

	function formatDate(value: string): string {
		return value.slice(0, 10);
	}

	function formatTime(value: string): string {
		return new Date(value).toLocaleString('zh-CN', {
			hour12: false
		});
	}

	function handleEdit(semesterCode: string) {
		void goto(resolve(`/semester-calendars/edit?semesterCode=${encodeURIComponent(semesterCode)}`));
	}

	let calendars = $state<AdminSemesterCalendar[]>([]);
	let loading = $state(true);
	let deletingSemesterCode = $state('');
	let loadError = $state('');

	async function loadCalendars() {
		loading = true;
		loadError = '';

		try {
			calendars = await listSemesterCalendars();
		} catch (error) {
			if (error instanceof AdminUnauthorizedError) {
				return;
			}

			loadError = error instanceof Error ? error.message : '加载失败';
		} finally {
			loading = false;
		}
	}

	async function handleDelete(item: AdminSemesterCalendar) {
		if (browser && !window.confirm(`确认删除 ${item.semesterCode}？`)) {
			return;
		}

		deletingSemesterCode = item.semesterCode;
		loadError = '';

		try {
			await deleteSemesterCalendar(item.semesterCode);
			await loadCalendars();
		} catch (error) {
			if (error instanceof AdminUnauthorizedError) {
				return;
			}

			loadError = error instanceof Error ? error.message : '删除失败';
		} finally {
			deletingSemesterCode = '';
		}
	}

	onMount(() => {
		void loadCalendars();
	});
</script>

<svelte:head>
	<title>校历管理</title>
</svelte:head>

<div class="admin-page">
	<div class="admin-page-header">
		<h1 class="admin-page-title">校历管理</h1>
		<a
			href={newPath}
			class="rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800"
		>
			新建
		</a>
	</div>

	<section class="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
		<div class="overflow-x-auto">
			<table class="min-w-full divide-y divide-slate-200 text-sm">
				<thead class="bg-slate-50 text-left text-slate-500">
					<tr>
						<th class="px-4 py-3 font-medium">学期代码</th>
						<th class="px-4 py-3 font-medium">标题</th>
						<th class="px-4 py-3 font-medium">副标题</th>
						<th class="px-4 py-3 font-medium">校历开始</th>
						<th class="px-4 py-3 font-medium">校历结束</th>
						<th class="px-4 py-3 font-medium">学期开始</th>
						<th class="px-4 py-3 font-medium">学期结束</th>
						<th class="px-4 py-3 font-medium">创建时间</th>
						<th class="px-4 py-3 font-medium">操作</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-200">
					{#if loading}
						<tr>
							<td colspan="9" class="px-4 py-6 text-center text-slate-500">加载中</td>
						</tr>
					{:else if calendars.length === 0}
						<tr>
							<td colspan="9" class="px-4 py-6 text-center text-slate-500">暂无数据</td>
						</tr>
					{:else}
						{#each calendars as item (item.semesterCode)}
							<tr>
								<td class="px-4 py-3 whitespace-nowrap text-slate-900">{item.semesterCode}</td>
								<td class="px-4 py-3 whitespace-nowrap text-slate-900">{item.title}</td>
								<td class="px-4 py-3 whitespace-nowrap text-slate-600">{item.subtitle}</td>
								<td class="px-4 py-3 whitespace-nowrap text-slate-600">
									{formatDate(item.calendarStart)}
								</td>
								<td class="px-4 py-3 whitespace-nowrap text-slate-600">
									{formatDate(item.calendarEnd)}
								</td>
								<td class="px-4 py-3 whitespace-nowrap text-slate-600">
									{formatDate(item.semesterStart)}
								</td>
								<td class="px-4 py-3 whitespace-nowrap text-slate-600">
									{formatDate(item.semesterEnd)}
								</td>
								<td class="px-4 py-3 whitespace-nowrap text-slate-600">
									{formatTime(item.createdAt)}
								</td>
								<td class="px-4 py-3">
									<div class="flex gap-3">
										<button
											type="button"
											class="text-sm text-slate-700 hover:text-slate-900"
											onclick={() => handleEdit(item.semesterCode)}
										>
											编辑
										</button>
										<button
											type="button"
											class="text-sm text-red-600 hover:text-red-700 disabled:text-red-300"
											onclick={() => handleDelete(item)}
											disabled={deletingSemesterCode === item.semesterCode}
										>
											{deletingSemesterCode === item.semesterCode ? '删除中' : '删除'}
										</button>
									</div>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>

		{#if loadError}
			<p class="border-t border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">{loadError}</p>
		{/if}
	</section>
</div>
