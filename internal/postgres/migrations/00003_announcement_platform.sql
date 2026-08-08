-- +goose Up
ALTER TABLE public.announcements
    ADD COLUMN platform varchar;

UPDATE public.announcements
SET platform = 'ios';

ALTER TABLE public.announcements
    ALTER COLUMN platform SET NOT NULL;

ALTER TABLE public.announcements
    ADD CONSTRAINT announcements_platform_check
        CHECK (platform IN ('ios', 'android', 'all'));

DROP INDEX public.idx_active_created;

CREATE INDEX idx_active_platform_created
    ON public.announcements USING btree (is_active, platform, created_at DESC);

COMMENT ON COLUMN public.announcements.platform IS '发布平台(ios、android或all)';
