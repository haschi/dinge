package photo_test

import (
	"image"
	_ "image/jpeg"
	"reflect"
	"testing"

	"github.com/haschi/dinge/photo"
)

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
			if got := photo.Resize(tt.args.src); !reflect.DeepEqual(got, tt.want) {
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
			zuschnitt := photo.Crop(tt.img)
			if !tt.want.Eq(zuschnitt.Bounds()) {
				t.Errorf("want zuschnitt.Bounds().Eq(tt.want)), want %v.Eq(%v)", zuschnitt.Bounds(), tt.want)
			}
		})
	}
}
