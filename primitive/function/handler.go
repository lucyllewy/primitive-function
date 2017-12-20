package function

import (
	"image"
	"math/rand"
	"net/http"
	"net/url"
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

// Handle a serverless request
func Handle(req []byte) string {
	if cnt, err := strconv.Atoi(os.Getenv("PRIMITIVE_COUNT")); err == nil {
		count = cnt
	}
	config := shapeConfig{
		Count:  count,
		Mode:   mode,
		Alpha:  alpha,
		Repeat: repeat,
	}

	rand.Seed(time.Now().UTC().UnixNano())
	var workers = runtime.NumCPU()

	inputURL, err := url.Parse(string(req[:]))
	if err != nil {
		return err.Error()
	}

	imgStream, err := http.Get(inputURL.String())
	if err != nil {
		return err.Error()
	}

	// read input image
	input, _, err := image.Decode(imgStream.Body)
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
