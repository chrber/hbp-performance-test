package main

import "fmt"
import (
	"net/http"
	"io/ioutil"
	"net/url"
	"time"
	"math/rand"
	"strconv"
	// https://github.com/Sirupsen/logrus
	log "github.com/Sirupsen/logrus"
	"os"
	"strings"
	"encoding/json"
	"sync"
)

// Specify number of bunches and requests per bunch
const numberOfBunches = 40
const bunchSize int = 8


// Global Variables
const (
	AverageTimeOverBunches = 1
	AverageTimeOverRequests = 2
	DetailedTimesPerRequest = 3
)

// for fixed or random tile requests
const randomTileRequests = true

// logging tile request URLs
const showRequests = false

// define level of detail in logging request times
const reportDetail = AverageTimeOverBunches // Values can be AverageTimeOverBunches |  AverageTimeOverRequests |  DetailedTimesPerRequest

// Check if correct image was returned
const checkForCorrectImage = true

// Specify the protocol to be used
const protocol = "https"

// DESY dCache Endpoint config
// F5 load balancer
//const hostname = "hbp-image.desy.de:8888"
//A10 load balancer
//const hostname = "131.169.4.31:8888"

// Xen CoreOS machine
//const hostname = "coreos01.desy.de:80"
//const hostname = "hbp-image-service.desy.de"

//Juelich Endpoint
//
const hostname = "fsd-cloud47.zam.kfa-juelich.de"

//OneData Endpoint config
//  149.156.9.143:8888/image/v0/api/bbic?fname=/srv/data
//const hostname = "149.156.9.143:8888"

// old Image data
//const imagePath = "/srv/data/HBP/BigBrain_jpeg.h5"

// new data
const imagePath = "/srv/data/HBP/template/human/bigbrain_20um/sections/bigbrain.h5"
//const imagePath = "/srv/data/HBP/template/rat/waxholm/v2/sections/whs.h5"

//NOT WORKING const imagePath = "/srv/data/HBP/template/rat/waxholm/v2/anno/whs_axial_v2.h5"
//MetaData get not working : const imagePath = "/srv/data/HBP/stacks/rat/r602/anno/r602_anno.h5"

//old data
//const imagePath = "/srv/data/HBP/BigBrain_jpeg.h5"
//new data
//const imagePath = "/srv/data/HBP/template/human/bigbrain_20um/sections/bigbrain.h5"

var mutex = &sync.Mutex{}
var predefinedStack = 0
var predefinedLevel = 0
var predefinedSlice = 0
var predefinedX = 0
var predefinedY = 0
var stacks []Stack

type Result struct {
	bunchNumber int
	requestNumber int
	requestTime time.Duration
}

// Channel used for returning Result struct of each bunch
const channelBuffer = numberOfBunches * bunchSize
var fromCreateRequestBunch = make(chan Result, channelBuffer)

// image formats and magic numbers
var magicTable = map[string]string{
	"\xff\xd8\xff":      "image/jpeg",
	"\x89PNG\r\n\x1a\n": "image/png",
	"GIF87a":            "image/gif",
	"GIF89a":            "image/gif",
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{})
	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
	//log.SetLevel(log.ErrorLevel)
	//log.SetLevel(log.DebugLevel)
}

func setup() {
	const prefix = "/image/v0/api/bbic?fname="
	const metadata_suffix = "&mode=meta" // This returns json with the metadata

	metaDataURL := fmt.Sprintf("%s%s%s", prefix, imagePath, metadata_suffix)
	stacks = getImageMetaData(metaDataURL)

	for stackNum := range stacks {
		log.Debugf("Levels in stack %d:", stackNum)
		log.Debug("=========================================================")
		for levelNum := range stacks[stackNum].Levels {
			log.Debugf("Level number %d", levelNum)
			log.Debugf("Level values: %+v", stacks[stackNum].Levels[levelNum].Attrs)
			log.Debug("--------------------------------------------------------------")
		}
		log.Debugf("----End of stack: %d--------------------------", stackNum)
	}

	//requestParameterDictionary = map[int]map[string]int {
	//	0:map[string] int{"stack":0, "slice":3699, "x":20, "y":20},
	//	1:map[string] int{"stack":0, "slice":3700, "x":10, "y":10},
	//	2:map[string] int{"stack":0, "slice":3694, "x":5,  "y":5} }

}

func createRandomValuesForLevel () (stack, level, slice, x, y int) {
	stack = rand.Intn(len(stacks))
	level = rand.Intn(len(stacks[stack].Levels))
	slice = rand.Intn(stacks[stack].Attrs.NumSlices)
	x = rand.Intn(stacks[stack].Levels[level].Attrs.NumXTiles)
	y = rand.Intn(stacks[stack].Levels[level].Attrs.NumYTiles)
	return
}

