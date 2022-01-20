# Changelog

## v0.2.0

### 💡 Enhancements 💡

- Added support for having multiple isolated projects. Projects are defined in the YAML config.
- Added basic authorization. Users are defined in the YAML config.
- Added `SAMPLE BY` to `spans_index` table.
- Added query limits to most queries to better support large datasets.
- Improved error handling on invalid queries.
- Added services and hostnames overview.
- Use MessagePack to store spans in `spans_data` table.

To upgrade, reset CH schema with the following command:

```go
# Using binary
./uptrace --config=/etc/uptrace/uptrace.yml ch reset

# Using sources
go run --config=config/uptrace.yml cmd/uptrace/main.go ch reset
```
