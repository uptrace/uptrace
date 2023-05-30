export enum AttrKey {
  spanId = '.id',
  spanParentId = '.parent_id',
  spanTraceId = '.trace_id',

  spanSystem = '.system',
  spanGroupId = '.group_id',

  spanName = '.name',
  spanEventName = '.event_name',
  spanIsEvent = '.is_event',
  spanKind = '.kind',
  spanTime = '.time',
  spanDuration = '.duration',

  spanStatusCode = '.status_code',
  spanStatusMessage = '.status_message',

  spanCount = '.count',
  spanCountPerMin = '.count_per_min',
  spanErrorCount = '.error_count',
  spanErrorRate = '.error_rate',

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

export enum EventName {
  Log = 'log',
}

export function isDummySystem(system: string | undefined): boolean {
  if (!system) {
    return false
  }
  return system === SystemName.all || system.endsWith(':all')
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
    if (s.slice(i + 1) === SystemName.all) {
      return [s.slice(0, i), SystemName.all]
    }
    return [s.slice(0, i), s]
  }
  return [s, s]
}
