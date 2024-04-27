export enum AttrKey {
  metricInstrument = '_instrument',

  spanId = '_id',
  spanParentId = '_parent_id',
  spanTraceId = '_trace_id',

  spanSystem = '_system',
  spanGroupId = '_group_id',

  spanName = '_name',
  spanEventName = '_event_name',

  spanIsEvent = '_is_event',
  spanKind = '_kind',
  spanTime = '_time',
  spanDuration = '_duration',

  spanStatusCode = '_status_code',
  spanStatusMessage = '_status_message',

  spanCount = '_count',
  spanCountSum = 'sum(_count)',
  spanCountPerMin = 'per_min(sum(_count))',
  spanErrorCount = '_error_count',
  spanErrorRate = '_error_rate',

  displayName = 'display_name',
  deploymentEnvironment = 'deployment_environment',
  service = 'service',
  serviceName = 'service_name',
  serviceVersion = 'service_version',
  hostName = 'host_name',
  enduserId = 'enduser_id',

  httpMethod = 'http_method',
  httpRoute = 'http_route',
  httpTarget = 'http_target',
  httpStatusCode = 'http_status_code',

  rpcMethod = 'rpc_method',

  dbSystem = 'db_system',
  dbOperation = 'db_operation',
  dbStatement = 'db_statement',
  dbSqlTable = 'db_sql_table',

  exceptionType = 'exception_type',
  exceptionMessage = 'exception_message',
  exceptionStacktrace = 'exception_stacktrace',

  logSeverity = 'log_severity',
  logSource = 'log_source',
  logFilePath = 'log_file_path',
  logFileName = 'log_file_name',
  logMessage = 'log_message',

  codeFunction = 'code_function',
  codeFilepath = 'code_filepath',

  otelLibraryName = 'otel_library_name',
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
  if (!system) {
    return false
  }
  return system === SystemName.All || system.endsWith(':all')
}

export function systemType(system: string | undefined): string {
  const [typ] = splitTypeSystem(system)
  return typ
}

export function splitTypeSystem(s: string | undefined): [string, string] {
  if (!s) {
    return ['', '']
  }

  const i = s.indexOf(':')
  if (i === -1) {
    return [s, s]
  }

  if (s.slice(i + 1) === SystemName.All) {
    return [s.slice(0, i), SystemName.All]
  }
  return [s.slice(0, i), s]
}
