package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	apiUrl = "https://rc-place.fly.dev"
	max_x  = 100
	max_y  = 100
)

var BearerToken string = os.Getenv("RC_TOKEN")

type PixelStruct struct {
	Color       string    `json:"color"`
	X           int       `json:"x"`
	Y           int       `json:"y"`
	LastUpdated time.Time `json:"lastUpdated"`
	LastEditor  string    `json:"lastEditor"`
}

func main() {
	checkToken()
	colorsAvailable := []string{"black", "forest", "green", "lime", "blue", "cornflowerblue", "sky", "cyan", "red", "burnt-orange", "orange", "yellow", "purple", "hot-pink", "pink", "white"}
	generalColor := getGeneralBoardColor()
	fmt.Println("Seems like most of the pixels are: ", generalColor)

	randomColor := randomColorPicker(colorsAvailable)
	fmt.Println("Changing pixels to: ", randomColor)
	if generalColor == randomColor {
		fmt.Println("The general color is the same as the random color, not changing anything.")
		return
	}

	// get user input for number of pixels to update
	var numPixelsToUpdate int
	fmt.Println("How many pixels would you like to update?")
	fmt.Scanf("%d", &numPixelsToUpdate)
	if numPixelsToUpdate > max_x*max_y {
		fmt.Println("That's too many pixels, I'm only going to update", max_x*max_y, "pixels.")
		numPixelsToUpdate = max_x * max_y
	}

	fmt.Println("Updating pixels...")

	updateRandomPixels(generalColor, randomColor, numPixelsToUpdate)
}

func checkToken() {
	if BearerToken == "" {
		fmt.Println("Please set the environment variable RC_TOKEN")
		os.Exit(1)
	}
}

// function that takes an originating color and updates random pixels to a new given color if it matches the originating color
func updateRandomPixels(originatingColor string, newColor string, numPixelsToUpdate int) {
	for i := 0; i < numPixelsToUpdate; i++ {
		x, y := randomCoordinatePicker()
		color, _ := getPixelState(x, y)
		if color == originatingColor {
			updatePixelState(x, y, newColor)
		}
	}
}

// This is a rough estimate of the general color of the board.
// TODO: get a batch state of the board?
func getGeneralBoardColor() string {
	// not the prettiest
	var colorCount = map[string]int{
		"black":          0,
		"forest":         0,
		"green":          0,
		"lime":           0,
		"blue":           0,
		"cornflowerblue": 0,
		"sky":            0,
		"cyan":           0,
		"red":            0,
		"burnt-orange":   0,
		"orange":         0,
		"yellow":         0,
		"purple":         0,
		"hot-pink":       0,
		"pink":           0,
		"white":          0,
	}

	sampleSize := 20 // 2% of the board, but is pretty good
	for i := 0; i < sampleSize; i++ {
		color, _ := getPixelState(randomCoordinatePicker())
		colorCount[color]++
	}

	// find the most common color
	var mostCommonColor string
	var mostCommonCount int
	for color, count := range colorCount {
		if count > mostCommonCount {
			mostCommonColor = color
			mostCommonCount = count
		}
	}

	return mostCommonColor
}

func getPixelState(x int, y int) (string, error) {
	url := fmt.Sprintf("%s/tile?x=%d&y=%d", apiUrl, x, y)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", BearerToken))

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	// parse body for color
	var p PixelStruct
	err = json.NewDecoder(res.Body).Decode(&p)
	return (p.Color), err
}

func updatePixelState(x int, y int, color string) error {
	url := fmt.Sprintf("%s/tile", apiUrl)
	method := "POST"

	bodyString := fmt.Sprintf(`{"x": %d, "y": %d, "color": "%s"}`, x, y, color)

	payload := strings.NewReader(bodyString)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", BearerToken))

	_, err = client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return err
}

// function that given a string array returns a random string from the array
func randomColorPicker(strArray []string) string {
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(strArray))
	return strArray[randomIndex]
}

// function that gives a random x and y coordinate given the max x and y
func randomCoordinatePicker() (int, int) {
	rand.Seed(time.Now().UnixNano())
	x := rand.Intn(max_x)
	y := rand.Intn(max_y)
	return x, y
}
