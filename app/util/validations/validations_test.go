package validations

import (
	"reflect"
	"testing"
)

func TestMustBe(t *testing.T) {
	tests := []struct {
		name       string
		validation func(v int) bool
		value      int
		want       MustBeOr[int]
	}{
		{
			name: "PositiveResult",
			validation: func(int) bool {
				return true
			},
			value: 5,
			want:  &mustBeOrImpl[int]{value: 5, result: true},
		},
		{
			name: "NegativeResult",
			validation: func(v int) bool {
				return false
			},
			value: 6,
			want:  &mustBeOrImpl[int]{value: 6, result: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustBe(tt.validation)(tt.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustBe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotEmptyString(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{
			name: "NotEmptyString",
			arg:  "Test",
			want: true,
		},
		{
			name: "EmptyString",
			arg:  "",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NotEmptyString(tt.arg); got != tt.want {
				t.Errorf("NotEmptyString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotEmptySlice(t *testing.T) {
	tests := []struct {
		name string
		arg  []any
		want bool
	}{
		{
			name: "NotEmptySlice",
			arg:  []any{1, 2, 3},
			want: true,
		},
		{
			name: "EmptySlice",
			arg:  []any{},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NotEmptySlice(tt.arg); got != tt.want {
				t.Errorf("NotEmptySlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotEmptyMap(t *testing.T) {
	tests := []struct {
		name string
		arg  map[int]string
		want bool
	}{
		{
			name: "NotEmptyMap",
			arg:  map[int]string{1: "one", 2: "two"},
			want: true,
		},
		{
			name: "EmptyMap",
			arg:  map[int]string{},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NotEmptyMap(tt.arg); got != tt.want {
				t.Errorf("NotEmptyMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mustBeOrImpl_OrElse(t *testing.T) {
	type fields struct {
		value  int
		result bool
	}
	type args struct {
		other int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name:   "SuccessfulValidation",
			fields: fields{value: 1, result: true},
			args:   args{other: 2},
			want:   1,
		},
		{
			name:   "FailedValidation",
			fields: fields{value: 1, result: false},
			args:   args{other: 2},
			want:   2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mustBeOrImpl[int]{
				value:  tt.fields.value,
				result: tt.fields.result,
			}
			if got := m.OrElse(tt.args.other); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OrElse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mustBeOrImpl_OrElseGet(t *testing.T) {
	type fields struct {
		value  int
		result bool
	}
	type args struct {
		getter func() int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name:   "SuccessfulValidation",
			fields: fields{value: 1, result: true},
			args: args{func() int {
				return 100
			}},
			want: 1,
		},
		{
			name:   "FailedValidation",
			fields: fields{value: 1, result: false},
			args: args{func() int {
				return 100
			}},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mustBeOrImpl[int]{
				value:  tt.fields.value,
				result: tt.fields.result,
			}
			if got := m.OrElseGet(tt.args.getter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OrElseGet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mustBeOrImpl_OrPanic(t *testing.T) {
	type fields struct {
		value  int
		result bool
	}
	tests := []struct {
		name        string
		fields      fields
		shouldPanic bool
	}{
		{
			name:        "WithoutPanic",
			fields:      fields{value: 1, result: true},
			shouldPanic: false,
		},
		{
			name:        "ShouldPanic",
			fields:      fields{value: 2, result: false},
			shouldPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if (recover() == nil) == tt.shouldPanic {
					t.Errorf("Unexpected panic or lack of it")
				}
			}()
			m := &mustBeOrImpl[int]{
				value:  tt.fields.value,
				result: tt.fields.result,
			}
			m.OrPanic()
		})
	}
}

func Test_mustBeOrImpl_OrPanicWithMessage(t *testing.T) {
	type fields struct {
		value  int
		result bool
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		shouldPanic bool
	}{
		{
			name:        "WithoutPanic",
			fields:      fields{value: 1, result: true},
			args:        args{msg: "AAAAA"},
			shouldPanic: false,
		},
		{
			name:        "ShouldPanic",
			fields:      fields{value: 2, result: false},
			args:        args{msg: "AAAAA"},
			shouldPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				err := recover()
				if (err == nil) == tt.shouldPanic {
					t.Errorf("Unexpected panic or lack of it")
				}
				if err != nil {
					println(reflect.TypeOf(err).Name())
				}
				if err != nil && err.(error).Error() != tt.args.msg {
					t.Errorf("Unexpected error message")
				}
			}()
			m := &mustBeOrImpl[int]{
				value:  tt.fields.value,
				result: tt.fields.result,
			}
			m.OrPanicWithMessage(tt.args.msg)
		})
	}
}
