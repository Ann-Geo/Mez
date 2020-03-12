package latcontroller

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"image"
	"io"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/Ann-Geo/Mez/api/controller"

	"gocv.io/x/gocv"

	"google.golang.org/grpc"
)

type Controller struct {
	jaadSim       *dataSet
	jaadMed       *dataSet
	jaadComp      *dataSet
	dukeSim       *dataSet
	dukeMed       *dataSet
	dukeComp      *dataSet
	ipaddr        string
	targetAcc     float64
	targetLat     float64
	targetDataset *dataSet
	prevFrame     gocv.Mat
	frameRate     int
	sumLatError   float64
}

type knobComb struct {
	knobSetting string
	accuracy    float64
}

type lookupTable struct {
	imSize []float64
	kc     []knobComb
}

type dataSet struct {
	csvFile        string
	knobFile       string
	lpTable        *lookupTable
	firstFrameName string
	firstFrameMat  gocv.Mat
	frameDiffKS    []int
	firstFrameRead gocv.Mat
}

func NewController(ipaddr, imPath string) *Controller {
	fKnobVals := []int{0, 234100, 254500, 271000, 281000,
		0, 354000, 366000, 380000, 391300,
		0, 475000, 501700, 512000, 523600,
		0, 513000, 530000, 540000, 553500,
		0, 218800, 225500, 232500, 237000,
		0, 239500, 250000, 257000, 263000}
	dukeSim := newDataSet(imPath+"Mez_upload_woa/latcontroller/duke/simple.csv",
		imPath+"Mez_upload_woa/latcontroller/duke/simple_knobs.txt",
		imPath+"Mez_upload_woa/latcontroller/duke/firstframes/093232.png", fKnobVals[15:20])
	dukeMed := newDataSet(imPath+"Mez_upload_woa/latcontroller/duke/medium.csv",
		imPath+"Mez_upload_woa/latcontroller/duke/medium_knobs.txt",
		imPath+"Mez_upload_woa/latcontroller/duke/firstframes/086667.png", fKnobVals[20:25])
	dukeComp := newDataSet(imPath+"Mez_upload_woa/latcontroller/duke/complex.csv",
		imPath+"Mez_upload_woa/latcontroller/duke/complex_knobs.txt",
		imPath+"Mez_upload_woa/latcontroller/duke/firstframes/071878.png", fKnobVals[25:30])
	jaadSim := newDataSet(imPath+"Mez_upload_woa/latcontroller/jaad/simple.csv",
		imPath+"Mez_upload_woa/latcontroller/jaad/simple_knobs.txt",
		imPath+"Mez_upload_woa/latcontroller/jaad/firstframes/00110.png", fKnobVals[0:5])
	jaadMed := newDataSet(imPath+"Mez_upload_woa/latcontroller/jaad/medium.csv",
		imPath+"Mez_upload_woa/latcontroller/jaad/medium_knobs.txt",
		imPath+"Mez_upload_woa/latcontroller/jaad/firstframes/00050.png", fKnobVals[5:10])
	jaadComp := newDataSet(imPath+"Mez_upload_woa/latcontroller/jaad/complex.csv",
		imPath+"Mez_upload_woa/latcontroller/jaad/complex_knobs.txt",
		imPath+"Mez_upload_woa/latcontroller/jaad/firstframes/00230.png", fKnobVals[10:15])

	return &Controller{
		dukeSim:  dukeSim,
		dukeMed:  dukeMed,
		dukeComp: dukeComp,
		jaadSim:  jaadSim,
		jaadMed:  jaadMed,
		jaadComp: jaadComp,
		ipaddr:   ipaddr,
	}
}

//creates dataset dtructure by calling loadknobs and readCsv
func newDataSet(csvFile, knobFile, firstFrameName string, fknobs []int) *dataSet {
	var knobs []string
	knobs = loadKnobs(knobFile)

	lpTable := readCsv(csvFile, knobs)

	//fmt.Println(lpTable)

	//convert first frame to Mat
	img := gocv.IMRead(firstFrameName, gocv.IMReadColor)
	buffer := img.ToBytes()
	mat, err := gocv.NewMatFromBytes(img.Rows(), img.Cols(), img.Type(), buffer)
	if err != nil {
		log.Fatalf("bytes to Mat conversion failed")
	}

	//fmt.Println("F knob values", fknobs)

	return &dataSet{
		csvFile:        csvFile,
		knobFile:       knobFile,
		lpTable:        lpTable,
		firstFrameName: firstFrameName,
		firstFrameMat:  mat,
		frameDiffKS:    fknobs,
		firstFrameRead: img,
	}
}

