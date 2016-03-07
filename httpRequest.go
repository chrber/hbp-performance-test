package main

import "fmt"
import (
	"net/http"
	"io/ioutil"
	"net/url"
	"time"
	"math/rand"
	"strconv"
	"github.com/op/go-logging"
	"os"
)

var log = logging.MustGetLogger("HttpRequestLogger")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.5s} %{id:03x}%{color:reset} %{message}`,
)
// logging
var backend = logging.NewLogBackend(os.Stderr, "", 0)
var backendFormatter = logging.NewBackendFormatter(backend, format)
var backendLeveled = logging.AddModuleLevel(backend)


func setup() (dictionary map[int]map[string]int) {

	backendLeveled.SetLevel(logging.INFO, "")
	logging.SetBackend(backendLeveled, backendFormatter)

	// value dictionary
	dictionary = map[int]map[string]int {
		0:map[string] int{"stack":0, "slice":3699, "x":20, "y":20},
		1:map[string] int{"stack":0, "slice":3700, "x":10, "y":10},
		2:map[string] int{"stack":0, "slice":3694, "x":5,  "y":5} }

	for outerMapKey := range dictionary {
		log.Debugf("Map for key %v: ", outerMapKey)
		log.Debug("\n")
		innerMap := dictionary[outerMapKey]
		for key := range innerMap{
			log.Debugf("Key: %v Value: %s", key, strconv.Itoa(innerMap[key]))
		}
		log.Debug("=========\n")
	}
	return dictionary
}

func createRandomValuesForLevel (level int, dictionary map[int]map[string] int) (stack, slice, x, y int) {
	log.Debugf("Stack: %v", dictionary[level]["stack"])
	log.Debugf("Slice: %v", dictionary[level]["slice"])
	log.Debugf("x: %v", dictionary[level]["x"])
	log.Debugf("y: %v", dictionary[level]["y"])
	//stack = rand.Intn(dictionary[level]["stack"])
	stack = 0
	slice = rand.Intn(dictionary[level]["slice"])
	x = rand.Intn(dictionary[level]["x"])
	y = rand.Intn(dictionary[level]["y"])
	return
}

func createRandTileRequest (dictionary map[int]map[string]int) (url string) {
	prefix := "/image/v0/api/bbic?fname="
	imagePath := "/srv/data/HBP/BigBrain_jpeg.h5"
	suffix := "&mode=ims&prog="

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	log.Debugf("Random: %s",strconv.Itoa(rand.Intn(20)))

	level := rand.Intn(2)
	log.Debugf("Level: %s",strconv.Itoa(level))

	stack, slice, x, y := createRandomValuesForLevel(level, dictionary)
	//             TILE%200%20{stack}%20{level}%20{slice}%20{x}%20{y}%20none%2010%201
	//tileString := "TILE+0+%c+%c+%c+%c+%c+none+10+1"
	tileString := "TILE+0+%d+%d+%d+%d+%d+none+10+1"
	tileString = fmt.Sprintf(tileString, stack, level, slice, x, y)
	log.Infof("TileString: %s", tileString)
	url = fmt.Sprintf("%s%s%s%s", prefix, imagePath, suffix, tileString)
	return url
}

func main() {
	var dictionary map[int]map[string] int= setup()
	urlSuffix := createRandTileRequest(dictionary)

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
	//log.Printf("StartTime: %q", startTime.String())
	resp, err := http.Get(u.String())
	endTime := time.Since(startTime)
	defer resp.Body.Close()
	log.Info("Time for request:"+endTime.String())

	if err != nil {
		fmt.Println("Error: {}", err)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		// OK
		fmt.Print(string(bodyBytes))
	} else {
		log.Errorf("Response status code: %i. The error was: %v", resp.StatusCode, err)
	}

	if err != nil {
		log.Fatal(err)
	}

}
