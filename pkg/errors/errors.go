package errors

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// Error 定义错误接口
type Error interface {
	error
	Code() int
	Message() string
	InternalError() error
	WithCode(code int) Error
	WithMessage(message string) Error
	WithError(err error) Error
}

// ResponseError 响应错误
type ResponseError struct {
	StatusCode int    `json:"-"`       // HTTP状态码
	ErrorCode  int    `json:"code"`    // 业务错误码
	ErrorMsg   string `json:"message"` // 错误消息
	ERR        error  `json:"-"`       // 内部错误
}

// Error 错误信息
func (r *ResponseError) Error() string {
	if r.ErrorMsg != "" {
		return r.ErrorMsg
	}
	return http.StatusText(r.StatusCode)
}

// Code 错误码
func (r *ResponseError) Code() int {
	return r.ErrorCode
}

// Message 错误消息
func (r *ResponseError) Message() string {
	return r.ErrorMsg
}

// InternalError 内部错误
func (r *ResponseError) InternalError() error {
	return r.ERR
}

// WithCode 设置错误码
func (r *ResponseError) WithCode(code int) Error {
	r.ErrorCode = code
	return r
}

// WithMessage 设置错误消息
func (r *ResponseError) WithMessage(message string) Error {
	r.ErrorMsg = message
	return r
}

// WithError 设置内部错误
func (r *ResponseError) WithError(err error) Error {
	r.ERR = err
	return r
}

// NewError 创建指定错误消息的自定义错误
func NewError(message string) Error {
	return &ResponseError{
		StatusCode: 500,
		ErrorCode:  500,
		ErrorMsg:   message,
	}
}

// NewErrorWithCode 创建指定错误码和错误消息的自定义错误
func NewErrorWithCode(code int, message string) Error {
	return &ResponseError{
		StatusCode: code,
		ErrorCode:  code,
		ErrorMsg:   message,
	}
}

// FromError 从错误创建自定义错误
func FromError(err error) Error {
	if err == nil {
		return nil
	}

	var rerr *ResponseError
	if errors.As(err, &rerr) {
		return rerr
	}

	return &ResponseError{
		StatusCode: 500,
		ErrorCode:  500,
		ErrorMsg:   err.Error(),
		ERR:        err,
	}
}

// FromErrorWithMsg 从错误创建指定错误消息的自定义错误
func FromErrorWithMsg(err error, message string) Error {
	if err == nil {
		return nil
	}

	var rerr *ResponseError
	if errors.As(err, &rerr) {
		if message != "" {
			rerr.ErrorMsg = message
		}
		return rerr
	}

	return &ResponseError{
		StatusCode: 500,
		ErrorCode:  500,
		ErrorMsg:   message,
		ERR:        err,
	}
}

// FromErrorWithAll 从错误创建指定错误码和错误消息的自定义错误
func FromErrorWithAll(err error, code int, message string) Error {
	if err == nil {
		return nil
	}

	var rerr *ResponseError
	if errors.As(err, &rerr) {
		rerr.ErrorCode = code
		rerr.StatusCode = code
		if message != "" {
			rerr.ErrorMsg = message
		}
		return rerr
	}

	return &ResponseError{
		StatusCode: code,
		ErrorCode:  code,
		ErrorMsg:   message,
		ERR:        err,
	}
}

// StatusCode 获取HTTP状态码
func StatusCode(err error) int {
	if err == nil {
		return 200
	}

	var rerr *ResponseError
	if errors.As(err, &rerr) {
		return rerr.StatusCode
	}

	return 500
}

// Format 格式化错误
func Format(err error) string {
	if err == nil {
		return ""
	}

	var rerr *ResponseError
	if errors.As(err, &rerr) {
		if rerr.ERR != nil {
			return fmt.Sprintf("%s: %s", rerr.ErrorMsg, rerr.ERR.Error())
		}
		return rerr.ErrorMsg
	}

	return err.Error()
}

// BadRequest 错误请求
func BadRequest(message string, a ...interface{}) Error {
	if message == "" {
		message = "请求发生错误"
	}
	return &ResponseError{
		StatusCode: 400,
		ErrorCode:  400,
		ErrorMsg:   fmt.Sprintf(message, a...),
	}
}

// RequestEntityTooLarge 请求实体过大
func RequestEntityTooLarge(message string, a ...interface{}) Error {
	if message == "" {
		message = "请求实体过大"
	}
	return &ResponseError{
		StatusCode: 413,
		ErrorCode:  413,
		ErrorMsg:   fmt.Sprintf(message, a...),
	}
}