/************helpers********************************/

//extracts the knobs string from a line in the knob file
func between(value string, a string, b string) string {
	// Get substring between two strings.
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

//opens the knob file and create knobs slice
func loadKnobs(knobFile string) []string {
	file, err := os.Open(knobFile)
	if err != nil {
		log.Fatal("Cannot load knob file", err)
	}
	defer file.Close()

	var knobs []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		knobs = append(knobs, between(scanner.Text(), "[", "]"))
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("scanner error", err)
	}

	return knobs
}

//reads csv file and populates the lookup table structure (slices)
func readCsv(csvFile string, knobs []string) *lookupTable {

	lpTable := &lookupTable{
		imSize: make([]float64, 0),
		kc:     make([]knobComb, 0),
	}

	//Opencsv file
	csvfile, err := os.Open(csvFile)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	defer csvfile.Close()

	// Parse the file
	r := csv.NewReader(csvfile)

	var i int
	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("Csv read error", err)
		}

		//get image size
		sz, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Fatalln("image size string parse error", err)
		}
		lpTable.imSize = append(lpTable.imSize, sz)

		//get accuracy
		acc, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Fatalln("accuracy string parse error", err)
		}
		lpTable.kc = append(lpTable.kc, knobComb{knobSetting: knobs[i], accuracy: acc})

		i = i + 1

	}

	return lpTable
}

/*******************helpers end***********************************************/

func (s *Controller) StartLatencyController() {
	s.frameRate = 5

	log.Println("Starting controller")
	lis, err := net.Listen("tcp", s.ipaddr) //port:9002
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	controller.RegisterLatencyControllerServer(grpcServer, s)

	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve %v\n", err)
	}
}

//choose dataset, regime and extract accuracy value from accuracy string
//format = "0.45 jaad complex"
func (s *Controller) chooseDataset(accStr string) {
	s.targetAcc, _ = strconv.ParseFloat(strings.Split(accStr, " ")[0], 64)
	dsetName := strings.Split(accStr, " ")[1]
	regime := strings.Split(accStr, " ")[2]

	if dsetName == "jaad" {
		if regime == "simple" {
			s.targetDataset = s.jaadSim
		} else if regime == "medium" {
			s.targetDataset = s.jaadMed
		} else {
			s.targetDataset = s.jaadComp
		}

	} else if dsetName == "duke" {
		if regime == "simple" {
			s.targetDataset = s.dukeSim
		} else if regime == "medium" {
			s.targetDataset = s.dukeMed
		} else {
			s.targetDataset = s.dukeComp
		}

	}

}

func (s *Controller) findInitialSize(lat float64) float64 {
	//coefficients from regression model
	m := 0.02982
	c := 2.13746
	imSize := (lat - c) / m

	return imSize
}

func (s *Controller) SetTarget(ctx context.Context, targets *controller.Targets) (*controller.Status, error) {

	//fmt.Println("SetTarget invoked")

	targetLat, err := strconv.ParseFloat(targets.GetTargetLat(), 64)
	if err != nil {
		log.Fatalln("target latency string parse error", err)
	}
	s.targetLat = targetLat
	s.chooseDataset(targets.GetTargetAcc())

	status := &controller.Status{
		Status: true,
	}
	return status, nil
}

func (s *Controller) findSizeDelta(currLatAvg, targetLat float64) float64 {
	//fmt.Println("findSizeDelta invoked", currLatAvg)
	latDiff := math.Abs(currLatAvg - targetLat)
	var sizeDelta float64
	var Kp float64
	var Ki float64

	s.sumLatError += latDiff
	if currLatAvg == 0 {
		sizeDelta = 0
	} else if latDiff < 10 {
		Kp = 3568.9
		Ki = 2446.32
	} else if latDiff >= 10 && latDiff < 20 {
		Kp = 4523.67
		Ki = 4094.9
	} else if latDiff >= 20 && latDiff < 40 {
		Kp = 4860.33
		Ki = 2979.11
	} else if latDiff >= 40 && latDiff < 80 {
		Kp = 3661.5
		Ki = 7189.39
	} else if latDiff >= 80 && latDiff < 100 {
		Kp = 3546.87
		Ki = 7375.38
	} else if latDiff >= 100 && latDiff < 200 {
		Kp = 2470.48
		Ki = 6127.73
	} else {
		Kp = 2149.77
		Ki = 2807.32

	}

	sizeDelta = Kp*latDiff + Ki*s.sumLatError

	return sizeDelta

}

