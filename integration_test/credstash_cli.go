package test

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
)

// Instead of specifying the table name for each CLI function, we create an object so we can specify it once
type credstashCli struct {
	tableName string
}

// Create a CLI object with the default table name
func newCredstashCli() *credstashCli {
	p := new(credstashCli)
	p.tableName = "credential-store"
	return p
}

// Create a CLI object and specify a custom table name
func newCredstashCliCustomTable(tableName string) *credstashCli {
	p := new(credstashCli)
	p.tableName = tableName
	return p
}

func (cli *credstashCli) list() {
	args := []string{"-r", "us-east-1", "-t", cli.tableName, "list"}
	out, err := exec.Command("credstash", args...).CombinedOutput()
	fmt.Println("ran credstash")
	if err != nil {
		log.Println(string(out))
		log.Fatal(err)
	}

	log.Print(string(out))
}

/*
	Put a credstash secret. If version is 0 it will auto increment
*/
func (cli *credstashCli) put(name string, value string, version int) {
	args := []string{"-t", cli.tableName, "put"}
	if version == 0 {
		args = append(args, "-a", name, value)
	} else {
		args = append(args, "-v", strconv.Itoa(version), name, value)
	}

	out, err := exec.Command("credstash", args...).Output()

	if err != nil {
		log.Println(string(out))
		log.Fatal(err)
	}
}

/*
	Delete a credstash secret. Will Delete all version
*/
func (cli *credstashCli) delete(name string) {
	args := []string{"-t", cli.tableName, "delete", name}
	out, err := exec.Command("credstash", args...).CombinedOutput()

	if err != nil {
		log.Println(string(out))
		log.Fatal(err)
	}
}

/*
	Get a credstash secret. If version is 0 it will pull latest version
*/
func (cli *credstashCli) get(name string, version int) string {
	args := []string{"-t", cli.tableName, "get", "-n"}
	if version == 0 {
		args = append(args, name)
	} else {
		args = append(args, "-v", strconv.Itoa(version), name)
	}

	out, err := exec.Command("credstash", args...).Output()

	if err != nil {
		log.Println(args)
		log.Println(out)
		log.Fatal(err)
	}

	return string(out)
}

/*
	Get the last version number of a secret from credstash
*/
func (cli *credstashCli) getLatestVersion(name string) string {
	// cmdString := "credstash -t " + cli.tableName + " list | grep " + name + " | tail -1 | sed --regexp-extended 's/.*?version 0*([1-9][0-9]*).*/\\1/'"
	cmdString := "credstash -t " + cli.tableName + " list | grep " + name + " | tail -1"
	out, err := exec.Command("bash", "-c", cmdString).Output()
	re := regexp.MustCompile(`.*?version 0*([1-9][0-9]*).*`)
	version := re.FindStringSubmatch(string(out))
	if err != nil {
		log.Println(out)
		log.Println(version)
		log.Fatal(err)
	}
	outString := version[1]
	fmt.Println(outString)
	return string(outString)

}
