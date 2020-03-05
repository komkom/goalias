package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var (
	AliasInUse  = fmt.Errorf(`alias-in-use`)
	deleteAlias string
)

func init() {
	flag.StringVar(&deleteAlias, "d", "", "")
}

func main() {
	flag.Parse()

	if len(os.Args) != 2 {
		fmt.Printf("usage: alias name\n")
		os.Exit(1)
	}

	alias := os.Args[1]

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	zshenv := home + `/.zshenv`

	_, err = os.Stat(zshenv)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	if os.IsNotExist(err) {
		fmt.Printf("no zshenv found")
		os.Exit(1)
	}

	err = insertAlias(zshenv, dir, alias)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	err = reloadZsh()
	if err != nil {
		panic(err)
	}
}

func insertAlias(zshenv string, dir string, alias string) error {

	file, err := os.OpenFile(zshenv, os.O_RDONLY, 0755)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	var inserted bool
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		ok := hasAliasInLine(line, alias)
		if ok {
			buf.WriteString(aliasConfig(alias, dir))
			buf.WriteString("\n")
			inserted = true

		} else {

			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	if !inserted {
		buf.WriteString(aliasConfig(alias, dir))
		buf.WriteString("\n")
	}

	ioutil.WriteFile(zshenv, buf.Bytes(), 0755)

	return nil
}

func hasAliasInLine(line string, alias string) bool {

	fields := strings.Fields(line)
	if len(fields) >= 2 {
		if fields[0] == `alias` {
			fields := strings.Split(fields[1], `=`)
			if len(fields) == 2 {
				if fields[0] == alias {
					return true
				}
			}
		}
	}
	return false
}

func aliasConfig(alias, dir string) string {
	return fmt.Sprintf("alias %v=%v", alias, dir)
}

func reloadZsh() error {
	_, err := exec.Command("killall", "-USR1", "zsh").Output()
	if err != nil {
		return err
	}
	return nil
}
