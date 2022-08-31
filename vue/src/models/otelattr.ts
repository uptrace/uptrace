export enum xkey {
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

  serviceName = 'service.name',
  hostName = 'host.name',
  enduserId = 'enduser.id',

  httpMethod = 'http.method',
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

export enum xsys {
  all = 'all',
  internal = 'internal',

  error = 'error',
  exception = 'exception',
  logWarn = 'log:warn',
  logError = 'log:error',
  logFatal = 'log:fatal',
  logPanic = 'log:panic',

  event = 'event',

  logPrefix = 'log:',
  messagePrefix = 'message:',
}

export function isDummySystem(system: string | undefined): boolean {
  const [type, sys] = splitTypeSystem(system)
  return type === xkey.allSystem || sys === xkey.allSystem
}

export function isEventSystem(system: string | undefined): boolean {
  if (!system) {
    return false
  }
  return (
    isErrorSystem(system) ||
    system === xsys.event ||
    system.startsWith(xsys.logPrefix) ||
    system.startsWith(xsys.messagePrefix)
  )
}

export function isErrorSystem(system: string | undefined): boolean {
  if (!system) {
    return false
  }
  switch (system) {
    case xsys.error:
    case xsys.exception:
    case xsys.logError:
    case xsys.logFatal:
    case xsys.logPanic:
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
    if (s.slice(i + 1) === xkey.allSystem) {
      return [s.slice(0, i), xkey.allSystem]
    }
    return [s.slice(0, i), s]
  }
  return [s, s]
}
