export enum AttrKey {
  spanId = '.id',
  spanParentId = '.parent_id',
  spanTraceId = '.trace_id',

  spanSystem = '.system',
  spanGroupId = '.group_id',

  spanName = '.name',
  spanEventName = '.event_name',
  displayName = 'display.name',

  spanIsEvent = '.is_event',
  spanKind = '.kind',
  spanTime = '.time',
  spanDuration = '.duration',

  spanStatusCode = '.status_code',
  spanStatusMessage = '.status_message',

  spanCount = '.count',
  spanCountPerMin = 'per_min(.count)',
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
  All = 'all',
  EventsAll = 'events:all',
  SpansAll = 'spans:all',
  LogAll = 'log:all',
  HttpAll = 'http:all',
  DbAll = 'db:all',

  LogWarn = 'log:warn',
  LogError = 'log:error',
  LogFatal = 'log:fatal',
  LogPanic = 'log:panic',
  Funcs = 'funcs',
  OtherEvents = 'other-events',

  LogPrefix = 'log:',
  MessagePrefix = 'message:',
}

export enum EventName {
  Log = 'log',
}

export function isSpanSystem(...systems: string[]): boolean {
  if (!systems.length) {
    return false
  }
  return systems.every((system) => {
    if (system === SystemName.All) {
      return false
    }
    return !isEventOrLogSystem(system)
  })
}

export function isEventOrLogSystem(...systems: any[]): boolean {
  return isEventSystem(...systems) || isLogSystem(...systems)
}

export function isLogSystem(...systems: any[]): boolean {
  if (!systems.length) {
    return false
  }
  return systems.every((system) => {
    if (typeof system !== 'string') {
      return true
    }
    return system.startsWith('log:')
  })
}

export function isEventSystem(...systems: any[]): boolean {
  if (!systems.length) {
    return false
  }
  return systems.every((system) => {
    if (typeof system !== 'string') {
      return true
    }
    return (
      system === SystemName.EventsAll ||
      system === SystemName.OtherEvents ||
      system.startsWith(SystemName.MessagePrefix)
    )
  })
}

export function isErrorSystem(...systems: string[]): boolean {
  if (!systems.length) {
    return false
  }
  return systems.every((system) => {
    switch (system) {
      case SystemName.LogError:
      case SystemName.LogFatal:
      case SystemName.LogPanic:
        return true
    }
    return false
  })
}

export function systemMatches(system: string, pattern: string): boolean {
  switch (pattern) {
    case SystemName.All:
      return true
    case SystemName.SpansAll:
      return isSpanSystem(system)
    case SystemName.LogAll:
      return isLogSystem(system)
    case SystemName.EventsAll:
      return isEventSystem(system)
    default: {
      const [systemType, systemName] = splitTypeSystem(pattern)
      if (systemName === SystemName.All) {
        return system.startsWith(systemType + ':')
      }
      return system === pattern
    }
  }
}

export function isGroupSystem(system: string | undefined): boolean {
  const [type, sys] = splitTypeSystem(system)
  return type === SystemName.All || sys === SystemName.All
}

export function splitTypeSystem(s: string | undefined): [string, string] {
  if (!s) {
    return ['', '']
  }

  const i = s.indexOf(':')
  if (i >= 0) {
    if (s.slice(i + 1) === SystemName.All) {
      return [s.slice(0, i), SystemName.All]
    }
    return [s.slice(0, i), s]
  }
  return [s, s]
}