func createDeterministicValuesForLevel () (stack, level, slice, x, y int) {

	mutex.Lock()
	log.Debug("I am in")
	stack = predefinedStack
	level = predefinedLevel
	slice = predefinedSlice
	x = predefinedX
	y = predefinedY

	if (predefinedStack < len(stacks)-1) {
		predefinedStack++
	} else {
		//log.Infof("Reseting stack number: %d", predefinedStack)
		predefinedStack = 0
	}

	if (predefinedSlice < stacks[predefinedStack].Attrs.NumSlices) {
		predefinedSlice++
	} else {
		predefinedSlice = 0
	}

	if (predefinedX < stacks[predefinedStack].Levels[predefinedLevel].Attrs.NumXTiles-1) {
		predefinedX++
	} else {
		predefinedX = 0
	}

	if (predefinedY < stacks[predefinedStack].Levels[predefinedLevel].Attrs.NumYTiles-1) {
		predefinedY++
	} else {
		predefinedY = 0
	}
	return
}

func createRandTileRequest () (url string) {
	prefix := "/image/v0/api/bbic?fname="
	suffix := "&mode=ims&prog="

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	log.Debugf("Creating test random %d",rand.Intn(20))

	stack, level, slice, x, y := createRandomValuesForLevel()
	tileString := "TILE+0+%d+%d+%d+%d+%d+none+10+1"
	tileString = fmt.Sprintf(tileString, stack, level, slice, x, y)
	log.Debugf("TileString: %s", tileString)
	url = fmt.Sprintf("%s%s%s%s", prefix, imagePath, suffix, tileString)
	log.Debugf("URL: %v", url)
	return url
}

func createSpecificTileRequest () (url string) {
	prefix := "/image/v0/api/bbic?fname="
	suffix := "&mode=ims&prog="

	level := 1
	log.Debugf("Level: %s",strconv.Itoa(level))

	stack, level, slice, x, y := createDeterministicValuesForLevel()
	tileString := "TILE+0+%d+%d+%d+%d+%d+none+10+1"
	tileString = fmt.Sprintf(tileString, stack, level, slice, x, y)
	mutex.Unlock()
	log.Debugf("Mutex: I am out: %s", tileString)
	url = fmt.Sprintf("%s%s%s%s", prefix, imagePath, suffix, tileString)
	log.Debugf("URL: %v", url)
	return url
}

// mimeFromIncipit returns the mime type of an image file from its first few
// bytes or the empty string if the file does not look like a known file type
func mimeFromReturnedBytes(bytes []byte) string {
	byteStr := string(bytes)
	for magic, mime := range magicTable {
		if strings.HasPrefix(byteStr, magic) {
			return mime
		}
	}

	return ""
}

func imageReturned (bytesFromRequest []byte) (bool, string) {
	mimeType := mimeFromReturnedBytes(bytesFromRequest)
	log.Debugf("Returned mimetype: %s", mimeType)

	if (mimeType != "") {
		return true, mimeType;
	} else {
		return false, mimeType;
	}
}

func getImageMetaData(urlString string) []Stack {
	// All numbers in the metadata are zero based
	// meaning: number of stacks = 3 allows to select stack 0,1,2
	// Same is true for levels, slices, x and y values

	u, err := url.Parse(urlString)

	if err != nil {
		log.Fatal(err)
	}

	u.Scheme = protocol
	u.Host = hostname
	q := u.Query()
	//q.Set("q", "golang")
	u.RawQuery = q.Encode()
	log.Debugf("Tile request URL: %q", u.String())

	res, err := http.Get(u.String())
	log.Infof("Metadata URL: %s", u.String())

	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	//decoder := json.NewDecoder(res.Body)

	body, _ := ioutil.ReadAll(res.Body)

	//n := bytes.IndexByte(body, 0)
	//s := string(body[:n])
	//log.Infof("String returned: %s", s)

	metadataEntry := MetadataEntry{}
	if error := json.Unmarshal(body, &metadataEntry); err != nil {
		log.Errorf("Metadata for requested image did not return JSON: %s",error)
	} else {
		log.Debugf("Results from metadata query unmarshaled attribute: %+v", metadataEntry)
		log.Debugf("Number of stacks: %d", len(metadataEntry.Stacks))
		for stackNumber := range metadataEntry.Stacks {
			log.Debugf("Levels in stacks %d: %d", stackNumber, len(metadataEntry.Stacks[stackNumber].Levels))
		}
	}

	// for each stack, multiple levels
	numberOfStacks := 0
	stacks := []Stack{}
	for stackNumber := range metadataEntry.Stacks {
		numberOfStacks++
		levels := []Level{}
        	numberOfLevels := 0
		for levelNumber := range metadataEntry.Stacks[stackNumber].Levels {
			attrForLevel := Level{ LevelAttrs{
				NumSlices : metadataEntry.Stacks[stackNumber].Levels[levelNumber].Attrs.NumSlices,
				NumXTiles: metadataEntry.Stacks[stackNumber].Levels[levelNumber].Attrs.NumXTiles,
				NumYTiles : metadataEntry.Stacks[stackNumber].Levels[levelNumber].Attrs.NumYTiles,
			}}
			levels = append(levels, attrForLevel)
			numberOfLevels++
		}
		stackAttrs := metadataEntry.Stacks[stackNumber].Attrs
		stackWithLevels := Stack{levels, stackAttrs}
		stacks = append(stacks, stackWithLevels)
	}
	return stacks
}

