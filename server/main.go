package main

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"log"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gocv.io/x/gocv"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(cors.Default())
	r.POST("/", func(c *gin.Context) {

		inputBASE, _ := c.GetRawData()

		parsedInputBASE := strings.TrimPrefix(string(inputBASE), `{"lol":"`)
		parsedInputBASE = strings.TrimSuffix(parsedInputBASE, `"}`)
		parsedInputBASE = strings.ReplaceAll(parsedInputBASE, "=", "")

		parsedOut := ""

		output := scan(parsedInputBASE)

		for _, out := range output {
			parsedOut += out.rectangle + " "
		}

		c.String(200, parsedOut)

	})

	fmt.Println("Server is running...")
	r.Run()
}

type detectedObjects struct {
	rectangle string
}

func scan(inputImage string) []detectedObjects {
	// webcam, _ := gocv.VideoCaptureDevice(0)
	// webcam.Set(3, 640)
	// webcam.Set(4, 480)

	// frame := gocv.NewMat()
	// webcam.Read(&frame)
	inputImage = inputImage[strings.IndexByte(inputImage, ',')+1:]

	parsedImage, err := base64.RawStdEncoding.Strict().DecodeString(inputImage)

	if err != nil {
		log.Fatal(err)
	}

	frame, _ := gocv.IMDecode(parsedImage, gocv.IMReadFlag(-1))

	gocv.MedianBlur(frame, &frame, 5) // Apply blur to reduce noise
	hsv := gocv.NewMat()
	gocv.CvtColor(frame, &hsv, gocv.ColorBGRToHSV)

	// First Filter Detection
	lowerBlue := gocv.NewMatFromScalar(gocv.NewScalar(100, 50, 50, 0), gocv.MatTypeCV8UC3)
	upperBlue := gocv.NewMatFromScalar(gocv.NewScalar(255, 190, 100, 0), gocv.MatTypeCV8UC3)
	blueMask := gocv.NewMat()
	gocv.InRange(hsv, lowerBlue, upperBlue, &blueMask)
	blueCnts := gocv.FindContours(blueMask, gocv.RetrievalExternal, gocv.ChainApproxSimple)

	var objects []detectedObjects

	if len(blueCnts) > 0 {
		for _, contour := range blueCnts {
			if gocv.ContourArea(contour) >= 1500 {
				rect := gocv.BoundingRect(contour)
				gocv.Rectangle(&frame, rect, color.RGBA{0, 0, 255, 0}, 2)
				object := detectedObjects{rect.String()}
				objects = append(objects, object)
			}
		}
	}

	gocv.IMWrite("out.png", frame)
	return objects
}
