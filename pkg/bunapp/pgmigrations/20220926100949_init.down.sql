DROP TABLE IF EXISTS users CASCADE;

--bun:split

DROP TABLE IF EXISTS projects CASCADE;

--bun:split

DROP TYPE IF EXISTS public.achiev_name_enum CASCADE;

--bun:split

DROP TABLE IF EXISTS achievements CASCADE;

--bun:split

DROP TABLE IF EXISTS user_project_data CASCADE;

--bun:split

DROP TABLE IF EXISTS saved_views CASCADE;

--bun:split

DROP TABLE IF EXISTS pinned_facets CASCADE;

--bun:split

DROP TYPE IF EXISTS public.monitor_state_enum CASCADE;

--bun:split

DROP TYPE IF EXISTS public.monitor_type_enum CASCADE;

--bun:split

DROP TABLE IF EXISTS monitors CASCADE;

--bun:split

DROP TYPE IF EXISTS public.notif_channel_type_enum CASCADE;

--bun:split

DROP TYPE IF EXISTS public.notif_channel_state_enum CASCADE;

--bun:split

DROP TABLE IF EXISTS notif_channels CASCADE;

--bun:split

DROP TABLE IF EXISTS monitor_channels CASCADE;

--bun:split

DROP TYPE IF EXISTS public.trackable_model_enum CASCADE;

--bun:split

DROP TYPE IF EXISTS public.alert_type_enum CASCADE;

--bun:split

DROP TABLE IF EXISTS alerts CASCADE;

--bun:split

DROP TYPE IF EXISTS public.alert_event_name_enum CASCADE;

--bun:split

DROP TYPE IF EXISTS public.alert_status_enum CASCADE;

--bun:split

DROP TABLE IF EXISTS alert_events CASCADE;

--bun:split

DROP TABLE IF EXISTS taskq_jobs CASCADE;

--bun:split

DROP TABLE IF EXISTS annotations CASCADE;

--bun:split

DROP TABLE IF EXISTS dashboards CASCADE;

--bun:split

DROP TYPE IF EXISTS public.grid_column_type_enum CASCADE;

--bun:split

DROP TABLE IF EXISTS dash_grid_columns CASCADE;

--bun:split

DROP TYPE IF EXISTS public.dash_kind_enum CASCADE;

--bun:split

DROP TABLE IF EXISTS dash_gauges CASCADE;

--bun:split

DROP TYPE IF EXISTS public.metric_instrument_enum CASCADE;

--bun:split

DROP TABLE IF EXISTS metrics CASCADE;
