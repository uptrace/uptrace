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
	CompressionDisabled = 0
	CompressionEnabled  = 1
)

const (
	ServerHello        = 0
	ServerData         = 1
	ServerException    = 2
	ServerProgress     = 3
	ServerPong         = 4
	ServerEndOfStream  = 5
	ServerProfileInfo  = 6
	ServerTotals       = 7
	ServerExtremes     = 8
	ServerTablesStatus = 9
	ServerLog          = 10
	ServerTableColumns = 11
)

const (
	QueryNo        = 0
	QueryInitial   = 1
	QuerySecondary = 2
)
