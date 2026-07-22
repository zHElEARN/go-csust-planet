-- +goose Up
-- +goose StatementBegin
DO $migration$
DECLARE
    existing_table_count integer;
BEGIN
    SELECT count(*)
    INTO existing_table_count
    FROM pg_catalog.pg_class AS c
    JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relkind IN ('r', 'p')
      AND c.relname IN (
          'announcements',
          'app_versions',
          'campus_map_features',
          'semester_calendars'
      );

    IF existing_table_count NOT IN (0, 4) THEN
        RAISE EXCEPTION
            'cannot establish migration baseline: expected none or all four business tables, found %',
            existing_table_count;
    END IF;

    IF existing_table_count = 4 AND EXISTS (
        WITH expected_columns (
            table_name,
            ordinal_position,
            column_name,
            data_type,
            udt_name,
            is_nullable,
            column_default,
            character_maximum_length
        ) AS (
            VALUES
                ('announcements'::text, 1, 'id'::text, 'uuid'::text, 'uuid'::text, 'NO'::text, 'gen_random_uuid()'::text, NULL::integer),
                ('announcements', 2, 'title', 'character varying', 'varchar', 'NO', NULL, NULL),
                ('announcements', 3, 'content', 'text', 'text', 'NO', NULL, NULL),
                ('announcements', 4, 'is_active', 'boolean', 'bool', 'NO', 'true', NULL),
                ('announcements', 5, 'is_banner', 'boolean', 'bool', 'NO', 'false', NULL),
                ('announcements', 6, 'created_at', 'timestamp with time zone', 'timestamptz', 'NO', 'CURRENT_TIMESTAMP', NULL),
                ('app_versions', 1, 'id', 'uuid', 'uuid', 'NO', 'gen_random_uuid()', NULL),
                ('app_versions', 2, 'platform', 'character varying', 'varchar', 'NO', NULL, NULL),
                ('app_versions', 3, 'version_code', 'integer', 'int4', 'NO', NULL, NULL),
                ('app_versions', 4, 'version_name', 'character varying', 'varchar', 'NO', NULL, NULL),
                ('app_versions', 5, 'is_force_update', 'boolean', 'bool', 'NO', 'false', NULL),
                ('app_versions', 6, 'release_notes', 'text', 'text', 'NO', NULL, NULL),
                ('app_versions', 7, 'download_url', 'character varying', 'varchar', 'NO', NULL, NULL),
                ('app_versions', 8, 'created_at', 'timestamp with time zone', 'timestamptz', 'NO', 'CURRENT_TIMESTAMP', NULL),
                ('campus_map_features', 1, 'id', 'uuid', 'uuid', 'NO', 'gen_random_uuid()', NULL),
                ('campus_map_features', 2, 'type', 'character varying', 'varchar', 'NO', '''Feature''::character varying', 20),
                ('campus_map_features', 3, 'properties', 'jsonb', 'jsonb', 'NO', NULL, NULL),
                ('campus_map_features', 4, 'geometry', 'jsonb', 'jsonb', 'NO', NULL, NULL),
                ('semester_calendars', 1, 'id', 'uuid', 'uuid', 'NO', 'gen_random_uuid()', NULL),
                ('semester_calendars', 2, 'semester_code', 'character varying', 'varchar', 'NO', NULL, NULL),
                ('semester_calendars', 3, 'title', 'character varying', 'varchar', 'NO', NULL, NULL),
                ('semester_calendars', 4, 'subtitle', 'character varying', 'varchar', 'NO', NULL, NULL),
                ('semester_calendars', 5, 'calendar_start', 'date', 'date', 'NO', NULL, NULL),
                ('semester_calendars', 6, 'calendar_end', 'date', 'date', 'NO', NULL, NULL),
                ('semester_calendars', 7, 'semester_start', 'date', 'date', 'NO', NULL, NULL),
                ('semester_calendars', 8, 'semester_end', 'date', 'date', 'NO', NULL, NULL),
                ('semester_calendars', 9, 'notes', 'jsonb', 'jsonb', 'NO', '''[]''::jsonb', NULL),
                ('semester_calendars', 10, 'custom_week_ranges', 'jsonb', 'jsonb', 'NO', '''[]''::jsonb', NULL),
                ('semester_calendars', 11, 'created_at', 'timestamp with time zone', 'timestamptz', 'NO', 'CURRENT_TIMESTAMP', NULL)
        ),
        actual_columns AS (
            SELECT
                c.table_name::text,
                c.ordinal_position::integer,
                c.column_name::text,
                c.data_type::text,
                c.udt_name::text,
                c.is_nullable::text,
                c.column_default::text,
                c.character_maximum_length::integer
            FROM information_schema.columns AS c
            WHERE c.table_schema = 'public'
              AND c.table_name IN (
                  'announcements',
                  'app_versions',
                  'campus_map_features',
                  'semester_calendars'
              )
        )
        SELECT 1
        FROM (
            (SELECT * FROM expected_columns EXCEPT SELECT * FROM actual_columns)
            UNION ALL
            (SELECT * FROM actual_columns EXCEPT SELECT * FROM expected_columns)
        ) AS column_difference
    ) THEN
        RAISE EXCEPTION 'cannot establish migration baseline: business column definitions do not match V1';
    END IF;

    IF existing_table_count = 4 AND (
        SELECT count(*)
        FROM pg_catalog.pg_constraint AS con
        JOIN pg_catalog.pg_class AS c ON c.oid = con.conrelid
        JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public'
          AND c.relname IN (
              'announcements',
              'app_versions',
              'campus_map_features',
              'semester_calendars'
          )
          AND con.contype = 'p'
          AND pg_catalog.pg_get_constraintdef(con.oid, true) = 'PRIMARY KEY (id)'
    ) <> 4 THEN
        RAISE EXCEPTION 'cannot establish migration baseline: business primary keys do not match V1';
    END IF;

    IF existing_table_count = 4 AND EXISTS (
        WITH expected_indexes (indexname, indexdef) AS (
            VALUES
                ('idx_active_created'::text, 'CREATE INDEX idx_active_created ON public.announcements USING btree (is_active, created_at)'::text),
                ('idx_platform_version', 'CREATE INDEX idx_platform_version ON public.app_versions USING btree (platform, version_code)'),
                ('idx_semester_calendars_semester_code', 'CREATE UNIQUE INDEX idx_semester_calendars_semester_code ON public.semester_calendars USING btree (semester_code)')
        ),
        actual_indexes AS (
            SELECT i.indexname::text, i.indexdef::text
            FROM pg_catalog.pg_indexes AS i
            WHERE i.schemaname = 'public'
              AND i.indexname IN (
                  'idx_active_created',
                  'idx_platform_version',
                  'idx_semester_calendars_semester_code'
              )
        )
        SELECT 1
        FROM (
            (SELECT * FROM expected_indexes EXCEPT SELECT * FROM actual_indexes)
            UNION ALL
            (SELECT * FROM actual_indexes EXCEPT SELECT * FROM expected_indexes)
        ) AS index_difference
    ) THEN
        RAISE EXCEPTION 'cannot establish migration baseline: business indexes do not match V1';
    END IF;
END
$migration$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS public.announcements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    title varchar NOT NULL,
    content text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    is_banner boolean DEFAULT false NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT announcements_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.app_versions (
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

CREATE TABLE IF NOT EXISTS public.campus_map_features (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    type varchar(20) DEFAULT 'Feature'::varchar NOT NULL,
    properties jsonb NOT NULL,
    geometry jsonb NOT NULL,
    CONSTRAINT campus_map_features_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.semester_calendars (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    semester_code varchar NOT NULL,
    title varchar NOT NULL,
    subtitle varchar NOT NULL,
    calendar_start date NOT NULL,
    calendar_end date NOT NULL,
    semester_start date NOT NULL,
    semester_end date NOT NULL,
    notes jsonb DEFAULT '[]'::jsonb NOT NULL,
    custom_week_ranges jsonb DEFAULT '[]'::jsonb NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT semester_calendars_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_active_created
    ON public.announcements USING btree (is_active, created_at);

CREATE INDEX IF NOT EXISTS idx_platform_version
    ON public.app_versions USING btree (platform, version_code);

CREATE UNIQUE INDEX IF NOT EXISTS idx_semester_calendars_semester_code
    ON public.semester_calendars USING btree (semester_code);

COMMENT ON COLUMN public.announcements.title IS '公告标题';
COMMENT ON COLUMN public.announcements.content IS '公告正文内容';
COMMENT ON COLUMN public.announcements.is_active IS '是否生效(控制公告上下线)';
COMMENT ON COLUMN public.announcements.is_banner IS '是否在App头部Banner处显示';
COMMENT ON COLUMN public.app_versions.platform IS '平台(ios或android)';
COMMENT ON COLUMN public.app_versions.version_code IS '内部版本号(用于逻辑比对)';
COMMENT ON COLUMN public.app_versions.version_name IS '展示版本号(例如1.5.1)';
COMMENT ON COLUMN public.app_versions.is_force_update IS '是否强制更新';
COMMENT ON COLUMN public.app_versions.release_notes IS '更新日志';
COMMENT ON COLUMN public.app_versions.download_url IS '下载地址';
COMMENT ON COLUMN public.campus_map_features.type IS 'GeoJSON要素类型';
COMMENT ON COLUMN public.campus_map_features.properties IS '业务属性(如名称、分类、校区)';
COMMENT ON COLUMN public.campus_map_features.geometry IS '几何数据(Polygon及坐标点)';
COMMENT ON COLUMN public.semester_calendars.semester_code IS '学期代码(如: 2024-2025-1)';
COMMENT ON COLUMN public.semester_calendars.title IS '校历标题(如: 2024-2025学年度校历)';
COMMENT ON COLUMN public.semester_calendars.subtitle IS '校历副标题(如: 第一学期)';
COMMENT ON COLUMN public.semester_calendars.calendar_start IS '校历开始日期';
COMMENT ON COLUMN public.semester_calendars.calendar_end IS '校历结束日期';
COMMENT ON COLUMN public.semester_calendars.semester_start IS '学期开学日期';
COMMENT ON COLUMN public.semester_calendars.semester_end IS '学期结束日期';
COMMENT ON COLUMN public.semester_calendars.notes IS '校历底部备注(JSON数组)';
COMMENT ON COLUMN public.semester_calendars.custom_week_ranges IS '自定义周次与假期范围(JSON数组)';