func (s *Controller) findClosest(key float64) int {

	//fmt.Println("findClosest invoked", key)

	var low int
	var high = len(s.targetDataset.lpTable.imSize) - 1
	var diff = 3000000.0
	ind := 0

	for low <= high {
		mid := low + (high-low)/2
		if math.Abs(key-s.targetDataset.lpTable.imSize[mid]) < diff {
			diff = math.Abs(key - s.targetDataset.lpTable.imSize[mid])
			ind = mid
		}
		if s.targetDataset.lpTable.imSize[mid] < key {
			low = mid + 1
		} else {
			high = mid - 1
		}

	}

	//fmt.Println("index found", ind)
	return ind

}

//find knobs from lpTable
func (s *Controller) findKnobs(newImSize float64) (float64, string, float64) {
	//fmt.Println("findKnobs invoked", newImSize)
	var acc float64
	var knob string
	var ind int
	//fmt.Println("initial acc", acc)
	//fmt.Println("targetacc", s.targetAcc)

	for acc < s.targetAcc {
		ind = s.findClosest(newImSize)
		newImSize = s.targetDataset.lpTable.imSize[ind]
		//fmt.Println("newImage size", newImSize)
		knob = s.targetDataset.lpTable.kc[ind].knobSetting
		//fmt.Println("newKnob", knob)
		acc = s.targetDataset.lpTable.kc[ind].accuracy
		//fmt.Println("new accuracy", acc)

	}

	return newImSize, knob, acc

}

/**************************image modification****************************/
func (s *Controller) modifyImage(knob string, imbytes []uint8, resultFile *os.File) []uint8 {
	fmt.Println(knob)
	var emptyByte []uint8

	t1 := gocv.GetTickCount()
	mat_array, err := gocv.NewMatFromBytes(s.targetDataset.firstFrameRead.Rows(),
		s.targetDataset.firstFrameRead.Cols(), s.targetDataset.firstFrameRead.Type(), imbytes)
	if err != nil {
		log.Fatalf("gocv: Bytes to Mat conversion failed")
	}

	if knob[25:27] != "F1" {
		//fmt.Println("went F", knob[25:27])
		mat_array = s.applyFrameDifferencing(knob[25:27], mat_array)

		if mat_array.Empty() == true {
			return emptyByte
		}

		s.prevFrame = mat_array

	}

	if knob[1:3] != "R2" {
		//fmt.Println("went R", knob[1:3])
		mat_array = s.changeResolution(knob[1:3], mat_array)

	}

	if knob[7:9] != "C1" {
		//fmt.Println("went C", knob[7:9])
		mat_array = s.changeColorspace(knob[7:9], mat_array)

	}

	if knob[13:15] != "K1" {
		//fmt.Println("went K", knob[13:15])
		mat_array = s.changeSmoothingFilterSize(knob[13:15], mat_array)
	}

	//convert returned numpy array to bytes
	imgBytes := mat_array.ToBytes()

	t2 := gocv.GetTickCount()
	t := (t2 - t1) / gocv.GetTickFrequency()
	//fmt.Println("total modification time", t*1000)
	fmt.Fprintf(resultFile, "controller latency: %f\n", t*1000)

	return imgBytes

}

func (s *Controller) changeSmoothingFilterSize(ker string, mat_array gocv.Mat) gocv.Mat {
	var filterSize int
	if ker == "K2" {
		filterSize = 5
	} else if ker == "K3" {
		filterSize = 8
	} else if ker == "K4" {
		filterSize = 10
	} else {
		filterSize = 15
	}

	gocv.Blur(mat_array, &mat_array, image.Pt(filterSize, filterSize))

	return mat_array
}

func (s *Controller) changeResolution(res string, mat_array gocv.Mat) gocv.Mat {
	var width int
	var height int
	if res == "R3" {
		width = 960
		height = 540
	} else if res == "R4" {
		width = 640
		height = 360
	} else {
		width = 480
		height = 270
	}
	resultImage := gocv.NewMatWithSize(width, height, gocv.MatTypeCV8U)
	gocv.Resize(mat_array, &resultImage, image.Pt(resultImage.Rows(), resultImage.Cols()), 0, 0, gocv.InterpolationCubic)

	return resultImage

}

