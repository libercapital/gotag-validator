package gotag_validator

import "testing"

func TestIsCPF(t *testing.T) {
	type args struct {
		doc string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid_cpf_without_formatting",
			args: args{"02092051059"},
			want: true,
		},
		{
			name: "valid_cpf_with_formatting",
			args: args{"020.920.510-59"},
			want: true,
		},
		{
			name: "invalid_cpf_wrong_check_digits",
			args: args{"02092051057"},
			want: false,
		},
		{
			name: "invalid_cpf_all_same_digits",
			args: args{"222.222.222-22"},
			want: false,
		},
		{
			name: "invalid_cpf_too_short",
			args: args{"123456"},
			want: false,
		},
		{
			name: "invalid_cpf_letters",
			args: args{"ABCDEF12345"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCPF(tt.args.doc); got != tt.want {
				t.Errorf("isCPF() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCNPJ(t *testing.T) {
	type args struct {
		doc string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Numeric CNPJs
		{
			name: "valid_numeric_cnpj_without_formatting",
			args: args{"33796617000114"},
			want: true,
		},
		{
			name: "valid_numeric_cnpj_with_formatting",
			args: args{"33.796.617/0001-14"},
			want: true,
		},
		{
			name: "valid_numeric_cnpj_2",
			args: args{"13661345000138"},
			want: true,
		},
		{
			name: "valid_numeric_cnpj_3",
			args: args{"95307061000102"},
			want: true,
		},
		{
			name: "valid_numeric_cnpj_4",
			args: args{"38883574000128"},
			want: true,
		},
		{
			name: "invalid_numeric_cnpj_wrong_check_digit",
			args: args{"38883574000127"},
			want: false,
		},
		{
			name: "invalid_numeric_cnpj_wrong_check_digit_2",
			args: args{"33796617000112"},
			want: false,
		},
		{
			name: "invalid_numeric_cnpj_all_same_digits",
			args: args{"11111111111111"},
			want: false,
		},
		// Alphanumeric CNPJs (Receita Federal IN 2229/2024)
		{
			name: "valid_alphanumeric_cnpj_3MD06BLA000140",
			args: args{"3MD06BLA000140"},
			want: true,
		},
		{
			name: "valid_alphanumeric_cnpj_with_formatting",
			args: args{"3M.D06.BLA/0001-40"},
			want: true,
		},
		{
			name: "valid_alphanumeric_cnpj_with_formatting_2",
			args: args{"9H.V16.8L4/0001-89"},
			want: true,
		},
		{
			name: "invalid_alphanumeric_cnpj_wrong_check_digit",
			args: args{"1A2B345C6D7851"},
			want: false,
		},
		{
			name: "invalid_alphanumeric_cnpj_wrong_check_digit_2",
			args: args{"LG.EEE.7V1/0001-58"},
			want: false,
		},
		{
			name: "invalid_alphanumeric_cnpj_all_same_chars",
			args: args{"AAAAAAAAAAAA14"},
			want: false,
		},
		// Edge cases
		{
			name: "invalid_cnpj_too_short",
			args: args{"1234567890123"},
			want: false,
		},
		{
			name: "invalid_cnpj_too_long",
			args: args{"123456789012345"},
			want: false,
		},
		{
			name: "invalid_cnpj_empty",
			args: args{""},
			want: false,
		},
		{
			name: "valid_cnpj_lowercase_alphanumeric",
			args: args{"3md06bla000140"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCNPJ(tt.args.doc); got != tt.want {
				t.Errorf("isCNPJ() = %v, want %v", got, tt.want)
			}
		})
	}
}