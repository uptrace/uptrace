CREATE OR REPLACE TABLE tracing_data AS logs_data
ENGINE=Merge(currentDatabase(), '^(spans|events|logs)_data$')
