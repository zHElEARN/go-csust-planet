<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';

	const homePath = resolve('/');

	onMount(() => {
		void goto(homePath, { replaceState: true });
	});
</script>

<svelte:head>
	<title>后台管理</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-slate-100 px-6 py-12">
	<p class="text-sm text-slate-500">正在返回后台入口...</p>
</div>