func (s *Controller) changeColorspace(col string, mat_array gocv.Mat) gocv.Mat {

	var colKnob gocv.ColorConversionCode
	if col == "C2" {
		colKnob = gocv.ColorBGRToGray
	} else if col == "C3" {
		colKnob = gocv.ColorBGRToHSV
	} else if col == "C4" {
		colKnob = gocv.ColorBGRToLab
	} else {
		colKnob = gocv.ColorBGRToLuv
	}
	resultImage := gocv.NewMat()

	gocv.CvtColor(mat_array, &resultImage, colKnob)

	return resultImage

}

func (s *Controller) applyFrameDifferencing(f string, mat_array gocv.Mat) gocv.Mat {
	var dif int
	if f == "F2" {
		dif = s.targetDataset.frameDiffKS[1]
	} else if f == "F3" {
		dif = s.targetDataset.frameDiffKS[2]
	} else if f == "F4" {
		dif = s.targetDataset.frameDiffKS[3]
	} else {
		dif = s.targetDataset.frameDiffKS[4]
	}

	//fmt.Println("dif==", dif)
	grayImageCurr := gocv.NewMat()
	grayImagePrev := gocv.NewMat()
	gocv.CvtColor(mat_array, &grayImageCurr, gocv.ColorBGRToGray)
	gocv.CvtColor(s.targetDataset.firstFrameMat, &grayImagePrev, gocv.ColorBGRToGray)

	diffIm := gocv.NewMat()
	gocv.AbsDiff(grayImageCurr, grayImagePrev, &diffIm)
	nonZeroCount := gocv.CountNonZero(diffIm)
	//fmt.Println("Non zero count", nonZeroCount)

	if nonZeroCount > dif {
		return mat_array
	} else {
		return gocv.NewMat()
	}

}

/*************************************************************************/

func (s *Controller) Control(stream controller.LatencyController_ControlServer) error {

	resultFile, err := os.Create("cont_lat.txt")
	if err != nil {
		log.Fatalf("Cannot create result file %v\n", err)
	}

	defer resultFile.Close()

	//fmt.Println("Control invoked")
	s.prevFrame = s.targetDataset.firstFrameMat

	var currentLat float64
	var currLatAvg float64
	imCount := 0
	var newImSize float64
	//prevImSize = initialSize
	prevImSize := s.findInitialSize(s.targetLat)
	knob := "'R2', 'C1', 'K1', 'D1', 'F1'"
	acheivedAcc := "0.4" //max accuracy
	var acc float64

	for {

		imCount = imCount + 1
		req, err := stream.Recv()
		//fmt.Println("received from ES broker---", time.Now())
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return err
		}

		//fmt.Println("received from EN broker")

		curLat, _ := strconv.ParseFloat(strings.Split(req.GetCurrentLat(), "and")[1], 64)
		currentLat += curLat

		if imCount == s.frameRate {
			currLatAvg = currentLat / float64(s.frameRate)

			sizeDelta := s.findSizeDelta(currLatAvg, s.targetLat)
			if currLatAvg-s.targetLat > 0 {

				if prevImSize != s.targetDataset.lpTable.imSize[0] {
					newImSize = prevImSize - sizeDelta
				} else {
					newImSize = prevImSize
				}

			} else {
				if prevImSize != s.targetDataset.lpTable.imSize[len(s.targetDataset.lpTable.imSize)-1] {
					newImSize = prevImSize + sizeDelta
				} else {
					newImSize = prevImSize
				}

			}

			newImSize, knob, acc = s.findKnobs(newImSize)
			acheivedAcc = fmt.Sprintf("%f", acc)
			prevImSize = newImSize

			imCount = 0
			currentLat = 0
		}

		modImBytes := s.modifyImage(knob, req.GetImage(), resultFile)
		if len(modImBytes) == 0 {
			continue
		}

		//fmt.Println("modified image length", len(modImBytes))

		imRecdTime := strings.Split(req.GetCurrentLat(), "and")[0]

		resp := &controller.CustomImage{
			Image:       modImBytes,
			AcheivedAcc: imRecdTime + "and" + acheivedAcc,
		}

		//fmt.Println(resp.AcheivedAcc)

		sendErr := stream.Send(resp)

		//fmt.Println()
		if sendErr != nil {
			log.Fatalf("Error while sending data to EN broker: %v", sendErr)
			return sendErr
		}

	}

}
