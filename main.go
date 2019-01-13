package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

var (
	exeFile string
	exeDir  string
	id      string
	destDir string
)

func init() {
	exeFile, _ := os.Executable()
	exeDir = filepath.Dir(exeFile)
}

func main() {
	flag.StringVar(&destDir, "d", exeDir, "Save location. If not specified, directory at executable file.\n")

	flag.StringVar(&id, "i", "", "Your tenhou.net id 'ID########-########'\n"+
		"If not specified, it refers to a file named id.txt or id \nin the directory containing mjloget")

	flag.Parse()

	if s, err := os.Stat(destDir); os.IsNotExist(err) || !s.IsDir() {
		fmt.Fprintf(os.Stderr, "-d:%v does not exist or is not directory\n", destDir)
		os.Exit(1)
	}

	if id == "" {
		gid, err := getID()
		if err != nil {
			log.Fatal(err)
		}
		id = gid
	}

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func getID() (string, error) {
	idFile := []string{"id.txt", "id"}
	for _, v := range idFile {
		target := filepath.Join(exeDir, v)
		file, err := os.Open(target)
		if err != nil {
			continue
		}
		defer file.Close()

		sc := bufio.NewScanner(file)
		for sc.Scan() {
			return sc.Text(), nil
		}
	}
	return "", fmt.Errorf("couldn't find id file")
}

func run() error {
	resp, err := http.Get("https://tenhou.net/0/log/find.cgi?un=" + id)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	r := regexp.MustCompile(`cgi\?log=(20\S*tw=.)`)
	logs := r.FindAllSubmatch(body, -1)

	for _, log := range logs {
		_, err := os.Stat(filepath.Join(destDir, string(log[1])+".mjlog"))
		if err != nil {
			err := getLog(string(log[1]))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getLog(log string) error {
	logpath := filepath.Join(destDir, log+".mjlog")
	fmt.Println(logpath)
	resp, err := http.Get("https://tenhou.net/0/log/find.cgi?log=" + log)
	if err != nil {
		return fmt.Errorf("failed to download log %s", log)
	}
	defer resp.Body.Close()
	file, err := os.Create(logpath)
	if err != nil {
		return fmt.Errorf("failed to make mjlog file")
	}
	defer file.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read logfile")
	}
	file.Write(body)

	return nil
}
