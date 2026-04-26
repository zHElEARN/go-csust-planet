<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { onMount } from 'svelte';

	import {
		clearStoredAdminToken,
		getStoredAdminToken,
		subscribeAdminAuthChange,
		validateAdminToken
	} from '$lib/admin-auth';
	import favicon from '$lib/assets/favicon.svg';
	import './layout.css';

	let { children } = $props();

	const homePath = resolve('/');
	const loginPath = resolve('/login');
	const announcementsPath = resolve('/announcements');
	const appVersionsPath = resolve('/app-versions');
	const semesterCalendarsPath = resolve('/semester-calendars');
	const navItems = [
		{ label: '后台管理', href: homePath },
		{ label: '公告管理', href: announcementsPath },
		{ label: '版本管理', href: appVersionsPath },
		{ label: '校历管理', href: semesterCalendarsPath }
	];

	let authState = $state<'checking' | 'authenticated' | 'unauthenticated'>('checking');
	let sidebarCollapsed = $state(false);

	const isLoginRoute = $derived(page.url.pathname === loginPath);

	async function syncAuthState() {
		const storedToken = getStoredAdminToken();
		if (!storedToken) {
			authState = 'unauthenticated';
			return;
		}

		try {
			const isValid = await validateAdminToken(storedToken);
			if (isValid) {
				authState = 'authenticated';
				return;
			}

			clearStoredAdminToken();
			authState = 'unauthenticated';
		} catch {
			clearStoredAdminToken();
			authState = 'unauthenticated';
		}
	}

	function handleLogout() {
		clearStoredAdminToken();
		void goto(loginPath, { replaceState: true });
	}

	onMount(() => {
		void syncAuthState();

		return subscribeAdminAuthChange(() => {
			void syncAuthState();
		});
	});

	$effect(() => {
		if (!browser || authState === 'checking') {
			return;
		}

		if (isLoginRoute && authState === 'authenticated') {
			void goto(homePath, { replaceState: true });
			return;
		}

		if (!isLoginRoute && authState === 'unauthenticated') {
			void goto(loginPath, { replaceState: true });
		}
	});
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

{#if authState === 'checking'}
	<div class="flex min-h-screen items-center justify-center bg-slate-100 px-6 py-12">
		<p class="text-sm text-slate-500">加载中</p>
	</div>
{:else if isLoginRoute}
	{@render children()}
{:else if authState === 'authenticated'}
	<div class="min-h-screen bg-slate-100 text-slate-900">
		<aside
			class={`fixed inset-y-0 left-0 flex border-r border-slate-200 bg-white py-6 ${
				sidebarCollapsed ? 'w-16 px-2' : 'w-56 px-4'
			}`}
		>
			<div class="flex min-h-0 flex-1 flex-col">
				<div class={sidebarCollapsed ? 'flex justify-center' : 'flex justify-end'}>
					<button
						type="button"
						class="rounded-md border border-slate-300 px-2 py-1 text-sm text-slate-700 hover:bg-slate-100"
						aria-label={sidebarCollapsed ? '展开侧栏' : '收起侧栏'}
						title={sidebarCollapsed ? '展开侧栏' : '收起侧栏'}
						onclick={() => {
							sidebarCollapsed = !sidebarCollapsed;
						}}
					>
						{sidebarCollapsed ? '>' : '<'}
					</button>
				</div>

				<nav class="mt-6 space-y-1">
					{#each navItems as item (item.href)}
						<a
							href={item.href}
							class={`block rounded-md py-2 text-sm ${
								page.url.pathname === item.href
									? 'bg-slate-900 text-white'
									: 'text-slate-700 hover:bg-slate-100'
							} ${sidebarCollapsed ? 'px-2 text-center' : 'px-3'}`}
							aria-label={item.label}
							title={item.label}
						>
							{#if sidebarCollapsed}
								{item.label.slice(0, 2)}
							{:else}
								{item.label}
							{/if}
						</a>
					{/each}
				</nav>

				<button
					type="button"
					class={`mt-6 rounded-md border border-slate-300 py-2 text-sm text-slate-700 hover:bg-slate-100 ${
						sidebarCollapsed ? 'px-2 text-center' : 'w-full px-3 text-left'
					}`}
					aria-label="退出登录"
					title="退出登录"
					onclick={handleLogout}
				>
					{sidebarCollapsed ? '退出' : '退出登录'}
				</button>
			</div>
		</aside>

		<main class={`min-h-screen min-w-0 px-6 py-6 ${sidebarCollapsed ? 'ml-16' : 'ml-56'}`}>
			{@render children()}
		</main>
	</div>
{/if}
