package main

import (
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
)

const desc = ``

var cli struct {
	Folders []string `short:"f" help:"paths to folder(s) of images" required:"true"`
}

// decode image
func decode(in io.Reader) (image.Image, error) {
	decoders := []func(io.Reader) (image.Image, error){
		png.Decode,
		decodeJPEG,
	}

	var lastErr error
	for _, decoder := range decoders {
		im, err := decoder(in)
		if err == nil {
			return im, nil
		}
		lastErr = err
	}

	return nil, lastErr
}

// fingerprint the image based on it's pixel RGB values.
// Nb. we ignore alpha
func fingerprint(in image.Image) string {
	vals := []interface{}{}

	for dx := in.Bounds().Min.X; dx < in.Bounds().Max.X; dx++ {
		for dy := in.Bounds().Min.Y; dy < in.Bounds().Max.Y; dy++ {
			r, g, b, _ := in.At(dx, dy).RGBA()
			vals = append(vals, r, g, b)
		}
	}

	return NewID(vals...)
}

// listdir finds files in a given dir
// Nb. we don't check sub folders
func listdir(dirname string) ([]string, error) {
	contained, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	files := []string{}
	for _, fdata := range contained {
		if fdata.IsDir() {
			continue
		}
		files = append(files, filepath.Join(dirname, fdata.Name()))
	}

	return files, nil
}

func main() {
	kong.Parse(
		&cli,
		kong.Name("dupe"),
		kong.Description(desc),
	)

	fingerprints := map[string]string{} // fingerprint => filename

	for _, folder := range cli.Folders {
		found, err := listdir(folder)
		if err != nil {
			panic(err)
		}

		for _, f := range found {
			log.Printf("checking %s\n", f)
			fd, err := os.Open(f)
			if err != nil {
				log.Printf("failed to open %s: %v\n", f, err)
				continue
			}

			im, err := decode(fd)
			if err != nil {
				log.Printf("failed to decode %s: %v\n", f, err)
				continue
			}

			finger := fingerprint(im)

			matched, ok := fingerprints[finger]
			if ok {
				log.Printf("duplicate: %s & %s\n", matched, f)
				continue
			}

			fingerprints[finger] = f
		}
	}
}
