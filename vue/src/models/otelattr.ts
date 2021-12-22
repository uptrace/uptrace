export enum xkey {
  allSystem = 'all',
  internalSystem = 'internal',

  itemId = 'item.id',
  itemName = 'item.name',
  spanEvents = 'span.events',

  spanId = 'span.id',
  spanParentId = 'span.parent_id',
  spanTraceId = 'span.trace_id',

  spanSystem = 'span.system',
  spanGroupId = 'span.group_id',

  spanName = 'span.name',
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
  dbSqlTable = 'db.sql.table',

  exceptionStacktrace = 'exception.stacktrace',

  codeFunction = 'code.function',
  codeFilepath = 'code.filepath',
}

export function isDummySystem(system: string | undefined): boolean {
  const [type, sys] = splitTypeSystem(system)
  return type === xkey.allSystem || sys === xkey.allSystem
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
