package error_handling

type Code string

const (
	CodeNotFound         Code = "NOT_FOUND"
	CodeBadRequest       Code = "BAD_REQUEST"
	CodeConflict         Code = "CONFLICT"
	CodeInternal         Code = "INTERNAL_ERROR"
	CodeValidation       Code = "VALIDATION_ERROR"
	CodeMethodNotAllowed Code = "METHOD_NOT_ALLOWED"
)
