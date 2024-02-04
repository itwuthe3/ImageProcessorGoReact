package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gocv.io/x/gocv"
)

func resizeImage(img gocv.Mat, width, height int) gocv.Mat {
	resized := gocv.NewMat()
	gocv.Resize(img, &resized, image.Point{width, height}, 0, 0, gocv.InterpolationLinear)
	return resized
}

func applyAntialiasing(img gocv.Mat) gocv.Mat {
	antialiased := gocv.NewMat()
	// 目的のサイズを指定して gocv.Resize を行う
	gocv.Resize(img, &antialiased, image.Point{X: img.Cols(), Y: img.Rows()}, 0, 0, gocv.InterpolationArea)
	return antialiased
}

func applySmoothingFilter(img gocv.Mat) gocv.Mat {
	smoothed := gocv.NewMat()
	gocv.GaussianBlur(img, &smoothed, image.Point{5, 5}, 0, 0, gocv.BorderDefault)
	return smoothed
}

func applyGaussianFilter(img gocv.Mat) gocv.Mat {
	gaussian := gocv.NewMat()
	gocv.GaussianBlur(img, &gaussian, image.Point{0, 0}, 2, 0, gocv.BorderDefault)
	return gaussian
}

func applyUnsharpMask(img gocv.Mat) gocv.Mat {
	blurred := gocv.NewMat()
	gocv.GaussianBlur(img, &blurred, image.Point{0, 0}, 5, 0, gocv.BorderDefault)

	unsharpMask := gocv.NewMat()
	gocv.Subtract(img, blurred, &unsharpMask)

	return unsharpMask
}
func processImageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10 MB max file size
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error reading the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if file == nil {
		http.Error(w, "Received nil file", http.StatusBadRequest)
		return
	}

	// Save the received image
	savePath := "received_image.jpg"
	saveFile, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "Error saving the received image", http.StatusInternalServerError)
		return
	}
	defer saveFile.Close()

	_, err = io.Copy(saveFile, file)
	if err != nil {
		http.Error(w, "Error saving the received image", http.StatusInternalServerError)
		return
	}

	// Check if the saved file is nil
	if saveFile == nil {
		http.Error(w, "Saved nil file", http.StatusInternalServerError)
		return
	}

	// Attempt to decode the saved image
	img := gocv.IMRead(savePath, gocv.IMReadColor)
	if img.Empty() {
		fmt.Println("Error reading image file:", savePath)
		http.Error(w, "Error reading the image file", http.StatusInternalServerError)
		return
	}

	mat := img

	// Apply image processing based on parameters
	// Example: Resize
	if resizeParam, err := strconv.ParseBool(r.FormValue("resize")); err == nil && resizeParam {
		mat = resizeImage(mat, 300, 200) // Example: Resize to 300x200
	}

	// Apply other image processing based on parameters
	if antialiasingParam, err := strconv.ParseBool(r.FormValue("antialiasing")); err == nil && antialiasingParam {
		fmt.Println("Applying antialiasing")
		mat = applyAntialiasing(mat)
	}
	if smoothingParam, err := strconv.ParseBool(r.FormValue("smoothing")); err == nil && smoothingParam {
		fmt.Println("Applying smoothing filter")
		mat = applySmoothingFilter(mat)
	}
	if gaussianParam, err := strconv.ParseBool(r.FormValue("gaussian")); err == nil && gaussianParam {
		fmt.Println("Applying gaussian filter")
		mat = applyGaussianFilter(mat)
	}
	if unsharpMaskParam, err := strconv.ParseBool(r.FormValue("unsharpMask")); err == nil && unsharpMaskParam {
		fmt.Println("Applying unsharp mask")
		mat = applyUnsharpMask(mat)
	}

	// Convert processed image to PNG
	var imgBuffer bytes.Buffer
	success := gocv.IMWrite("processed_image.png", mat)
	if !success {
		fmt.Println("Error encoding the image")
		http.Error(w, "Error encoding the image", http.StatusInternalServerError)
		return
	}

	// Write the processed image to the buffer
	if _, err := imgBuffer.Write(mat.ToBytes()); err != nil {
		fmt.Println("Error writing the processed image to buffer:", err)
		http.Error(w, "Error writing the processed image to buffer", http.StatusInternalServerError)
		return
	}

	// Send the processed image back to the client
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(imgBuffer.Bytes())))
	if _, err := w.Write(imgBuffer.Bytes()); err != nil {
		fmt.Println("Error sending the processed image:", err)
		http.Error(w, "Error sending the processed image", http.StatusInternalServerError)
		return
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/process", processImageHandler).Methods("POST")
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("."))))

	headers := handlers.AllowedHeaders([]string{"Content-Type"})
	origins := handlers.AllowedOrigins([]string{"http://localhost:5173"})
	methods := handlers.AllowedMethods([]string{"POST", "GET"})

	fmt.Println("Server listening on :8080")
	http.ListenAndServe(":8080", handlers.CORS(headers, origins, methods)(r))
}