// Unauthorized 未授权
func Unauthorized(message string, a ...interface{}) Error {
	if message == "" {
		message = "未授权"
	}
	return &ResponseError{
		StatusCode: 401,
		ErrorCode:  401,
		ErrorMsg:   fmt.Sprintf(message, a...),
	}
}

// Forbidden 禁止访问
func Forbidden(message string, a ...interface{}) Error {
	if message == "" {
		message = "访问被禁止"
	}
	return &ResponseError{
		StatusCode: 403,
		ErrorCode:  403,
		ErrorMsg:   fmt.Sprintf(message, a...),
	}
}

// NotFound 资源不存在
func NotFound(message string, a ...interface{}) Error {
	if message == "" {
		message = "资源不存在"
	}
	return &ResponseError{
		StatusCode: 404,
		ErrorCode:  404,
		ErrorMsg:   fmt.Sprintf(message, a...),
	}
}

// MethodNotAllowed 方法不被允许
func MethodNotAllowed(message string, a ...interface{}) Error {
	if message == "" {
		message = "方法不被允许"
	}
	return &ResponseError{
		StatusCode: 405,
		ErrorCode:  405,
		ErrorMsg:   fmt.Sprintf(message, a...),
	}
}

// Conflict 资源冲突
func Conflict(message string, a ...interface{}) Error {
	if message == "" {
		message = "资源冲突"
	}
	return &ResponseError{
		StatusCode: 409,
		ErrorCode:  409,
		ErrorMsg:   fmt.Sprintf(message, a...),
	}
}

// TooManyRequests 请求过于频繁
func TooManyRequests(message string, a ...interface{}) Error {
	if message == "" {
		message = "请求过于频繁，请稍后再试"
	}
	return &ResponseError{
		StatusCode: 429,
		ErrorCode:  429,
		ErrorMsg:   fmt.Sprintf(message, a...),
	}
}

// InternalServerError 服务器错误
func InternalServerError(message string, a ...interface{}) Error {
	if message == "" {
		message = "服务器发生错误，请稍后再试"
	}
	return &ResponseError{
		StatusCode: 500,
		ErrorCode:  500,
		ErrorMsg:   fmt.Sprintf(message, a...),
	}
}

func ServiceUnavailableError(message string, a ...interface{}) Error {
	if message == "" {
		message = "服务暂时不可用，请稍后再试"
	}
	return &ResponseError{
		StatusCode: 503,
		ErrorCode:  503,
		ErrorMsg:   fmt.Sprintf(message, a...),
	}
}

// IsNotFound 检查是否为资源不存在错误
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	var rerr *ResponseError
	if errors.As(err, &rerr) {
		return rerr.StatusCode == 404
	}

	return false
}

// IsBadRequest 检查是否为请求错误
func IsBadRequest(err error) bool {
	if err == nil {
		return false
	}

	var rerr *ResponseError
	if errors.As(err, &rerr) {
		return rerr.StatusCode == 400
	}

	return false
}

// IsRequestEntityTooLarge 检查是否为请求实体过大错误
func IsRequestEntityTooLarge(err error) bool {
	if err == nil {
		return false
	}

	var rerr *ResponseError
	if errors.As(err, &rerr) {
		return rerr.StatusCode == 413
	}

	return false
}

// IsUnauthorized 检查是否为未授权错误
func IsUnauthorized(err error) bool {
	if err == nil {
		return false
	}

	var rerr *ResponseError
	if errors.As(err, &rerr) {
		return rerr.StatusCode == 401
	}

	return false
}

// IsForbidden 检查是否为禁止访问错误
func IsForbidden(err error) bool {
	if err == nil {
		return false
	}

	var rerr *ResponseError
	if errors.As(err, &rerr) {
		return rerr.StatusCode == 403
	}

	return false
}

// IsConflict 检查是否为资源冲突错误
func IsConflict(err error) bool {
	if err == nil {
		return false
	}

	var rerr *ResponseError
	if errors.As(err, &rerr) {
		return rerr.StatusCode == 409
	}

	return false
}

// Is 检查错误是否相同
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 断言错误类型
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Wrap 包装错误
func Wrap(err error, message string) error {
	return errors.Wrap(err, message)
}

// Wrapf 格式化包装错误
func Wrapf(err error, format string, args ...interface{}) error {
	return errors.Wrapf(err, format, args...)
}

// WithStack 为错误添加堆栈信息
func WithStack(err error) error {
	return errors.WithStack(err)
}

// Errorf 格式化错误
func Errorf(format string, args ...interface{}) error {
	return errors.Errorf(format, args...)
}
