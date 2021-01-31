package scanner

import (
	"bytes"
	"dxkite.cn/go-c11/token"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_scanner_next(t *testing.T) {
	s := &scanner{}
	code := "a中♥\r\n\ra\\\nb\na\\\r\nb"
	s.init("", bytes.NewBufferString(code))
	tests := []struct {
		Lit rune
		Pos token.Position
	}{
		{'a', token.Position{Filename: "", Line: 1, Column: 1, Offset: 0}},
		{'中', token.Position{Filename: "", Line: 1, Column: 2, Offset: 1}},
		{'♥', token.Position{Filename: "", Line: 1, Column: 3, Offset: 4}},
		{'\n', token.Position{Filename: "", Line: 1, Column: 4, Offset: 7}},
		{'\n', token.Position{Filename: "", Line: 2, Column: 1, Offset: 9}},
		{'a', token.Position{Filename: "", Line: 3, Column: 1, Offset: 10}},
		{'b', token.Position{Filename: "", Line: 4, Column: 1, Offset: 13}},
		{'\n', token.Position{Filename: "", Line: 4, Column: 2, Offset: 14}},
		{'a', token.Position{Filename: "", Line: 5, Column: 1, Offset: 15}},
		{'b', token.Position{Filename: "", Line: 6, Column: 1, Offset: 19}},
	}

	for _, tt := range tests {
		t.Run(string(tt.Lit), func(t *testing.T) {
			ch, pos := s.next()
			if ch != tt.Lit {
				t.Errorf("want %v got %v", tt.Lit, ch)
			}
			if !reflect.DeepEqual(tt.Pos, pos) {
				t.Errorf("want %+v got %+v", tt.Pos, pos)
			}
		})
	}
}

func exists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func TestScanFile(t *testing.T) {
	if err := filepath.Walk("testdata/", func(p string, info os.FileInfo, err error) error {
		ext := filepath.Ext(p)
		if ext == ".c" {
			t.Run(p, func(t *testing.T) {
				got, err := ScanFile(p)
				if err != nil {
					t.Errorf("ScanFile() error = %v", err)
					return
				}
				if !exists(p + ".json") {
					if err := token.SaveJson(p+".json", got); err != nil {
						t.Errorf("SaveJson error = %v", err)
						return
					}
				}
				want, err := token.LoadJson(p + ".json")
				if err != nil {
					t.Errorf("LoadJson() error = %v", err)
					return
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("ScanFile() got = %+v, want %+v", got, want)
				}
			})
		}
		return nil
	}); err != nil {
		t.Error(err)
	}
}

func Test_scanner_scanQuote(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			"success", `'1' '\123' '\x12' '\u1234' '\U12345678'`, false,
		},
		{
			"error", `'12'`, true,
		},
		{
			"error", `'\u12'`, true,
		},
		{
			"error", `'\1234'`, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ScanString("", tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_peekScanner_Peek(t *testing.T) {
	s := NewStringScan("", "int float u'1' '\\123' '\\x12' '\\u1234' '\\U12345678'")
	p := NewPeekScan(s)
	pn := p.Peek(3)
	for i , item := range pn {
		r := p.Scan()
		if !reflect.DeepEqual(item, r) {
			t.Errorf("PeekScanError(%d): want %v got %v", i, item, r)
		}
	}
}