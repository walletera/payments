package werrors

import "fmt"

type ErrorCode int

const (
    InternalErrorCode ErrorCode = iota + 1
    ResourceNotFoundErrorCode
    TimeoutErrorCode
    ValidationErrorCode
    ResourceAlreadyExistErrorCode
    WrongResourceVersionErrorCode
    UnprocessableMessageErrorCode
)

type InternalError struct {
    wError
}

func newInternalError(msg string, retryable bool) InternalError {
    return InternalError{
        wError{
            retryable: retryable,
            code:      InternalErrorCode,
            message:   fmt.Sprintf("internal error: %s", msg),
        },
    }
}

func NewRetryableInternalError(msg string) InternalError {
    return newInternalError(msg, true)
}

func NewNonRetryableInternalError(msg string) InternalError {
    return newInternalError(msg, false)
}

type ResourceNotFoundError struct {
    wError
}

func NewResourceNotFoundError(msg string) ResourceNotFoundError {
    return ResourceNotFoundError{
        wError{
            retryable: false,
            code:      ResourceNotFoundErrorCode,
            message:   fmt.Sprintf("Resource not found error: %s", msg),
        },
    }
}

type TimeoutError struct {
    wError
}

func NewTimeoutError(msg string) TimeoutError {
    return TimeoutError{
        wError{
            retryable: true,
            code:      TimeoutErrorCode,
            message:   fmt.Sprintf("timeout error: %s", msg),
        },
    }
}

type ValidationError struct {
    wError
}

func NewValidationError(msg string) ValidationError {
    return ValidationError{
        wError{
            retryable: true,
            code:      ValidationErrorCode,
            message:   fmt.Sprintf("validation error: %s", msg),
        },
    }
}

type ResourceAlreadyExistError struct {
    wError
}

func NewResourceAlreadyExistError(msg string) ResourceAlreadyExistError {
    return ResourceAlreadyExistError{
        wError{
            retryable: false,
            code:      ResourceAlreadyExistErrorCode,
            message:   fmt.Sprintf("Resource already exist error: %s", msg),
        },
    }
}

type WrongResourceVersionError struct {
    wError
}

func NewWrongResourceVersionError(msg string) WrongResourceVersionError {
    return WrongResourceVersionError{
        wError{
            retryable: true,
            code:      WrongResourceVersionErrorCode,
            message:   fmt.Sprintf("wrong Resource version error: %s", msg),
        },
    }
}

type UnprocessableMessageError struct {
    wError
}

func NewUnprocessableMessageError(msg string) WrongResourceVersionError {
    return WrongResourceVersionError{
        wError{
            retryable: false,
            code:      UnprocessableMessageErrorCode,
            message:   fmt.Sprintf("unprocessable message: %s", msg),
        },
    }
}
