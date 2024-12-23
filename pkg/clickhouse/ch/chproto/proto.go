package chproto

const (
	ClientHello        = 0
	ClientQuery        = 1
	ClientData         = 2
	ClientCancel       = 3
	ClientPing         = 4
	ClientTablesStatus = 5
	ClientKeepAlive    = 6
)
const (
	ServerHello               = 0
	ServerData                = 1
	ServerException           = 2
	ServerProgress            = 3
	ServerPong                = 4
	ServerEndOfStream         = 5
	ServerProfileInfo         = 6
	ServerTotals              = 7
	ServerExtremes            = 8
	ServerTablesStatus        = 9
	ServerLog                 = 10
	ServerTableColumns        = 11
	ServerPartUUIDs           = 12
	ServerReadTaskRequest     = 13
	ServerProfileEvents       = 14
	ServerTreeReadTaskRequest = 15
)
const (
	QueryNo        = 0
	QueryInitial   = 1
	QuerySecondary = 2
)
const (
	DBMS_MIN_REVISION_WITH_CLIENT_INFO                        = 54032
	DBMS_MIN_REVISION_WITH_SERVER_TIMEZONE                    = 54058
	DBMS_MIN_REVISION_WITH_QUOTA_KEY_IN_CLIENT_INFO           = 54060
	DBMS_MIN_REVISION_WITH_SERVER_DISPLAY_NAME                = 54372
	DBMS_MIN_REVISION_WITH_VERSION_PATCH                      = 54401
	DBMS_MIN_REVISION_WITH_CLIENT_WRITE_INFO                  = 54420
	DBMS_MIN_REVISION_WITH_SETTINGS_SERIALIZED_AS_STRINGS     = 54429
	DBMS_MIN_REVISION_WITH_INTERSERVER_SECRET                 = 54441
	DBMS_MIN_REVISION_WITH_OPENTELEMETRY                      = 54442
	DBMS_MIN_PROTOCOL_VERSION_WITH_DISTRIBUTED_DEPTH          = 54448
	DBMS_MIN_PROTOCOL_VERSION_WITH_INITIAL_QUERY_START_TIME   = 54449
	DBMS_MIN_PROTOCOL_VERSION_WITH_INCREMENTAL_PROFILE_EVENTS = 54451
	DBMS_MIN_REVISION_WITH_PARALLEL_REPLICAS                  = 54453
	DBMS_MIN_REVISION_WITH_CUSTOM_SERIALIZATION               = 54454
	DBMS_MIN_REVISION_WITH_ADDENDUM                           = 54458
	DBMS_MIN_REVISION_WITH_QUOTA_KEY                          = 54458
	DBMS_MIN_REVISION_WITH_PARAMETERS                         = 54459
	DBMS_MIN_REVISION_WITH_SERVER_QUERY_TIME_IN_PROGRESS      = 54460
	DBMS_TCP_PROTOCOL_VERSION                                 = DBMS_MIN_REVISION_WITH_PARALLEL_REPLICAS
)
