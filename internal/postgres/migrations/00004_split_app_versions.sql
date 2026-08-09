-- +goose Up
ALTER TABLE public.app_versions
    RENAME TO legacy_app_versions;

ALTER TABLE public.legacy_app_versions
    RENAME CONSTRAINT app_versions_pkey TO legacy_app_versions_pkey;

ALTER INDEX public.idx_platform_version
    RENAME TO idx_legacy_platform_version;

CREATE TABLE public.app_versions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    platform varchar NOT NULL,
    version_code integer NOT NULL,
    version_name varchar NOT NULL,
    is_force_update boolean DEFAULT false NOT NULL,
    release_notes text NOT NULL,
    download_url varchar NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT app_versions_pkey PRIMARY KEY (id)
);

CREATE UNIQUE INDEX idx_platform_version
    ON public.app_versions USING btree (platform, version_code);

COMMENT ON TABLE public.legacy_app_versions IS '旧客户端使用的历史应用版本';
COMMENT ON TABLE public.app_versions IS '新客户端使用的应用版本';
COMMENT ON COLUMN public.app_versions.platform IS '平台(ios或android)';
COMMENT ON COLUMN public.app_versions.version_code IS '内部版本号(用于逻辑比对)';
COMMENT ON COLUMN public.app_versions.version_name IS '展示版本号(例如1.5.1)';
COMMENT ON COLUMN public.app_versions.is_force_update IS '是否强制更新';
COMMENT ON COLUMN public.app_versions.release_notes IS '更新日志';
COMMENT ON COLUMN public.app_versions.download_url IS '下载地址';
