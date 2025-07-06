package errors

import (
	"errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		want string
	}{
		{
			name: "error with wrapped error",
			err: &AppError{
				Type:    ErrNotFound,
				Message: "task not found",
				Err:     errors.New("underlying error"),
			},
			want: "task not found: underlying error",
		},
		{
			name: "error without wrapped error",
			err: &AppError{
				Type:    ErrInvalidInput,
				Message: "invalid title",
			},
			want: "invalid title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &AppError{
		Type:    ErrInternal,
		Message: "internal error",
		Err:     underlying,
	}

	if got := err.Unwrap(); got != underlying {
		t.Errorf("AppError.Unwrap() = %v, want %v", got, underlying)
	}
}

func TestAppError_Is(t *testing.T) {
	err1 := &AppError{Type: ErrNotFound}
	err2 := &AppError{Type: ErrNotFound}
	err3 := &AppError{Type: ErrInvalidInput}

	if !err1.Is(err2) {
		t.Error("AppError.Is() should return true for same error type")
	}

	if err1.Is(err3) {
		t.Error("AppError.Is() should return false for different error type")
	}

	if err1.Is(errors.New("regular error")) {
		t.Error("AppError.Is() should return false for non-AppError")
	}
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		errFunc  func() error
		wantType ErrorType
		wantMsg  string
	}{
		{
			name: "NotFound",
			errFunc: func() error {
				return NotFound("task", "12345")
			},
			wantType: ErrNotFound,
			wantMsg:  "task not found: 12345",
		},
		{
			name: "InvalidInput",
			errFunc: func() error {
				return InvalidInput("status", "INVALID")
			},
			wantType: ErrInvalidInput,
			wantMsg:  "invalid status: INVALID",
		},
		{
			name: "InternalError",
			errFunc: func() error {
				return InternalError("database error", errors.New("connection failed"))
			},
			wantType: ErrInternal,
			wantMsg:  "database error: connection failed",
		},
		{
			name: "Duplicate",
			errFunc: func() error {
				return Duplicate("task", "12345")
			},
			wantType: ErrDuplicate,
			wantMsg:  "task already exists: 12345",
		},
		{
			name: "Permission",
			errFunc: func() error {
				return Permission("delete", "archived task")
			},
			wantType: ErrPermission,
			wantMsg:  "permission denied: cannot delete archived task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errFunc()
			appErr, ok := err.(*AppError)
			if !ok {
				t.Fatal("expected AppError type")
			}

			if appErr.Type != tt.wantType {
				t.Errorf("got error type %v, want %v", appErr.Type, tt.wantType)
			}

			if err.Error() != tt.wantMsg {
				t.Errorf("got error message %q, want %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestErrorTypeCheckers(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		checker func(error) bool
		want    bool
	}{
		{
			name:    "IsNotFound with NotFound error",
			err:     NotFound("task", "123"),
			checker: IsNotFound,
			want:    true,
		},
		{
			name:    "IsNotFound with other error",
			err:     InvalidInput("field", "value"),
			checker: IsNotFound,
			want:    false,
		},
		{
			name:    "IsInvalidInput with InvalidInput error",
			err:     InvalidInput("field", "value"),
			checker: IsInvalidInput,
			want:    true,
		},
		{
			name:    "IsInternal with Internal error",
			err:     InternalError("error", nil),
			checker: IsInternal,
			want:    true,
		},
		{
			name:    "IsDuplicate with Duplicate error",
			err:     Duplicate("task", "123"),
			checker: IsDuplicate,
			want:    true,
		},
		{
			name:    "IsPermission with Permission error",
			err:     Permission("action", "resource"),
			checker: IsPermission,
			want:    true,
		},
		{
			name:    "checker with regular error",
			err:     errors.New("regular error"),
			checker: IsNotFound,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.checker(tt.err); got != tt.want {
				t.Errorf("checker() = %v, want %v", got, tt.want)
			}
		})
	}
}