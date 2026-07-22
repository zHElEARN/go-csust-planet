-- +goose Up
DROP INDEX public.idx_platform_version;

CREATE UNIQUE INDEX idx_platform_version
    ON public.app_versions USING btree (platform, version_code);
