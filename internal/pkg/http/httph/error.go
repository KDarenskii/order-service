package httph

import (
	"context"
	"net/http"
)

type contextKeyError struct{}

type contextValueError struct {
	err        error
	statusCode int
}

func errorPrepare(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyError{}, &contextValueError{})
}

func errorApply(ctx context.Context, err error) {
	contextError, ok := ctx.Value(contextKeyError{}).(*contextValueError)

	if ok && contextError != nil {
		contextError.err = err
	}
}

func errorGet(ctx context.Context) error {
	contextError, ok := ctx.Value(contextKeyError{}).(*contextValueError)

	if ok && contextError != nil {
		return contextError.err
	}

	return nil
}

func errorApplyStatusCode(ctx context.Context, statusCode int) {
	contextError, ok := ctx.Value(contextKeyError{}).(*contextValueError)

	if ok && contextError != nil {
		contextError.statusCode = statusCode
	}
}

func errorGetStatusCode(ctx context.Context) int {
	contextError, ok := ctx.Value(contextKeyError{}).(*contextValueError)

	if ok && contextError != nil {
		return contextError.statusCode
	}

	return 0
}

func ErrorPrepare(r *http.Request) *http.Request {
	return r.WithContext(errorPrepare(r.Context()))
}

func ErrorGet(r *http.Request) error {
	return errorGet(r.Context())
}

func ErrorApply(r *http.Request, err error) {
	errorApply(r.Context(), err)
}

func ErrorApplyStatusCode(r *http.Request, statusCode int) {
	errorApplyStatusCode(r.Context(), statusCode)
}

func ErrorGetStatusCode(r *http.Request) int {
	return errorGetStatusCode(r.Context())
}

type Middleware = func(http.Handler) http.Handler

func NewErrorMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, ErrorPrepare(r))
		})
	}
}
