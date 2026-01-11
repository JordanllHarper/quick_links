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

	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addName := addCmd.String("name", "", "Name of the quick list item.")
	addURL := addCmd.String("url", "", "URL which should open. Note: make sure to include https at the start!")

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		userError("Missing subcommand")
	}

	switch args[0] {
	case "list":
		fmt.Println("Name: <URL>")
		fmt.Println("---")
		for k, v := range ntu {
			fmt.Println(k, ":", v)
		}
		return
	case "add":
		addCmd.Parse(args[1:])
		if strings.TrimSpace(*addName) == "" {
			userError(fmt.Sprintln("Invalid or empty name", *addName))
		}
		if strings.TrimSpace(*addURL) == "" {
			userError(fmt.Sprintln("Invalid or empty url", *addURL))
		}
		ntu[*addName] = *addURL
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
		if err := exec.Command(cmd, args...).Start(); err != nil {
			log.Fatalln("Error whilst trying to launch browser with url:", err)
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
