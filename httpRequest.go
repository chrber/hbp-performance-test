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
)

// Global Variables
const (
	AverageTimeOverBunches = 1
	AverageTimeOverRequests = 2
	DetailedTimesPerRequest = 3
)
const reportDetail = AverageTimeOverBunches // Values can be AverageTimeOverBunches |  AverageTimeOverRequests |  DetailedTimesPerRequest
const numberOfBunches = 10
const bunchSize int = 8
const imagePath = "/srv/data/HBP/BigBrain_jpeg.h5"

var requestParameterDictionary map[int]map[string]int;
const channelBuffer = numberOfBunches * bunchSize
type Result struct {
	bunchNumber int
	requestNumber int
	requestTime time.Duration
}

var fromCreateRequestBunch = make(chan Result, channelBuffer)

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

	// value dictionary
	requestParameterDictionary = map[int]map[string]int {
		0:map[string] int{"stack":0, "slice":3699, "x":20, "y":20},
		1:map[string] int{"stack":0, "slice":3700, "x":10, "y":10},
		2:map[string] int{"stack":0, "slice":3694, "x":5,  "y":5} }

	for outerMapKey := range requestParameterDictionary {
		log.Debugf("Map for key %v: ", outerMapKey)
		log.Debug("\n")
		innerMap := requestParameterDictionary[outerMapKey]
		for key := range innerMap{
			log.Debugf("Key: %v Value: %s", key, strconv.Itoa(innerMap[key]))
		}
		log.Debug("=========\n")
	}
}

func createRandomValuesForLevel (level int) (stack, slice, x, y int) {
	log.Debugf("Stack: %v", requestParameterDictionary[level]["stack"])
	log.Debugf("Slice: %v", requestParameterDictionary[level]["slice"])
	log.Debugf("x: %v", requestParameterDictionary[level]["x"])
	log.Debugf("y: %v", requestParameterDictionary[level]["y"])
	//stack = rand.Intn(requestParameterDictionary[level]["stack"])
	stack = 0
	slice = rand.Intn(requestParameterDictionary[level]["slice"])
	x = rand.Intn(requestParameterDictionary[level]["x"])
	y = rand.Intn(requestParameterDictionary[level]["y"])
	return
}

func createRandTileRequest () (url string) {
	prefix := "/image/v0/api/bbic?fname="
	suffix := "&mode=ims&prog="

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	log.Debugf("Random: %s",strconv.Itoa(rand.Intn(20)))

	level := rand.Intn(2)
	log.Debugf("Level: %s",strconv.Itoa(level))

	stack, slice, x, y := createRandomValuesForLevel(level)
	tileString := "TILE+0+%d+%d+%d+%d+%d+none+10+1"
	tileString = fmt.Sprintf(tileString, stack, level, slice, x, y)
	log.Debugf("TileString: %s", tileString)
	url = fmt.Sprintf("%s%s%s%s", prefix, imagePath, suffix, tileString)
	log.Debugf("URL: %v", url)
	return url
}

func fireTileRequest(bunchNumber int, requestNumber int, urlSuffix string) Result {
	u, err := url.Parse(urlSuffix)
	if err != nil {
		log.Fatal(err)
	}
	u.Scheme = "http"
	u.Host = "hbp-image.desy.de:8888"
	q := u.Query()
	//q.Set("q", "golang")
	u.RawQuery = q.Encode()
	log.Debugf("Tile request URL: %q", u.String())

	startTime := time.Now()
	resp, err := http.Get(u.String())
	endTime := time.Since(startTime)

	// request time write into globally visible array
	//requestTimes[bunchNumber][requestNumber] = endTime

	defer resp.Body.Close()
	log.Debug("Time for request:"+endTime.String())

	if err != nil {
		log.Errorf("{}", err)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		// OK
		log.Debug("#######################################")
		log.Debugf("Request number %d: %s\n", requestNumber, string(bodyBytes))
	} else {
		log.Errorf("Response status code: %i. The error was: %v", resp.StatusCode, err)
	}

	if err != nil {
		log.Fatal(err)
	}

	return  Result{bunchNumber, requestNumber, endTime}
}

func createRequestBunch(bunchNumber int) {
	for requestNumber := 0; requestNumber < bunchSize; requestNumber++ {
		result := fireTileRequest(bunchNumber, requestNumber, createRandTileRequest())
		fromCreateRequestBunch <-result
	}
}

//var sem = make(chan int, 5)

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

	for bunchNumber:=0; bunchNumber<numberOfBunches; bunchNumber++ {
		for requestNumber:=0; requestNumber<bunchSize; requestNumber++ {
			result := <-fromCreateRequestBunch
			log.Debugf("Result for bunch %v, request %v: %v", result.bunchNumber, result.requestNumber, result.requestTime)
			requestTimes[result.bunchNumber][result.requestNumber] = result.requestTime
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
	log.Infof("Number of bunches: %v", numberOfBunches)
	log.Infof("Number of requests/bunch: %v", bunchSize)
	log.Infof("Average request time: %v", average)
}
