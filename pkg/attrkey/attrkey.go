package attrkey

const (
	MetricInstrument = "_instrument"

	AlertType   = "alert_type"
	AlertStatus = "alert_status"

	SpanSystem  = "_system"
	SpanGroupID = "_group_id"

	SpanID       = "_id"
	SpanParentID = "_parent_id"
	SpanTraceID  = "_trace_id"

	SpanName      = "_name"
	SpanType      = "_type"
	SpanEventName = "_event_name"
	SpanIsEvent   = "_is_event"

	SpanKind     = "_kind"
	SpanTime     = "_time"
	SpanDuration = "_duration"

	SpanStatusCode    = "_status_code"
	SpanStatusMessage = "_status_message"

	SpanCount         = "_count"
	SpanCountSum      = "sum(_count)"
	SpanCountPerMin   = "per_min(sum(_count))"
	SpanErrorCount    = "_error_count"
	SpanErrorCountSum = "sum(_error_count)"
	SpanErrorRate     = "_error_rate"

	SpanLinkCount       = "_link_count"
	SpanEventCount      = "_event_count"
	SpanEventErrorCount = "_event_error_count"
	SpanEventLogCount   = "_event_log_count"

	DisplayName = "display_name"

	TelemetrySDKName     = "telemetry_sdk_name"
	TelemetrySDKVersion  = "telemetry_sdk_version"
	TelemetrySDKLanguage = "telemetry_sdk_language"
	TelemetryAutoVersion = "telemetry_auto_version"

	OtelLibraryName    = "otel_library_name"
	OtelLibraryVersion = "otel_library_version"

	DeploymentEnvironment = "deployment_environment"

	ServiceName       = "service_name"
	ServiceVersion    = "service_version"
	ServiceNamespace  = "service_namespace"
	ServiceInstanceID = "service_instance_id"
	PeerService       = "peer_service"

	HostID           = "host_id"
	HostName         = "host_name"
	HostType         = "host_type"
	HostArch         = "host_arch"
	HostImageName    = "host_image_name"
	HostImageID      = "host_image_id"
	HostImageVersion = "host_image_version"

	ServerAddress       = "server_address"
	ServerPort          = "server_port"
	ServerSocketDomain  = "server_socket_domain"
	ServerSocketAddress = "server_socket_address"
	ServerSocketPort    = "server_socket_port"

	ClientAddress       = "client_address"
	ClientPort          = "client_port"
	ClientSocketAddress = "client_socket_address"
	ClientSocketPort    = "client_socket_port"

	URLScheme   = "url_scheme"
	URLFull     = "url_full"
	URLPath     = "url_path"
	URLQuery    = "url_query"
	URLFragment = "url_fragment"

	HTTPRequestMethod       = "http_request_method"
	HTTPRequestBodySize     = "http_request_body_size"
	HTTPResponseBodySize    = "http_response_body_size"
	HTTPResponseStatusCode  = "http_response_status_code"
	HTTPResponseStatusClass = "http_response_status_class"
	HTTPRoute               = "http_route"

	RPCSystem  = "rpc_system"
	RPCService = "rpc_service"
	RPCMethod  = "rpc_method"

	UserAgentOriginal  = "user_agent_original"
	UserAgentName      = "user_agent_name"
	UserAgentVersion   = "user_agent_version"
	UserAgentOSName    = "user_agent_os_name"
	UserAgentOSVersion = "user_agent_os_version"
	UserAgentDevice    = "user_agent_device"
	UserAgentIsBot     = "user_agent_is_bot"

	DBSystem    = "db_system"
	DBName      = "db_name"
	DBStatement = "db_statement"
	DBOperation = "db_operation"
	DBSqlTable  = "db_sql_table"

	EnduserID    = "enduser_id"
	EnduserRole  = "enduser_role"
	EnduserScope = "enduser_scope"

	LogMessage        = "log_message"
	LogSeverity       = "log_severity"
	LogSeverityNumber = "log_severity_number"
	LogSource         = "log_source"
	LogFilePath       = "log_file_path"
	LogFileName       = "log_file_name"
	LogIOStream       = "log_iostream"

	ExceptionType       = "exception_type"
	ExceptionMessage    = "exception_message"
	ExceptionStacktrace = "exception_stacktrace"

	CodeFunction = "code_function"
	CodeFilepath = "code_filepath"

	MessagingSystem                            = "messaging_system"
	MessagingOperation                         = "messaging_operation"
	MessagingDestinationName                   = "messaging_destination_name"
	MessagingDestinationKind                   = "messaging_destination_kind"
	MessagingDestinationTemporary              = "messaging_destination_temporary"
	MessagingMessageID                         = "messaging_message_id"
	MessagingMessageType                       = "messaging_message_type" // TODO: remove
	MessagingMessagePayloadSizeBytes           = "messaging_message_payload_size_bytes"
	MessagingMessagePayloadCompressedSizeBytes = "messaging_message_payload_compressed_size_bytes"

	MessagingClientID           = "messaging_client_id"
	MessagingKafkaConsumerGroup = "messaging_kafka_consumer_group"

	CloudProvider         = "cloud_provider"
	CloudAccountID        = "cloud_account_id"
	CloudRegion           = "cloud_region"
	CloudResourceID       = "cloud_resource_id"
	CloudAvailabilityZone = "cloud_availability_zone"
	CloudPlatform         = "cloud_platform"

	K8SNamespaceName  = "k8s_namespace_name"
	K8SClusterName    = "k8s_cluster_name"
	K8SNodeName       = "k8s_node_name"
	K8SPodUID         = "k8s_pod_uid"
	K8SPodName        = "k8s_pod_name"
	K8SContainerName  = "k8s_container_name"
	K8SReplicasetName = "k8s_replicaset_name"
	K8SDeploymentName = "k8s_deployment_name"
	K8sDaemonsetName  = "k8s_daemonset_name"

	ProcessPID                = "process_pid"
	ProcessCommand            = "process_command"
	ProcessRuntimeName        = "process_runtime_name"
	ProcessRuntimeVersion     = "process_runtime_version"
	ProcessRuntimeDescription = "process_runtime_description"
)
