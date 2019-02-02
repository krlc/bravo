package main

import "testing"

func Test_toReadable(t *testing.T) {
	type args struct {
		bytes int64
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{"Test bytes", args{1000}, "1000.00 bytes"},
		{"Test Kb", args{1024}, "1.00 KiB"},
		{"Test Mb", args{1024 << (10 * 1)}, "1.00 MiB"},
		{"Test Gb", args{1024 << (10 * 2)}, "1.00 GiB"},
		{"Test Gb", args{10233548576}, "9.53 GiB"}, // some random const
		{"Test Tb", args{1024 << (10 * 3)}, "1.00 TiB"},
		{"Test Tb", args{1024 << (10 * 4)}, "1024.00 TiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toReadable(tt.args.bytes); got != tt.want {
				t.Errorf("toReadable() = %v, expected %v", got, tt.want)
			}
		})
	}
}
