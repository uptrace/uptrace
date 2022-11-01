export enum AttrKey {
  allSystem = 'all',
  internalSystem = 'internal',

  itemId = 'item.id',
  itemName = 'item.name',
  itemQuery = 'item.query',
  spanEvents = 'span.events',

  spanId = 'span.id',
  spanParentId = 'span.parent_id',
  spanTraceId = 'span.trace_id',

  spanSystem = 'span.system',
  spanGroupId = 'span.group_id',

  spanName = 'span.name',
  spanEventName = 'span.event_name',
  spanIsEvent = 'span.is_event',
  spanKind = 'span.kind',
  spanTime = 'span.time',
  spanDuration = 'span.duration',

  spanStatusCode = 'span.status_code',
  spanStatusMessage = 'span.status_message',

  spanCount = 'span.count',
  spanCountPerMin = 'span.count_per_min',
  spanErrorCount = 'span.error_count',
  spanErrorPct = 'span.error_pct',

  deploymentEnvironment = 'deployment.environment',
  service = 'service',
  serviceName = 'service.name',
  serviceVersion = 'service.version',
  hostName = 'host.name',
  enduserId = 'enduser.id',

  httpMethod = 'http.method',
  httpRoute = 'http.route',
  httpTarget = 'http.target',
  httpStatusCode = 'http.status_code',

  rpcMethod = 'rpc.method',
  dbStatement = 'db.statement',
  dbStatementPretty = '_db.statement_pretty',
  dbOperation = 'db.operation',
  dbSqlTable = 'db.sql.table',
  dbSqlTables = 'db.sql.tables',

  exceptionType = 'exception.type',
  exceptionMessage = 'exception.message',
  exceptionStacktrace = 'exception.stacktrace',

  logSeverity = 'log.severity',
  logSource = 'log.source',
  logFilePath = 'log.file.path',
  logFileName = 'log.file.name',
  logMessage = 'log.message',

  codeFunction = 'code.function',
  codeFilepath = 'code.filepath',
}

export enum SystemName {
  all = 'all',
  eventsAll = 'events:all',
  spansAll = 'spans:all',

  httpAll = 'http:all',
  dbAll = 'db:all',
  logAll = 'log:all',

  funcs = 'funcs',
  exceptions = 'exceptions',
  logWarn = 'log:warn',
  logError = 'log:error',
  logFatal = 'log:fatal',
  logPanic = 'log:panic',
  otherEvents = 'other-events',

  logPrefix = 'log:',
  messagePrefix = 'message:',
}

export function isDummySystem(system: string | undefined): boolean {
  const [type, sys] = splitTypeSystem(system)
  return type === AttrKey.allSystem || sys === AttrKey.allSystem
}

export function isEventSystem(system: string | undefined): boolean {
  if (!system) {
    return false
  }
  return (
    system === SystemName.eventsAll ||
    isErrorSystem(system) ||
    system === SystemName.otherEvents ||
    system.startsWith(SystemName.logPrefix) ||
    system.startsWith(SystemName.messagePrefix)
  )
}

export function isErrorSystem(system: string | undefined): boolean {
  if (!system) {
    return false
  }
  switch (system) {
    case SystemName.exceptions:
    case SystemName.logError:
    case SystemName.logFatal:
    case SystemName.logPanic:
      return true
    default:
      return false
  }
}

export function splitTypeSystem(s: string | undefined): [string, string] {
  if (!s) {
    return ['', '']
  }

  const i = s.indexOf(':')
  if (i >= 0) {
    if (s.slice(i + 1) === AttrKey.allSystem) {
      return [s.slice(0, i), AttrKey.allSystem]
    }
    return [s.slice(0, i), s]
  }
  return [s, s]
}
