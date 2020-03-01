package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	uiLine         = "--------------------------------------------------\n"
	defaultTimeout = 3 // second
)

var (
	flagConfigFilePath = flag.String("c", "config.json", "config file path")
)

func init() {
	clearTerminal()
}

func getIntroStr() string {
	buf := bytes.Buffer{}
	buf.WriteString("\n")
	buf.WriteString(uiLine)
	buf.WriteString(alignCenter("Remote File Sender", len(uiLine)))
	buf.WriteString("\n")
	buf.WriteString(uiLine)
	return buf.String()
}

func clearTerminal() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	// TODO: will be supported per os type and terminal type
	// osType := runtime.GOOS
	// if osType == "linux" {
	// 	cmd := exec.Command("clear")
	// 	cmd.Stdout = os.Stdout
	// 	cmd.Run()

	// } else if osType == "window" {
	// 	cmd := exec.Command("cmd", "/c", "cls")
	// 	cmd.Stdout = os.Stdout
	// 	cmd.Run()
	// }
}

func selectHosts(config *Config) {
	scanner := bufio.NewScanner(os.Stdin)
	selectedIndexes := []int{}

	fmt.Print(getIntroStr())
	fmt.Print(config.Hosts, " > ")

	for scanner.Scan() {
		line := scanner.Text()
		trimedLine := strings.TrimSpace(line)

		if trimedLine == "q" || trimedLine == "quit" || trimedLine == "exit" {
			os.Exit(0)
		}

		if trimedLine == "n" || trimedLine == "next" {
			clearTerminal()
			return
		}

		for _, v := range strings.Fields(line) {
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				// handle error
				continue
			}

			if !isValidIndex(int(i), len(config.Hosts)) {
				// invalid index
				continue
			}

			selectedIndexes = append(selectedIndexes, int(i)-1)
		}

		for _, selected := range selectedIndexes {
			config.Hosts[selected].isSelected = !config.Hosts[selected].isSelected
		}

		selectedIndexes = []int{}
		clearTerminal()
		fmt.Print(getIntroStr())
		fmt.Print(config.Hosts, " > ")
	}
}

func selectFiles(config *Config) bool {
	scanner := bufio.NewScanner(os.Stdin)
	selectedIndexes := []int{}

	fmt.Print(getIntroStr())
	fmt.Print(config.Files, " > ")

	for scanner.Scan() {
		line := scanner.Text()
		trimedLine := strings.TrimSpace(line)

		if trimedLine == "q" || trimedLine == "quit" || trimedLine == "exit" {
			os.Exit(0)
		}

		if trimedLine == "p" || trimedLine == "pre" {
			clearTerminal()
			return false
		}

		if trimedLine == "n" || trimedLine == "next" {
			clearTerminal()
			return true
		}

		for _, v := range strings.Fields(line) {
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				// handle error
				continue
			}

			if !isValidIndex(int(i), len(config.Files)) {
				// invalid index
				continue
			}

			selectedIndexes = append(selectedIndexes, int(i)-1)
		}

		for _, selected := range selectedIndexes {
			config.Files[selected].isSelected = !config.Files[selected].isSelected
		}

		selectedIndexes = []int{}
		clearTerminal()
		fmt.Print(getIntroStr())
		fmt.Print(config.Files, " > ")
	}

	return true
}

func sendFiles(config *Config) {
	failedHosts := []*Host{}
	failedFiles := map[string]error{}

	for _, host := range config.Hosts {
		if !host.isSelected {
			continue
		}

		sftpConfig := &ssh.ClientConfig{
			User: host.User,
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
			Auth: []ssh.AuthMethod{
				ssh.Password(host.Password),
			},
			Timeout: time.Second * defaultTimeout,
		}
		sftpConfig.SetDefaults()

		// connect
		conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host.IP, host.Port), sftpConfig)
		if err != nil {
			// handle error
			failedHosts = append(failedHosts, host)
			fmt.Println(err)
			continue
		}
		defer conn.Close()

		// create new SFTP client
		client, err := sftp.NewClient(conn)
		if err != nil {
			// handle error
			failedHosts = append(failedHosts, host)
			continue
		}
		defer client.Close()

		for _, file := range config.Files {
			if !file.isSelected {
				continue
			}

			if !pathExists(file.Src) {
				failedFiles[file.Src] = fmt.Errorf("[Remote File Sender] %s Could Not Be Found", file.Src)
				continue
			}

			if isDir(file.Src) {
				srcDir, err := os.Open(file.Src)
				if err != nil {
					failedFiles[file.Src] = err
				}
				defer srcDir.Close()

				files, err := srcDir.Readdir(-1)
				if err != nil {
					failedFiles[file.Src] = err
				}

				client.MkdirAll(file.Dest)

				for _, f := range files {
					if isDir(file.Src) {
						continue
					}

					srcPath := filepath.Join(file.Src, f.Name())
					srcFile, err := os.Open(srcPath)
					if err != nil {
						failedFiles[srcPath] = err
					}

					defer srcFile.Close()

					destFile, err := client.Create(client.Join(file.Dest, f.Name()))
					if err != nil {
						failedFiles[srcPath] = err
						continue
					}
					defer destFile.Close()

					_, err = io.Copy(destFile, srcFile)
					if err != nil {
						failedFiles[file.Src] = err
						continue
					}

				}

			} else if isFile(file.Src) {
				srcFile, err := os.Open(file.Src)
				if err != nil {
					failedFiles[file.Src] = err
				}

				defer srcFile.Close()

				destFile, err := client.Create(file.Dest)
				if err != nil {
					failedFiles[file.Src] = err
					continue
				}
				defer destFile.Close()

				_, err = io.Copy(destFile, srcFile)
				if err != nil {
					failedFiles[file.Src] = err
					continue
				}
			} else {
				// unexpected case
			}
		} // for files
	} // for hosts

	fmt.Printf("Failed Hosts: %d\n", len(failedHosts))
	fmt.Printf("Failed Files: %d\n", len(failedFiles))
}

func saveConfig() {

}

func main() {
	flag.Parse()

	config := &Config{}
	err := config.init(*flagConfigFilePath)
	if err != nil {
		log.Fatal(err)
	}

	for {
		selectHosts(config)
		res := selectFiles(config)
		if res {
			break
		}
	}

	sendFiles(config)

}
