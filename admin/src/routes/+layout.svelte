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
		<div class="mx-auto flex min-h-screen max-w-7xl">
			<aside class="w-56 border-r border-slate-200 bg-white px-4 py-6">
				<nav class="space-y-1">
					{#each navItems as item (item.href)}
						<a
							href={item.href}
							class={`block rounded-md px-3 py-2 text-sm ${
								page.url.pathname === item.href
									? 'bg-slate-900 text-white'
									: 'text-slate-700 hover:bg-slate-100'
							}`}
						>
							{item.label}
						</a>
					{/each}
				</nav>

				<button
					type="button"
					class="mt-6 w-full rounded-md border border-slate-300 px-3 py-2 text-left text-sm text-slate-700 hover:bg-slate-100"
					onclick={handleLogout}
				>
					退出登录
				</button>
			</aside>

			<main class="min-w-0 flex-1 px-6 py-6">
				{@render children()}
			</main>
		</div>
	</div>
{/if}
