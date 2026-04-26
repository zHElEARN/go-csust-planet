<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';

	import {
		AdminUnauthorizedError,
		deleteAnnouncement,
		listAnnouncements,
		type AdminAnnouncement
	} from '$lib/admin-api';

	const newPath = resolve('/announcements/new');

	function formatTime(value: string): string {
		return new Date(value).toLocaleString('zh-CN', {
			hour12: false
		});
	}

	function handleEdit(id: string) {
		void goto(resolve(`/announcements/edit?id=${encodeURIComponent(id)}`));
	}

	let announcements = $state<AdminAnnouncement[]>([]);
	let loading = $state(true);
	let deletingId = $state('');
	let loadError = $state('');

	async function loadAnnouncements() {
		loading = true;
		loadError = '';

		try {
			announcements = await listAnnouncements();
		} catch (error) {
			if (error instanceof AdminUnauthorizedError) {
				return;
			}

			loadError = error instanceof Error ? error.message : '加载失败';
		} finally {
			loading = false;
		}
	}

	async function handleDelete(item: AdminAnnouncement) {
		if (browser && !window.confirm('确认删除？')) {
			return;
		}

		deletingId = item.id;
		loadError = '';

		try {
			await deleteAnnouncement(item.id);
			await loadAnnouncements();
		} catch (error) {
			if (error instanceof AdminUnauthorizedError) {
				return;
			}

			loadError = error instanceof Error ? error.message : '删除失败';
		} finally {
			deletingId = '';
		}
	}

	onMount(() => {
		void loadAnnouncements();
	});
</script>

<svelte:head>
	<title>公告管理</title>
</svelte:head>

<div class="admin-page">
	<div class="admin-page-header">
		<h1 class="admin-page-title">公告管理</h1>
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
						<th class="px-4 py-3 font-medium">标题</th>
						<th class="px-4 py-3 font-medium">生效</th>
						<th class="px-4 py-3 font-medium">Banner</th>
						<th class="px-4 py-3 font-medium">创建时间</th>
						<th class="px-4 py-3 font-medium">操作</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-200">
					{#if loading}
						<tr>
							<td colspan="5" class="px-4 py-6 text-center text-slate-500">加载中</td>
						</tr>
					{:else if announcements.length === 0}
						<tr>
							<td colspan="5" class="px-4 py-6 text-center text-slate-500">暂无数据</td>
						</tr>
					{:else}
						{#each announcements as item (item.id)}
							<tr>
								<td class="px-4 py-3 text-slate-900">{item.title}</td>
								<td class="px-4 py-3 text-slate-600">{item.isActive ? '是' : '否'}</td>
								<td class="px-4 py-3 text-slate-600">{item.isBanner ? '是' : '否'}</td>
								<td class="px-4 py-3 text-slate-600">{formatTime(item.createdAt)}</td>
								<td class="px-4 py-3">
									<div class="flex gap-3">
										<button
											type="button"
											class="text-sm text-slate-700 hover:text-slate-900"
											onclick={() => handleEdit(item.id)}
										>
											编辑
										</button>
										<button
											type="button"
											class="text-sm text-red-600 hover:text-red-700 disabled:text-red-300"
											onclick={() => handleDelete(item)}
											disabled={deletingId === item.id}
										>
											{deletingId === item.id ? '删除中' : '删除'}
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
