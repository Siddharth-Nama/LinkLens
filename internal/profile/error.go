package profile

const (
	CodeInvalidURL      = "INVALID_URL"
	CodeNotFound        = "NOT_FOUND"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeSessionExpired  = "SESSION_EXPIRED"
	CodeRateLimited     = "RATE_LIMITED"
	CodeUpstreamError   = "UPSTREAM_ERROR"
	CodeUpstreamTimeout = "UPSTREAM_TIMEOUT"
	CodeInternal        = "INTERNAL_ERROR"
)

type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewError(code, message string) ErrorBody {
	return ErrorBody{Error: ErrorDetail{Code: code, Message: message}}
}
