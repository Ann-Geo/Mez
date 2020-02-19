package latcontroller

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"vsc_workspace/Mez_upload/api/controller"

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
	frameDiffKS    []int32
	firstFrameRead gocv.Mat
}

//creates dataset dtructure by calling loadknobs and readCsv
func newDataSet(csvFile, knobFile, firstFrameName string, fknobs []int32) *dataSet {
	var knobs []string
	knobs = loadKnobs(knobFile)

	var lpTable *lookupTable
	lpTable = readCsv(csvFile, knobs)

	//convert first frame to Mat
	img := gocv.IMRead(firstFrameName, gocv.IMReadColor)
	buffer := img.ToBytes()
	mat, err := gocv.NewMatFromBytes(img.Rows(), img.Cols(), img.Type(), buffer)
	if err != nil {
		log.Fatalf("bytes to Mat conversion failed")
	}

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

	var lpTable *lookupTable

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
	fKnobVals := []int32{0, 250000, 283000, 308000, 332000,
		0, 377500, 401500, 430000, 445000,
		0, 507000, 540000, 551500, 580000,
		0, 520000, 537000, 546000, 557800,
		0, 225000, 229300, 236800, 241500,
		0, 245000, 256000, 261300, 267500}
	s.dukeSim = newDataSet("duke/simple.csv", "duke/simple_knobs.txt", "duke/firstframes/093232.png", fKnobVals[15:20])

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

	var low int
	var high = len(s.targetDataset.lpTable.imSize)
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
	return ind

}

//find knobs from lpTable
func (s *Controller) findKnobs(newImSize float64) (float64, string, float64) {
	var acc float64
	var knob string
	var ind int

	for acc < s.targetAcc {
		ind = s.findClosest(newImSize)
		newImSize = s.targetDataset.lpTable.imSize[ind]
		knob = s.targetDataset.lpTable.kc[ind].knobSetting
		acc = s.targetDataset.lpTable.kc[ind].accuracy

	}

	return newImSize, knob, acc

}

/**************************image modification****************************/
func (s *Controller) modifyImage(knob string, imbytes []uint8) {
	mat_array, err := gocv.NewMatFromBytes(s.targetDataset.firstFrameRead.Rows(),
		s.targetDataset.firstFrameRead.Cols(), s.targetDataset.firstFrameRead.Type(), imbytes)
	if err != nil {
		log.Fatalf("gocv: Bytes to Mat conversion failed")
	}

	if knob[25:27] != "F1" {
		mat_array = s.applyFrameDifferencing(knob[25:27], mat_array)

		if len(mat_array) == 0 {
			return ""
		}

		s.prevFrame = mat_array

	}

	if knob[1:3] != "R2" {
		mat_array = s.changeResolution(knob[1:3], mat_array)

	}

	if knob[7:9] != "C1" {
		mat_array = s.changeColorspace(knob[7:9], mat_array)

	}

	if knob[13:15] != "K1" {
		mat_array = s.changeSmoothingFilterSize(knob[13:15], mat_array)
	}

	//convert returned numpy array to bytes


	return mat_array

}

func (s *Controller) applyFrameDifferencing(f string, mat_array gocv.Mat) {
	var diff int32
	if f=="F2" {
		dif = s.targetDataset.frameDiffKS[1]
	} else if f=="F3" {
		dif = s.targetDataset.frameDiffKS[2]
	} else if f=="F4" {
		dif = s.targetDataset.frameDiffKS[3]
	} else {
		dif = s.targetDataset.frameDiffKS[4]
	}
	

curr = cv2.cvtColor(image_array, cv2.COLOR_BGR2GRAY)
prev = cv2.cvtColor(prev_frame, cv2.COLOR_BGR2GRAY)


	non_zero_count = cv2.countNonZero(cv2.absdiff(curr, prev))

#print(non_zero_count)

	if non_zero_count > dif:
			return image_array
else:
	return np.zeros(0)

}

/*************************************************************************/

func (s *Controller) Control(stream controller.LatencyController_ControlServer) error {
	s.prevFrame = s.targetDataset.firstFrameMat

	var currentLat float64
	var currLatAvg float64
	imCount := 0
	var newImSize float64
	//prevImSize = initialSize
	prevImSize := s.findInitialSize(s.targetLat)
	knob := "'R2', 'C1', 'K1', 'D1', 'F1'"
	acheivedAcc := "0.4" //max accuracy

	for {

		imCount = imCount + 1
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return err
		}

		imRecTimeAndCurrLat, _ := strconv.ParseFloat(strings.Split(req.GetCurrentLat(), "and")[1], 64)
		currentLat += imRecTimeAndCurrLat

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

			newImSize, knob, acc := s.findKnobs(newImSize)
			acheivedAcc = fmt.Sprintf("%f", acc)
			prevImSize = newImSize

			imCount = 0
			currentLat = 0
		}

		modImBytes = s.modifyImage(knob, req.GetImage())
		if len(modImBytes) == 0 {
			continue
		}

	}

}
