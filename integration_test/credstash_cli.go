package test

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

/*
	Put a credstash secret. If version is 0 it will auto increment
*/
func credstash_put(name string, value string, version int) {

	var args []string
	if version == 0 {
		args = []string{"put", "-a", name, value}
	} else {
		args = []string{"put", "-v", strconv.Itoa(version), name, value}
	}

	out, err := exec.Command("credstash", args...).Output()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(out))
}

/*
	Delete a credstash secret. Will Delete all version
*/
func credstash_delete(name string) {
	args := []string{"delete", name}

	out, err := exec.Command("credstash", args...).Output()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(out))
}

/*
	Get a credstash secret. If version is 0 it will pull latest version
*/
func credstash_get(name string, version int) string {

	var args []string
	if version == 0 {
		args = []string{"get", "-n", name}
	} else {
		args = []string{"get", "-n", "-v", strconv.Itoa(version), name}
	}

	out, err := exec.Command("credstash", args...).Output()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(out))
	return string(out)
}

/*
	Get the last version number of a secret from credstash
*/
func credstash_latest_version(name string) string {
	cmdString := "credstash list | grep " + name + " | tail -1 | sed --regexp-extended 's/.*?version 0*([1-9][0-9]*).*/\\1/'"

	out, err := exec.Command("bash", "-c", cmdString).Output()
	if err != nil {
		log.Fatal(err)
	}
	outString := strings.TrimSpace(string(out))
	fmt.Println(outString)
	return string(outString)

}
