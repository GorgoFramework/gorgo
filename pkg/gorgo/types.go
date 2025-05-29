package gorgo

const (
	// 1xx Informational responses
	ContinueStatus           = 100
	SwitchingProtocolsStatus = 101
	ProcessingStatus         = 102
	EarlyHintsStatus         = 103

	// 2xx Success
	OKStatus                   = 200
	SuccessStatus              = 200 // Alias for backward compatibility
	CreatedStatus              = 201
	AcceptedStatus             = 202
	NonAuthoritativeInfoStatus = 203
	NoContentStatus            = 204
	ResetContentStatus         = 205
	PartialContentStatus       = 206
	MultiStatusStatus          = 207
	AlreadyReportedStatus      = 208
	IMUsedStatus               = 226

	// 3xx Redirection
	MultipleChoicesStatus   = 300
	MovedPermanentlyStatus  = 301
	FoundStatus             = 302
	SeeOtherStatus          = 303
	NotModifiedStatus       = 304
	UseProxyStatus          = 305
	TemporaryRedirectStatus = 307
	PermanentRedirectStatus = 308

	// 4xx Client errors
	BadRequestStatus                  = 400
	UnauthorizedStatus                = 401
	PaymentRequiredStatus             = 402
	ForbiddenStatus                   = 403
	NotFoundStatus                    = 404
	MethodNotAllowedStatus            = 405
	NotAcceptableStatus               = 406
	ProxyAuthRequiredStatus           = 407
	RequestTimeoutStatus              = 408
	ConflictStatus                    = 409
	GoneStatus                        = 410
	LengthRequiredStatus              = 411
	PreconditionFailedStatus          = 412
	PayloadTooLargeStatus             = 413
	URITooLongStatus                  = 414
	UnsupportedMediaTypeStatus        = 415
	RangeNotSatisfiableStatus         = 416
	ExpectationFailedStatus           = 417
	TeapotStatus                      = 418 // I'm a teapot (RFC 2324)
	MisdirectedRequestStatus          = 421
	UnprocessableEntityStatus         = 422
	LockedStatus                      = 423
	FailedDependencyStatus            = 424
	TooEarlyStatus                    = 425
	UpgradeRequiredStatus             = 426
	PreconditionRequiredStatus        = 428
	TooManyRequestsStatus             = 429
	RequestHeaderFieldsTooLargeStatus = 431
	UnavailableForLegalReasonsStatus  = 451

	// 5xx Server errors
	InternalServerErrorStatus           = 500
	NotImplementedStatus                = 501
	BadGatewayStatus                    = 502
	ServiceUnavailableStatus            = 503
	GatewayTimeoutStatus                = 504
	HTTPVersionNotSupportedStatus       = 505
	VariantAlsoNegotiatesStatus         = 506
	InsufficientStorageStatus           = 507
	LoopDetectedStatus                  = 508
	NotExtendedStatus                   = 510
	NetworkAuthenticationRequiredStatus = 511
)
