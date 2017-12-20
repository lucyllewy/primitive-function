package function

import (
	"image"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/fogleman/primitive/primitive"
	"github.com/nfnt/resize"
)

var (
	count      = 10
	mode       = 1
	alpha      = 128
	repeat     = 0
	inputSize  = 256
	outputSize = 1024
)

type shapeConfig struct {
	Count  int
	Mode   int
	Alpha  int
	Repeat int
}

type shapeConfigArray []shapeConfig

func getConfig() shapeConfig {
	if cnt, err := strconv.Atoi(os.Getenv("PRIMITIVE_COUNT")); err == nil {
		count = cnt
	}
	config := shapeConfig{
		Count:  count,
		Mode:   mode,
		Alpha:  alpha,
		Repeat: repeat,
	}
	return config
}

func getImage(url string) (image.Image, error) {
	imgStream, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	imgData, _, err := image.Decode(imgStream.Body)
	if err != nil {
		return nil, err
	}

	return imgData, nil
}

// Handle a serverless request
func Handle(req []byte) string {
	config := getConfig()

	rand.Seed(time.Now().UTC().UnixNano())
	var workers = runtime.NumCPU()

	// get image
	input, err := getImage(string(req[:]))
	if err != nil {
		return err.Error()
	}

	// scale down input image if needed
	size := uint(inputSize)
	if size > 0 {
		input = resize.Thumbnail(size, size, input, resize.Bilinear)
	}

	// determine background color
	bg := primitive.MakeColor(primitive.AverageImageColor(input))

	// run algorithm
	model := primitive.NewModel(input, bg, outputSize, workers)

	for i := 0; i < config.Count; i++ {
		// find optimal shape and add it to the model
		model.Step(primitive.ShapeType(config.Mode), config.Alpha, config.Repeat)
	}

	res := model.SVG()
	return res
}
