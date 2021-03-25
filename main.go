package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
	"os"
	"flag"
	// "os/user"
)

func main() {
	args := os.Args[1:]

	var interfaces []string
	var max_brightness int
	var selected_interface string
	var selected_brightness int

	interf := flag.String("i", "", "Select interface")
	list := flag.Bool("l", false, "List interfaces")
	max := flag.Bool("m", false, "Show max brightness")
	flag.Parse()

	if err := FetchInterfaces(&interfaces); err != nil {
		fmt.Println("Failed to fetch interfaces")
		return
	}

	if *list {
		m := fmt.Sprintf("Found these interfaces: \n%s", strings.Join(interfaces[:], ",\n"))
		fmt.Println(m)
		return
	}

	if *interf != "" {
		selected_interface = *interf
		if !ArrayIncludes(interfaces, selected_interface) {
			fmt.Println("Invalid interface selection")
			return
		}
		if err := GetMaxBrightness(selected_interface, &max_brightness); err != nil {
			fmt.Println("Could not read max_brightness")
			return
		}
	}

	if selected_interface == "" {
		if len(interfaces) == 0 {
			fmt.Println("Could not find value for interface")
			return
		}
		selected_interface = interfaces[0]
		fmt.Println("No interface selected, defaulting to " + selected_interface)
	}

	if err := GetMaxBrightness(selected_interface, &max_brightness); err != nil {
		fmt.Println("Failed to read max_brightness")
		return
	}

	if *max {
		m := fmt.Sprintf("Max brightness for %s: %d", selected_interface, max_brightness)
		fmt.Println(m)
	}

	for _, arg := range args {
		br, err := strconv.Atoi(arg)
		if err == nil { selected_brightness = br }
		if arg == "max" { selected_brightness = max_brightness }
	}

	if selected_brightness > max_brightness {
		fmt.Println("Selected brightness above maximum, setting to maximum")
		selected_brightness = max_brightness
	}

	if selected_brightness == 0 {
		return
	}

	if err := SetBrightness(selected_interface, selected_brightness); err != nil {
		fmt.Println("Failed to set brightness. Are you root?")
		return
	}
}

func FetchInterfaces(interfaces *[]string) error {
	var result []string
	files, err := ioutil.ReadDir("/sys/class/backlight")
	if err != nil { return err }
	for _, file := range files {
		result = append(result, file.Name()) 
	}
	*interfaces = result
	return nil
}

func GetMaxBrightness(interf string, brightness *int) error {
	var max int
	data, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/backlight/%s/max_brightness", interf))
	if err != nil { return err }
	max, err = strconv.Atoi(strings.Split(string(data), "\n")[0])
	if err != nil { return err }
	*brightness = max
	return nil
}

func SetBrightness(interf string, brightness int) error {
	err := ioutil.WriteFile(fmt.Sprintf("/sys/class/backlight/%s/brightness", interf), 
	[]byte(strconv.Itoa(brightness)), 0644)
	if err != nil { return err }
	return nil
}

func ArrayIncludes(array []string, value string) bool {
	for _, i := range array {
		if i == value { return true }
	}
	return false
}