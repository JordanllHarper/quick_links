package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const helpMessage = `

A quick way to open your favourite locations from the CLI.

1. Adding items

To add a location, use the add subcommand.

ql add -name <YOUR_NAME> -location <YOUR_LOCATION>
Example: ql add -name yt -location https://www.youtube.com

2. Using your item

To use the registered item, use it in the place of a subcommand.

ql <YOUR_NAME>

Example:
ql add -name yt -location https://www.youtube.com
ql yt

3. Removing items

To remove an item you've registered, use the remove subcommand.

ql remove <YOUR_NAME>

Example:
ql add -name yt -location https://www.youtube.com
ql remove yt

4. Listing items

To list out your quick links, use the list subcommand.

ql list

`

type nameToURL map[string]string

const quickListDir = ".ql"
const quickListFile = "list.json"

func userError(message string) {
	fmt.Println(message)
	os.Exit(1)
}

func main() {

	ntu, err := readFromFile()
	if err != nil {
		log.Fatalln("Something went wrong while reading from the quick list file:", err)
	}

	showHelpMessage := flag.Bool("help", false, "Shows the help message")

	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addName := addCmd.String("name", "", "Name of the quick list item.")
	addLocation := addCmd.String("location", "", "Location which should open. Note: make sure to include https if adding a web item!")

	flag.Parse()
	args := flag.Args()

	if *showHelpMessage {
		fmt.Print(helpMessage)
		return
	}

	if len(args) < 1 {
		fmt.Print(helpMessage)
		userError("Missing subcommand")
	}

	switch args[0] {
	case "list":
		fmt.Println("Name : Location")
		fmt.Println()
		for k, v := range ntu {
			fmt.Println(k, ":", v)
		}
		return
	case "add":
		addCmd.Parse(args[1:])
		if strings.TrimSpace(*addName) == "" {
			userError(fmt.Sprintln("Invalid or empty name", *addName))
		}
		if strings.TrimSpace(*addLocation) == "" {
			userError(fmt.Sprintln("Invalid or empty location", *addLocation))
		}
		ntu[*addName] = *addLocation
	case "remove":
		name := args[1]
		_, ok := ntu[name]
		if !ok {
			userError(fmt.Sprintf("Specified name \"%s\" has not been registered\n", name))
		}
		delete(ntu, name)
	default:
		name := args[0]
		url, exists := ntu[name]
		if !exists {
			userError(fmt.Sprintf("Specified name \"%s\" has not been registered\n", name))
		}
		var cmd string
		var args []string

		switch runtime.GOOS {
		case "windows":
			cmd = "cmd"
			args = []string{"/c", "start"}
		case "darwin":
			cmd = "open"
		default: // "linux", "freebsd", "openbsd", "netbsd"
			cmd = "xdg-open"
		}
		args = append(args, url)
		cmdHandle := exec.Command(cmd, args...)
		if err := cmdHandle.Start(); err != nil {
			log.Fatalln("Error whilst trying to launch browser with location:", err)
		}

		if err := cmdHandle.Wait(); err != nil {
			log.Fatalln("Error whilst waiting for launch browser command to finish:", err)
		}

		return
	}

	if err := writeToFile(ntu); err != nil {
		log.Fatalln("Something went wrong while writing to file:", err)
	}

}

func writeToFile(ntu nameToURL) error {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return errors.Join(errors.New("writeToFile err"), err)
	}

	dir := fmt.Sprintf("%s/%s", userHome, quickListDir)

	if _, err = os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(dir, 0777)
	}
	if err != nil {
		return errors.Join(errors.New("writeToFile err"), err)
	}

	file := fmt.Sprintf("%s/%s", dir, quickListFile)
	f, err := os.Create(file)
	if err != nil {
		return errors.Join(errors.New("writeToFile err"), err)
	}

	defer f.Close()

	return json.NewEncoder(f).Encode(ntu)
}

func readFromFile() (nameToURL, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Join(errors.New("readFromFile err"), err)
	}

	dir := fmt.Sprintf("%s/%s", userHome, quickListDir)
	if _, err = os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(dir, 0777)
	}
	if err != nil {
		return nil, errors.Join(errors.New("readFromFile err"), err)
	}

	file := fmt.Sprintf("%s/%s", dir, quickListFile)
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Join(errors.New("readFromFile err"), err)
	}

	var ntu nameToURL
	if err = json.NewDecoder(f).Decode(&ntu); err != nil {
		return nil, err
	}

	return ntu, nil
}
