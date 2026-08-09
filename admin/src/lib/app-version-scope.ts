export type AppVersionScope = 'current' | 'legacy';
export type AppVersionListPath = '/app-versions' | '/app-versions?scope=legacy';
export type AppVersionNewPath = '/app-versions/new' | '/app-versions/new?scope=legacy';
export type AppVersionEditPath = `/app-versions/edit?${string}`;

export type AppVersionScopeDetails = {
	label: string;
	description: string;
};

const scopeDetails: Record<AppVersionScope, AppVersionScopeDetails> = {
	current: {
		label: '新版发布库',
		description: '服务调用 v2 接口的新客户端；这里的版本不会被旧客户端看到。'
	},
	legacy: {
		label: '旧版迁移库',
		description: '只服务仍调用 v1 接口的旧客户端；请仅在迁移准备阶段维护这里的版本。'
	}
};

export function parseAppVersionScope(value: string | null): AppVersionScope {
	return value === 'legacy' ? 'legacy' : 'current';
}

export function getAppVersionScopeDetails(scope: AppVersionScope): AppVersionScopeDetails {
	return scopeDetails[scope];
}

export function appVersionListPath(scope: AppVersionScope): AppVersionListPath {
	return scope === 'legacy' ? '/app-versions?scope=legacy' : '/app-versions';
}

export function appVersionNewPath(scope: AppVersionScope): AppVersionNewPath {
	return scope === 'legacy' ? '/app-versions/new?scope=legacy' : '/app-versions/new';
}

export function appVersionEditPath(scope: AppVersionScope, id: string): AppVersionEditPath {
	const params = new URLSearchParams({ id });
	if (scope === 'legacy') {
		params.set('scope', 'legacy');
	}

	return `/app-versions/edit?${params.toString()}`;
}
