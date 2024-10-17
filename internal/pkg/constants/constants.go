package constants

import (
	"time"
)

const (
	RootConfig = "/home/handysuherman/Workspaces/Project/2024/wd-payment-backend"
	DDMMYY     = "02-01-2006"
	DMY        = "02/01/2006"
	TZ         = "2006-01-02T15:04:05Z"
	// Configurations Keys
	ETCD             = "etcd"
	Services         = "services"
	Storages         = "storages"
	Brokers          = "brokers"
	ServiceDiscovery = "serviceDiscovery"
	Encryption       = "encryption"
	Databases        = "databases"
	Monitoring       = "monitoring"
	OAUTH            = "oauth"
	TLS              = "tls"

	// Configurations
	AppTLS     = "app"
	ConsulTLS  = "consul"
	KafkaTLS   = "kafka"
	MariadbTLS = "mariadb"
	PasetoTLS  = "paseto"
	RedisTLS   = "redis"

	Validate        = "validate"
	FieldValidation = "field validation"
	RequiredHeaders = "required header"
	Base64          = "base64"
	Unmarshal       = "unmarshal"
	Uuid            = "uuid"
	Cookie          = "cookie"
	Token           = "token"
	Bcrypt          = "bcrypt"
	SQLState        = "sqlstate"
	EnvPath         = "./apigw_service/config"
	EnvType         = "env"
	Yaml            = "yaml"
	EnvName         = "app"

	// uuidErr
	InvalidUUIDLength = "uuid length"
	InvalidUUIDFormat = "uuid format"

	// http-headers
	Accept              = "Accept"
	ContentType         = "Content-Type"
	ApplicationJSON     = "application/json; charset=utf8"
	XForwardedFor       = "X-Forwarded-For"
	UserAgent           = "user-agent"
	XPlatform           = "x-pt-e3f264be"
	WithCredentials     = "withCredentials"
	WithCredentialsTrue = "true"
	Authorization       = "Authorization"
	Bearer              = "Bearer"

	// string concatenation
	TrailingSlash = "/"
	Ampersand     = "&"
	Equal         = "="
	QuestionMark  = "?"
	Delete        = "delete"
	Publish       = "publish"

	// http-form
	FileForm = "file"

	// http-cookies
	AccessToken    = "_s_04250fa5_session_id"
	RefreshToken   = "_s_65r60230_session_id"
	InstanceUUID   = "instance_uuid"
	AsesmenUUID    = "asesmen_uuid"
	SiapUUID       = "siap_uuid"
	SuratUUID      = "surat_uuid"
	PemasaranUUID  = "pemasaran_uuid"
	AkuntingUUID   = "akunting_uuid"
	LegalitasUUID  = "legalitas_uuid"
	SameSite       = "strict"
	Path           = "/"
	APICallTimeout = 15 * time.Second
	Production     = "production"

	// http-params / query-params
	UserID        = "userId"
	PelangganUid  = "pelangganUid"
	StokistUserID = "stokistUserId"
	SecretCode    = "secretCode"
	SerialCode    = "serialCode"
	StokisLevel   = "stokisLevel"
	SerialID      = "serialId"
	OTPCode       = "otpCode"
	ID            = "id"
	UID           = "uid"
	NIK           = "nik"
	Search        = "search"
	PageSize      = "page_size"
	PageID        = "page_id"

	// mime types
	ImageWEBP = "image/webp"
	ImageJPEG = "image/jpeg"
	ImagePNG  = "image/png"

	// Databases
	ElasticSearch = "ElasticSearch"
	MongoDB       = "MongoDB"
	Redis         = "redis"
	PostgreSQL    = "PostgreSQL"
	MariaDB       = "mariadb"

	// Logger
	GRPC       = "gRPC"
	Size       = "size"
	URI        = "URI"
	Status     = "status"
	StatusCode = "status_code"
	StatusText = "status_text"
	HTTP       = "HTTP"
	Error      = "ERROR"
	Protocol   = "protocol"
	Duration   = "duration"
	Method     = "method"
	MetaData   = "metadata"
	Request    = "request"
	Reply      = "reply"
	Time       = "time"
	Took       = "took"

	// MessageBroker
	Kafka            = "Kafka"
	KafkaTopic       = "Kafka-Topic"
	KafkaPartition   = "Kafka-Partition"
	KafkaMessage     = "Kafka-Message"
	KafkaWorkerID    = "Kafka-WorkerID"
	KafkaOffset      = "Kafka-Offset"
	KafkaGroupName   = "Kafka-GroupName"
	KafkaStreamID    = "Kafka-StreamID"
	KafkaEventID     = "Kafka-EventID"
	KafkaEventType   = "Kafka-EventType"
	KafkaEventNumber = "Kafka-EventNumber"
	CreatedDate      = "CreatedDate"
	UserMetadata     = "UserMetadata"

	// LSP
	KJK = "kjk"

	// PDF
	PDFAPL01   = "APL01.pdf"
	PDFAPL02   = "APL02.pdf"
	PDFMAPA01  = "MAPA01.pdf"
	PDFMAPA02  = "MAPA02.pdf"
	PDFFRAK01  = "FRAK01.pdf"
	PDFFRAK02  = "FRAK02.pdf"
	PDFFRAK03  = "FRAK03.pdf"
	PDFFRAK04  = "FRAK04.pdf"
	PDFFRAK05  = "FRAK05.pdf"
	PDFFRAK06  = "FRAK06.pdf"
	PDFMKVA    = "MKVA.pdf"
	PDFFRIA11  = "FRIA11.pdf"
	PDFFRIA01  = "FRIA01.pdf"
	PDFFRIA02  = "FRIA02.pdf"
	PDFFRIA03  = "FRIA03.pdf"
	PDFFRIA04  = "FRIA04.pdf"
	PDFFRIA05  = "FRIA05.pdf"
	PDFFRIA05A = "FRIA05A.pdf"
	PDFFRIA05B = "FRIA05B.pdf"
	PDFFRIA06  = "FRIA06.pdf"
	PDFFRIA06A = "FRIA06A.pdf"
	PDFFRIA06B = "FRIA06B.pdf"
	PDFFRIA07  = "FRIA07.pdf"
	PDFFRIA08  = "FRIA08.pdf"
	PDFFRIA09  = "FRIA09.pdf"
	PDFFRIA10  = "FRIA10.pdf"

	// MinioResponseType
	Minio                   = "Minio"
	MinioContentType        = "response-content-type"
	MinioContentDisposition = "response-content-disposition"
	MinioContentEncoding    = "response-content-encoding"
	MinioCacheControl       = "response-cache-control"
	MinioNoCache            = "No-Cache"
	MinioNoStore            = "No-Store"

	// QueryParams
	Query = "q"
	CT    = "ct" // stands for content-type

	YYYYMMDD = "2006-01-02"

	Email    = "email"
	Whatsapp = "whatsapp"

	// STOKIST
	StokisApplicatorID = "01HJW9A0Y39GCJ4QZTTM8W0H34"
	StokistSerialCode  = "998"

	// SYSTEM PYRAMID
	Stokist        = "stokist"
	Pelanggan      = "pelanggan"
	Aplikator      = "aplikator"
	System         = "system"
	StokistLevel   = int32(1)
	PelangganLevel = int32(2)
	AplikatorLevel = int32(3)
	SystemLevel    = int32(4)

	// AplikatorLevel
	Manager        = "manager"
	Inventory      = "inventory"
	Akunting       = "akunting"
	ManagerLevel   = int32(1)
	InventoryLevel = int32(2)
	AkuntingLevel  = int32(3)

	// TokenType
	AccessType  = "access"
	RefreshType = "refresh"

	// CleanArchitecture Layers
	Resolver       = "resolver"
	Handler        = "handler"
	Usecase        = "usecase"
	Repository     = "repository"
	ProducerWorker = "producerWorker"
	Worker         = "worker"
)
