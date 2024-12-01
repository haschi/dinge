package validation_test

import (
	"errors"
	"testing"

	"github.com/haschi/dinge/validation"
)

func TestStringOptions(t *testing.T) {
	type args struct {
		options []string
		value   string
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "one two three four",
			args: args{
				options: []string{"one", "two", "three"},
				value:   "one",
			},
			want: nil,
		},
		{
			name: "empty options",
			args: args{
				options: []string{},
				value:   "one",
			},
			want: validation.ErrValueNotIncluded,
		},
		{
			name: "not included",
			args: args{
				options: []string{"one", "two", "three"},
				value:   "four",
			},
			want: validation.ErrValueNotIncluded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := validation.StringOptions(tt.args.options...)
			err := validator(tt.args.value)

			if !errors.Is(err, tt.want) {
				t.Errorf("validation(%v) = %v; want %v", tt.args.value, err, tt.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	type args struct {
		lowerbound int
		value      int
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "less",
			args: args{
				lowerbound: 10,
				value:      9,
			},
			want: validation.ErrNumberTooSmall,
		},
		{
			name: "equal",
			args: args{
				lowerbound: 10,
				value:      10,
			},
			want: nil,
		},
		{
			name: "greater",
			args: args{
				lowerbound: 10,
				value:      11,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := validation.Min(tt.args.lowerbound)
			got := validator(tt.args.value)

			if !errors.Is(got, tt.want) {
				t.Errorf("validation(%v) = %v; want %v", tt.args.value, got, tt.want)
			}
		})
	}
}

func TestMaxLength(t *testing.T) {
	type args struct {
		max   uint
		value string
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "fewer characters than allowed",
			args: args{
				max:   10,
				value: "012345678",
			},
			want: nil,
		},
		{
			name: "as many characters as allowed",
			args: args{
				max:   10,
				value: "0123456789",
			},
			want: nil,
		},
		{
			name: "more characters than allowed",
			args: args{
				max:   10,
				value: "0123456789a",
			},
			want: validation.ErrTooManyCharacters,
		},
		{
			name: "no input allowed - empty",
			args: args{
				max:   0,
				value: "",
			},
			want: nil,
		},
		{
			name: "no input allowed - some characters",
			args: args{
				max:   0,
				value: "1",
			},
			want: validation.ErrTooManyCharacters,
		},
		{
			name: "no input allowed - spaces only",
			args: args{
				max:   0,
				value: " ",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := validation.MaxLength(tt.args.max)
			got := validator(tt.args.value)

			if !errors.Is(got, tt.want) {
				t.Errorf("validator(%v) = %v; want %v", tt.args.value, got, tt.want)
			}
		})
	}
}

func TestIsNotBlank(t *testing.T) {

	tests := []struct {
		name    string
		arg     string
		wantErr error
	}{
		{
			name:    "empty string",
			arg:     "",
			wantErr: validation.ErrEmptyString,
		},
		{
			name:    "string with spaces only",
			arg:     " ",
			wantErr: validation.ErrEmptyString,
		},
		{
			name:    "passing validation",
			arg:     "1",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validation.IsNotBlank(tt.arg)

			if !errors.Is(got, tt.wantErr) {
				t.Errorf("IsNotBlank() error = %v, wantErr %v", got, tt.wantErr)
			}
		})
	}
}
