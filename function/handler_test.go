package function

import (
	"image"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/onsi/gomega/ghttp"
)

func Test_getConfig(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want shapeConfig
	}{
		{
			name: "Default Count",
			env:  "",
			want: shapeConfig{
				Count:  10,
				Mode:   1,
				Alpha:  128,
				Repeat: 0,
			},
		},
		{
			name: "Overridden Count",
			env:  "88",
			want: shapeConfig{
				Count:  88,
				Mode:   1,
				Alpha:  128,
				Repeat: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env == "" {
				os.Unsetenv("PRIMITIVE_COUNT")
			} else {
				os.Setenv("PRIMITIVE_COUNT", tt.env)
			}
			if got := getConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getImageFails(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    image.Image
		wantErr bool
	}{
		{
			name: "Malformed URL",
			args: args{
				url: "bad url data",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "URL unresponsive",
			args: args{
				url: "http://localhost:1234/nothing",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "URL 404",
			args: args{
				url: "http://localhost:8080/nothing",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getImage(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("getImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getImageSucceeds(t *testing.T) {
	const filename string = "test-assets/test.jpg"

	testImageFile, err := os.Open(filename)
	if err != nil {
		t.Errorf("%s: Could not open file: %v", filename, err)
		return
	}
	defer testImageFile.Close()

	testImageData, _, err := image.Decode(testImageFile)
	if err != nil {
		t.Errorf("%s: Decode JPG: %v", filename, err)
		return
	}

	testImageBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("%s: Could not read file: %v", filename, err)
		return
	}

	svr := ghttp.NewServer()
	svr.AppendHandlers(ghttp.RespondWith(200, testImageBytes))

	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    image.Image
		wantErr bool
	}{
		{
			name: "URL Success",
			args: args{
				url: svr.URL(),
			},
			want:    testImageData,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getImage(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("getImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getImage() = %v, want %v", got, tt.want)
			}
		})
	}
}