func fireTileRequest(bunchNumber int, requestNumber int, urlSuffix string) Result {
	u, err := url.Parse(urlSuffix)
	if err != nil {
		log.Fatal(err)
	}
	u.Scheme = "http"
	u.Host = hostname
	q := u.Query()
	//q.Set("q", "golang")
	u.RawQuery = q.Encode()
	log.Debugf("Tile request URL: %q", u.String())
	if (showRequests) {
		log.Infof("Tile request URL: %q", u.String())
	}

	startTime := time.Now()
	resp, err := http.Get(u.String())
	endTime := time.Since(startTime)

	defer resp.Body.Close()
	log.Debug("Time for request:"+endTime.String())

	if err != nil {
		log.Errorf("{}", err)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		// OK
		//log.Debug("#######################################")
		//log.Debugf("Request number %d: %s\n", requestNumber, string(bodyBytes))
		if (checkForCorrectImage) {
			didReturnImage, imageType := imageReturned(bodyBytes)
			if (didReturnImage) {
				log.Debugf("Correct image: %s returned for Bunch: %d Request: %d URLsuffix: %s", imageType, bunchNumber, requestNumber, urlSuffix)
			} else {
				log.Errorf("No image returned for Bunch: %d Request: %d URLsuffix: %s", bunchNumber, requestNumber, urlSuffix)
			}
		}
		return  Result{bunchNumber, requestNumber, endTime}
	} else if resp.StatusCode == 500 {
		log.Errorf("Server-side error (500) for URL: %s.", u.String())
		return Result{}
	} else {
		log.Errorf("Unknown Error with response code: %s and error %s", resp.StatusCode, err)
		return Result{}
	}

	if err != nil {
		log.Fatal(err)
	}
	return Result{}
}

func createRequestBunch(bunchNumber int) {
	for requestNumber := 0; requestNumber < bunchSize; requestNumber++ {
		var result Result;
		if randomTileRequests {
			result = fireTileRequest(bunchNumber, requestNumber, createRandTileRequest())
		} else {
			result = fireTileRequest(bunchNumber, requestNumber, createSpecificTileRequest())
		}
		fromCreateRequestBunch <-result
	}
}

func main() {
	requestTimes := map[int]map[int]time.Duration {}
	log.Debugf("RequestTimes length: %v", len(requestTimes))

	setup()

	for bunchNumber:=0; bunchNumber<numberOfBunches; bunchNumber++ {
		requestTimes[bunchNumber] = make(map[int]time.Duration)
		log.Debugf("Creating bunch %v", bunchNumber)
		go func(bunchNumber int) {
			createRequestBunch(bunchNumber)
		}(bunchNumber)
	}

	requestErrorCount := 0

	for bunchNumber:=0; bunchNumber<numberOfBunches; bunchNumber++ {

		for requestNumber:=0; requestNumber<bunchSize; requestNumber++ {
			result := <-fromCreateRequestBunch
			log.Debugf("Result for bunch %v, request %v: %v", result.bunchNumber, result.requestNumber, result.requestTime)
			if (result != Result{}) {
				requestTimes[result.bunchNumber][result.requestNumber] = result.requestTime
			} else {
				requestErrorCount++
			}
		}
	}

	average := 0*time.Millisecond
	for outerMapKey := range requestTimes {
		if (reportDetail > 2) {
			log.Info("=================================================")
			log.Info("=================================================")
			log.Infof("Map for bunch %v: ", outerMapKey)
		}
		innerMap := requestTimes[outerMapKey]
		sum := 0*time.Millisecond
		for key := range innerMap{
			if (reportDetail > 2) {
				log.Infof("Request: %v RequestTime: %s", key, innerMap[key])
			}
			sum += innerMap[key];
		}

		average += sum / numberOfBunches
		if (reportDetail > 2) {
			log.Info("=========")
		}
		if (reportDetail > 1) {
			log.Infof("==== Request time for bunch %v: =====: %v",outerMapKey, sum)
		}
		if (reportDetail > 2) {
			log.Info("=========\n")
		}
	}

	if randomTileRequests {
		log.Info("Random Tile Requests")
	} else {
		log.Info("Fixed Tile Requests")
	}
	log.Infof("Number of bunches: %v", numberOfBunches)
	log.Infof("Number of requests/bunch: %v", bunchSize)
	log.Infof("Requests/Errors: %d/%d", numberOfBunches*bunchSize, requestErrorCount)
	log.Infof("Average request time: %v", average)
}
