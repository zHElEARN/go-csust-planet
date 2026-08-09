<script lang="ts">
	import { getAppVersionScopeDetails, type AppVersionScope } from '$lib/app-version-scope';

	let { scope }: { scope: AppVersionScope } = $props();
	const details = $derived(getAppVersionScopeDetails(scope));
</script>

<div
	class={`rounded-md border px-4 py-3 text-sm ${
		scope === 'legacy'
			? 'border-amber-300 bg-amber-50 text-amber-900'
			: 'border-slate-200 bg-slate-50 text-slate-700'
	}`}
>
	<span class="font-medium">{details.label}：</span>{details.description}
</div>
