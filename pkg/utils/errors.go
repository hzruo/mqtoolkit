package utils

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorType 错误类型
type ErrorType string

const (
	ErrorTypeConnection   ErrorType = "CONNECTION"
	ErrorTypeValidation   ErrorType = "VALIDATION"
	ErrorTypeTimeout      ErrorType = "TIMEOUT"
	ErrorTypeAuth         ErrorType = "AUTHENTICATION"
	ErrorTypeNotFound     ErrorType = "NOT_FOUND"
	ErrorTypeInternal     ErrorType = "INTERNAL"
	ErrorTypeNetwork      ErrorType = "NETWORK"
	ErrorTypeConfig       ErrorType = "CONFIG"
	ErrorTypeSubscription ErrorType = "SUBSCRIPTION"
)

// AppError 应用错误
type AppError struct {
	Type    ErrorType `json:"type"`
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Cause   error     `json:"-"`
	Stack   string    `json:"stack,omitempty"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s:%s] %s - %s", e.Type, e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewError 创建新的应用错误
func NewError(errorType ErrorType, code, message string) *AppError {
	return &AppError{
		Type:    errorType,
		Code:    code,
		Message: message,
		Stack:   getStack(),
	}
}

// NewErrorWithCause 创建带原因的应用错误
func NewErrorWithCause(errorType ErrorType, code, message string, cause error) *AppError {
	return &AppError{
		Type:    errorType,
		Code:    code,
		Message: message,
		Cause:   cause,
		Stack:   getStack(),
	}
}

// NewErrorWithDetails 创建带详情的应用错误
func NewErrorWithDetails(errorType ErrorType, code, message, details string) *AppError {
	return &AppError{
		Type:    errorType,
		Code:    code,
		Message: message,
		Details: details,
		Stack:   getStack(),
	}
}

// getStack 获取调用栈
func getStack() string {
	var stack []string
	for i := 2; i < 10; i++ { // 跳过当前函数和调用者
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		// 只保留相对路径
		if idx := strings.LastIndex(file, "/"); idx >= 0 {
			file = file[idx+1:]
		}

		stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}
	return strings.Join(stack, "\n")
}

// 预定义的错误创建函数

// NewConnectionError 创建连接错误
func NewConnectionError(message string, cause error) *AppError {
	return NewErrorWithCause(ErrorTypeConnection, "CONN_001", message, cause)
}

// NewValidationError 创建验证错误
func NewValidationError(message, details string) *AppError {
	return NewErrorWithDetails(ErrorTypeValidation, "VAL_001", message, details)
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(message string) *AppError {
	return NewError(ErrorTypeTimeout, "TIMEOUT_001", message)
}

// NewAuthError 创建认证错误
func NewAuthError(message string) *AppError {
	return NewError(ErrorTypeAuth, "AUTH_001", message)
}

// NewNotFoundError 创建未找到错误
func NewNotFoundError(resource, id string) *AppError {
	return NewError(ErrorTypeNotFound, "NOT_FOUND_001",
		fmt.Sprintf("%s not found: %s", resource, id))
}

// NewInternalError 创建内部错误
func NewInternalError(message string, cause error) *AppError {
	return NewErrorWithCause(ErrorTypeInternal, "INT_001", message, cause)
}

// NewNetworkError 创建网络错误
func NewNetworkError(message string, cause error) *AppError {
	return NewErrorWithCause(ErrorTypeNetwork, "NET_001", message, cause)
}

// NewConfigError 创建配置错误
func NewConfigError(message, details string) *AppError {
	return NewErrorWithDetails(ErrorTypeConfig, "CFG_001", message, details)
}

// NewSubscriptionError 创建订阅错误
func NewSubscriptionError(message string, cause error) *AppError {
	return NewErrorWithCause(ErrorTypeSubscription, "SUB_001", message, cause)
}

// WrapError 包装错误
func WrapError(err error, message string) *AppError {
	if err == nil {
		return nil
	}

	// 如果已经是 AppError，直接返回
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	return NewErrorWithCause(ErrorTypeInternal, "WRAP_001", message, err)
}

// IsErrorType 检查错误类型
func IsErrorType(err error, errorType ErrorType) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == errorType
	}
	return false
}

// GetErrorCode 获取错误代码
func GetErrorCode(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return "UNKNOWN"
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
