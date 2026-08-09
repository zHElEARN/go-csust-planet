<script lang="ts">
	import { browser } from '$app/environment';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	import AppVersionScopeNotice from '$lib/AppVersionScopeNotice.svelte';
	import {
		AdminUnauthorizedError,
		deleteAppVersion,
		listAppVersions,
		type AdminAppVersion
	} from '$lib/admin-api';
	import {
		appVersionEditPath,
		appVersionListPath,
		appVersionNewPath,
		getAppVersionScopeDetails,
		parseAppVersionScope,
		type AppVersionScope
	} from '$lib/app-version-scope';

	const currentPath = appVersionListPath('current');
	const legacyPath = appVersionListPath('legacy');
	const scope = $derived(parseAppVersionScope(page.url.searchParams.get('scope')));
	const newPath = $derived(appVersionNewPath(scope));

	function formatTime(value: string): string {
		return new Date(value).toLocaleString('zh-CN', {
			hour12: false
		});
	}

	let versions = $state<AdminAppVersion[]>([]);
	let loading = $state(true);
	let deletingId = $state('');
	let loadError = $state('');
	let loadRequestID = 0;

	async function loadVersions(targetScope: AppVersionScope) {
		const requestID = ++loadRequestID;
		loading = true;
		loadError = '';

		try {
			const result = await listAppVersions(targetScope);
			if (requestID === loadRequestID) {
				versions = result;
			}
		} catch (error) {
			if (error instanceof AdminUnauthorizedError) {
				return;
			}

			if (requestID === loadRequestID) {
				loadError = error instanceof Error ? error.message : '加载失败';
			}
		} finally {
			if (requestID === loadRequestID) {
				loading = false;
			}
		}
	}

	async function handleDelete(item: AdminAppVersion) {
		const targetScope = scope;
		const targetDetails = getAppVersionScopeDetails(targetScope);
		if (browser && !window.confirm(`确认从${targetDetails.label}删除版本 ${item.versionName}？`)) {
			return;
		}

		deletingId = item.id;
		loadError = '';

		try {
			await deleteAppVersion(targetScope, item.id);
			if (scope === targetScope) {
				await loadVersions(targetScope);
			}
		} catch (error) {
			if (error instanceof AdminUnauthorizedError) {
				return;
			}

			loadError = error instanceof Error ? error.message : '删除失败';
		} finally {
			deletingId = '';
		}
	}

	$effect(() => {
		void loadVersions(scope);
	});
</script>

<svelte:head>
	<title>{scope === 'legacy' ? '旧版迁移' : '新版发布'} · 版本管理</title>
</svelte:head>

<div class="admin-page">
	<div class="admin-page-header">
		<h1 class="admin-page-title">版本管理</h1>
		<a
			href={resolve(newPath)}
			class="rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800"
		>
			新建
		</a>
	</div>

	<nav class="flex gap-2 border-b border-slate-200" aria-label="版本库">
		<a
			href={resolve(currentPath)}
			aria-current={scope === 'current' ? 'page' : undefined}
			class={`border-b-2 px-4 py-2 text-sm font-medium ${
				scope === 'current'
					? 'border-slate-900 text-slate-900'
					: 'border-transparent text-slate-500 hover:border-slate-300 hover:text-slate-700'
			}`}
		>
			新版发布
		</a>
		<a
			href={resolve(legacyPath)}
			aria-current={scope === 'legacy' ? 'page' : undefined}
			class={`border-b-2 px-4 py-2 text-sm font-medium ${
				scope === 'legacy'
					? 'border-amber-600 text-amber-800'
					: 'border-transparent text-slate-500 hover:border-slate-300 hover:text-slate-700'
			}`}
		>
			旧版迁移
		</a>
	</nav>

	<AppVersionScopeNotice {scope} />

	<section class="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
		<div class="overflow-x-auto">
			<table class="min-w-full divide-y divide-slate-200 text-sm">
				<thead class="bg-slate-50 text-left text-slate-500">
					<tr>
						<th class="px-4 py-3 font-medium">平台</th>
						<th class="px-4 py-3 font-medium">版本号</th>
						<th class="px-4 py-3 font-medium">展示版本</th>
						<th class="px-4 py-3 font-medium">强更</th>
						<th class="px-4 py-3 font-medium">下载地址</th>
						<th class="px-4 py-3 font-medium">创建时间</th>
						<th class="px-4 py-3 font-medium">操作</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-200">
					{#if loading}
						<tr>
							<td colspan="7" class="px-4 py-6 text-center text-slate-500">加载中</td>
						</tr>
					{:else if versions.length === 0}
						<tr>
							<td colspan="7" class="px-4 py-6 text-center text-slate-500">暂无数据</td>
						</tr>
					{:else}
						{#each versions as item (item.id)}
							<tr>
								<td class="px-4 py-3 text-slate-600">{item.platform}</td>
								<td class="px-4 py-3 text-slate-600">{item.versionCode}</td>
								<td class="px-4 py-3 text-slate-900">{item.versionName}</td>
								<td class="px-4 py-3 text-slate-600">{item.isForceUpdate ? '是' : '否'}</td>
								<td class="max-w-56 truncate px-4 py-3 text-slate-600">{item.downloadUrl}</td>
								<td class="px-4 py-3 text-slate-600">{formatTime(item.createdAt)}</td>
								<td class="px-4 py-3">
									<div class="flex gap-3">
										<a
											href={resolve(appVersionEditPath(scope, item.id))}
											class="text-sm text-slate-700 hover:text-slate-900"
										>
											编辑
										</a>
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
