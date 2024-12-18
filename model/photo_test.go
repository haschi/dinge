package model

import (
	"bytes"
	"image"
	_ "image/jpeg"
	"io"
	"reflect"
	"testing"
)

func TestLoadImage(t *testing.T) {
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    image.Image
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadImage(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeImage(t *testing.T) {
	type args struct {
		image image.Image
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if err := EncodeImage(w, tt.args.image); (err != nil) != tt.wantErr {
				t.Errorf("EncodeImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("EncodeImage() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestResize(t *testing.T) {
	type args struct {
		src image.Image
	}
	tests := []struct {
		name string
		args args
		want image.Image
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Resize(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resize() = %v, want %v", got, tt.want)
			}
		})
	}
}

var origin = image.Pt(0, 0)

func TestCrop(t *testing.T) {
	tests := []struct {
		name string
		img  image.Image
		want image.Rectangle
	}{
		{
			name: "breiter als thumbnail",
			img:  image.Rect(0, 0, 600, 400),
			want: image.Rectangle{Min: image.Pt(0, 0), Max: image.Pt(400, 400)},
		},
		{
			name: "höher als thumbnail",
			img:  image.Rect(0, 0, 400, 600),
			want: image.Rectangle{Min: image.Pt(0, 0), Max: image.Pt(400, 400)},
		},
		{
			name: "gleiche größe",
			img:  image.Rect(0, 0, 600, 600),
			want: image.Rectangle{Min: image.Pt(0, 0), Max: image.Pt(600, 600)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zuschnitt := Crop(tt.img)
			if !tt.want.Eq(zuschnitt.Bounds()) {
				t.Errorf("want zuschnitt.Bounds().Eq(tt.want)), want %v.Eq(%v)", zuschnitt.Bounds(), tt.want)
			}
		})
	}
}
