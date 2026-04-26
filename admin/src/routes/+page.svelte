<script lang="ts">
	import { onMount } from 'svelte';

	import {
		clearStoredAdminToken,
		getStoredAdminToken,
		setStoredAdminToken,
		validateAdminToken
	} from '$lib/admin-auth';

	type AuthState = 'checking' | 'unauthenticated' | 'submitting' | 'authenticated';

	let authState = $state<AuthState>('checking');
	let token = $state('');
	let errorMessage = $state('');

	onMount(async () => {
		const storedToken = getStoredAdminToken();
		if (!storedToken) {
			authState = 'unauthenticated';
			return;
		}

		token = storedToken;

		try {
			const isValid = await validateAdminToken(storedToken);
			if (isValid) {
				authState = 'authenticated';
				return;
			}

			clearStoredAdminToken();
			token = '';
			authState = 'unauthenticated';
		} catch {
			errorMessage = '暂时无法验证已保存的管理令牌，请重新登录。';
			authState = 'unauthenticated';
		}
	});

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();

		const normalizedToken = token.trim();
		if (!normalizedToken) {
			errorMessage = '请输入管理令牌。';
			authState = 'unauthenticated';
			return;
		}

		authState = 'submitting';
		errorMessage = '';

		try {
			const isValid = await validateAdminToken(normalizedToken);
			if (!isValid) {
				clearStoredAdminToken();
				errorMessage = '管理令牌无效，请重新输入。';
				authState = 'unauthenticated';
				return;
			}

			setStoredAdminToken(normalizedToken);
			token = normalizedToken;
			authState = 'authenticated';
		} catch {
			errorMessage = '暂时无法连接后台服务，请稍后重试。';
			authState = 'unauthenticated';
		}
	}

	function handleLogout() {
		clearStoredAdminToken();
		token = '';
		errorMessage = '';
		authState = 'unauthenticated';
	}
</script>

<svelte:head>
	<title>后台管理</title>
</svelte:head>

{#if authState === 'authenticated'}
	<div class="flex min-h-screen items-center justify-center bg-slate-100 px-6 py-12">
		<section class="w-full max-w-xl rounded-lg border border-slate-200 bg-white p-8 shadow-sm">
			<div class="space-y-2">
				<p class="text-sm font-medium text-slate-500">后台管理</p>
				<h1 class="text-2xl font-semibold text-slate-900">占位内容</h1>
			</div>

			<div class="mt-8 flex justify-end">
				<button
					type="button"
					class="inline-flex items-center rounded-md border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-100"
					onclick={handleLogout}
				>
					退出登录
				</button>
			</div>
		</section>
	</div>
{:else if authState === 'checking'}
	<div class="flex min-h-screen items-center justify-center bg-slate-100 px-6 py-12">
		<p class="text-sm text-slate-500">正在检查登录状态...</p>
	</div>
{:else}
	<div class="flex min-h-screen items-center justify-center bg-slate-100 px-6 py-12">
		<section class="w-full max-w-md rounded-lg border border-slate-200 bg-white p-8 shadow-sm">
			<div class="space-y-2">
				<p class="text-sm font-medium text-slate-500">后台管理</p>
				<h1 class="text-2xl font-semibold text-slate-900">后台登录</h1>
			</div>

			<form class="mt-8 space-y-4" onsubmit={handleSubmit}>
				<div class="space-y-2">
					<label class="block text-sm font-medium text-slate-700" for="admin-token">管理令牌</label>
					<input
						id="admin-token"
						name="admin-token"
						type="password"
						bind:value={token}
						placeholder="请输入 ADMIN_BEARER_TOKEN"
						class="block w-full rounded-md border-slate-300 text-sm text-slate-900 placeholder:text-slate-400 focus:border-slate-500 focus:ring-slate-500"
						disabled={authState === 'submitting'}
						autocomplete="current-password"
					/>
				</div>

				{#if errorMessage}
					<p class="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600">
						{errorMessage}
					</p>
				{/if}

				<button
					type="submit"
					class="inline-flex w-full items-center justify-center rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
					disabled={authState === 'submitting'}
				>
					{#if authState === 'submitting'}
						登录中...
					{:else}
						登录
					{/if}
				</button>
			</form>
		</section>
	</div>
{/if}
