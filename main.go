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
		return
	}

	var brightness_arg string


	for _, arg := range args {
		if arg != "-i" ||
		arg != "-l" ||
		arg != "-m" {
			if !ArrayIncludes(interfaces, arg) {
				brightness_arg = arg
			}
		}
	}

	if brightness_arg != "" {
		brightness_arg = strings.ReplaceAll(brightness_arg, "max", strconv.Itoa(max_brightness))
		if err := ArgMath(brightness_arg, &brightness_arg); err != nil {
			panic(err)
		}
		selected_brightness, _ = strconv.Atoi(brightness_arg)
	} else {
		selected_brightness = 0
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

func ArgMath(arg string, result *string) error {
	b, op := ContainsMath(arg)
	if !b {
		return nil
	}
	var ops = []string{"-", "+", "*", "/"}
	for _, operator := range ops {
		if op == operator {
			spl := strings.SplitN(arg, operator, 2)
			var left, right string
			var solved string

			left = spl[0]
			right = spl[1]

			if c, _ := ContainsMath(spl[0]); c {
				ArgMath(spl[0], &left)
			}

			if c, _ := ContainsMath(spl[1]); c {
				ArgMath(spl[1], &right)
			}


			if err := DoMath(left, right, op, &solved); err != nil {
				return err
			}

			*result = solved
		}
	}
	return nil
}

func ContainsMath(str string) (bool, string) {
	for _, op := range []string{"-", "+", "*", "/"} {
		if strings.Contains(str, op) { return true, op }
	}
	return false, ""
}

func DoMath(str1, str2, operator string, result *string) error {
	var parsed1, parsed2 int
	var err error
	parsed1, err = strconv.Atoi(str1)
	if err != nil { return err }
	parsed2, err = strconv.Atoi(str2)
	if err != nil { return err }
	if operator == "/" { *result = strconv.Itoa(parsed1/parsed2)  }
	if operator == "*" { *result = strconv.Itoa(parsed1*parsed2)   }
	if operator == "+" { *result = strconv.Itoa(parsed1+parsed2) }
	if operator == "-" { *result = strconv.Itoa(parsed1-parsed2)   }
	return nil
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
