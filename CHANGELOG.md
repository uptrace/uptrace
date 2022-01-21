# Changelog

## v0.2.0

### 💡 Enhancements 💡

- Added support for exceptions and in-app logs.
- Added services and hostnames overview.
- Added SQL query formatting when viewing spans.
- Require user authentication. Users are defined in the YAML config.
- Added support for having multiple isolated projects in the same database. Projects are defined in
  the YAML config.
- Added `SAMPLE BY` to `spans_index` table.
- Added query limits to `spans_index` queries to better support large datasets.
- Improved error handling on invalid Uptrace queries.
- Use faster and more compact MessagePack encoding to store spans in `spans_data` table.
- Add more attributes to ClickHouse index.

To upgrade, reset ClickHouse schema with the following command (existing data will be lost):

```go
# Using binary
./uptrace --config=/etc/uptrace/uptrace.yml ch reset

# Using sources
go run --config=config/uptrace.yml cmd/uptrace/main.go ch reset
```
