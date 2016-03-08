package main

import "fmt"
import (
	"net/http"
	"io/ioutil"
	"net/url"
	"time"
	"math/rand"
	"strconv"
	log "github.com/Sirupsen/logrus"
	"os"
)

// Global Variables
var requestParameterDictionary map[int]map[string]int;
const numberOfBunches = 4
const bunchSize int = 8
var requestTimes = [numberOfBunches][bunchSize]time.Duration{}
var pointerToRequestTimes = &requestTimes

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{})
	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.ErrorLevel)
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
	imagePath := "/srv/data/HBP/BigBrain_jpeg.h5"
	suffix := "&mode=ims&prog="

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	log.Debugf("Random: %s",strconv.Itoa(rand.Intn(20)))

	level := rand.Intn(2)
	log.Debugf("Level: %s",strconv.Itoa(level))

	stack, slice, x, y := createRandomValuesForLevel(level)
	tileString := "TILE+0+%d+%d+%d+%d+%d+none+10+1"
	tileString = fmt.Sprintf(tileString, stack, level, slice, x, y)
	log.Infof("TileString: %s", tileString)
	url = fmt.Sprintf("%s%s%s%s", prefix, imagePath, suffix, tileString)
	return url
}

func fireTileRequest(bunchNumber int, requestNumber int, urlSuffix string) {
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
	requestTimes[bunchNumber][requestNumber] = endTime

	defer resp.Body.Close()
	log.Info("Time for request:"+endTime.String())

	if err != nil {
		log.Errorf("{}", err)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		// OK
		log.Info("#######################################")
		log.Infof("Request number %d: %s\n", requestNumber, string(bodyBytes))
	} else {
		log.Errorf("Response status code: %i. The error was: %v", resp.StatusCode, err)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func createRequestBunch(bunchNumber int) {
	for requestNumber := 0; requestNumber < bunchSize; requestNumber++ {
		fireTileRequest(bunchNumber, requestNumber, createRandTileRequest())
	}
}

//var sem = make(chan int, 5)

func main() {
	setup()

	for bunchNumber:=0; bunchNumber<numberOfBunches; bunchNumber++ {
		createRequestBunch(bunchNumber)
	}

	for bunchMapKey := range requestTimes {
		bunchMap := requestTimes[bunchMapKey]
		fmt.Printf("Bunch number %d\n", bunchMapKey)
		for requestNumber := range bunchMap {
			fmt.Printf("Print time for request %d: %s\n", requestNumber, bunchMap[requestNumber].String())
		}
	}
	//urlSuffix := createRandTileRequest()
	//fireTileRequest(urlSuffix)

	//startTime := time.Now()
	//time.Sleep(3*time.Second)
	//endTime := time.Since(startTime)
	//log.Info("Time for request:"+endTime.String())
	//i := 0

	//startTime := time.Now()
	//for i < 5 {
	//	sem <- 1
	//	go func() {
	//		time.Sleep(3*time.Second)
	//		<-sem
	//	}()
	//	i++
	//	fmt.Println(i)
	//}
	//endTime := time.Since(startTime)
	//log.Info("Time for request:"+endTime.String())




}
