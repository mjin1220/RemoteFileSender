package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// Config is a struct
type Config struct {
	Hosts Hosts `json:"hosts"`
	Files Files `json:"files"`
}

// Hosts is a Host's slice
type Hosts []*Host

// Host is a struct
type Host struct {
	Name       string `json:"name"`
	IP         string `json:"ip"`
	Port       string `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	isSelected bool
}

// Files is a File's slice
type Files []*File

// File is a struct
type File struct {
	Name       string `json:"name"`
	Src        string `json:"src"`
	Dest       string `json:"dest"`
	isSelected bool
}

func (c *Config) init(path string) error {
	if path == "" {
		return fmt.Errorf("[Remote File Sender] Config File Path is Not Setup")
	}

	if !pathExists(path) {
		path = os.Getenv("RFS_CONFIG")
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("[Remote File Sender] %s Could Not Be Found", path)
		}
		return fmt.Errorf("[Remote File Sender] %s Could Not Open Config File Path: %w", path, err)
	}

	if err := json.NewDecoder(f).Decode(c); err != nil {
		return fmt.Errorf("[Remote File Sender] %s Could Not Be Decoded To Json: %w", path, err)
	}

	return nil
}

func (hs Hosts) String() string {
	buf := bytes.Buffer{}

	buf.WriteString(fmt.Sprintf("%2s%d. %s\n", "", 1, "Select Hosts"))
	buf.WriteString(uiLine)
	buf.WriteString(fmt.Sprintf("%3s  %3s %-25s %-15s\n", "*", "No.", "Name", "IP"))
	buf.WriteString(uiLine)

	for i, h := range hs {
		prefix := ""
		if h.isSelected {
			prefix = "*"
		}
		buf.WriteString(fmt.Sprintf("%3s [%2d] %-25s %15s\n", prefix, i+1, h.Name, h.IP))
	}

	buf.WriteString("\n")
	return buf.String()
}

func (fs Files) String() string {
	buf := bytes.Buffer{}

	buf.WriteString(fmt.Sprintf("%2s%d. %s\n", "", 2, "Select Files"))
	buf.WriteString(uiLine)
	buf.WriteString(fmt.Sprintf("%3s  %3s  %-20s %s\n", "*", "No.", "Name", "Source File Path"))
	buf.WriteString(fmt.Sprintf("%3s  %3s  %-20s %s\n", "", "", "", "Destination File Path"))
	buf.WriteString(uiLine)

	for i, f := range fs {
		prefix := ""
		if f.isSelected {
			prefix = "*"
		}
		buf.WriteString(fmt.Sprintf("  %1s [%2d]  %-20s %s\n", prefix, i+1, f.Name, f.Src))
		buf.WriteString(fmt.Sprintf("  %1s  %2s   %-20s %s\n\n", "", "", "", f.Dest))
	}

	buf.WriteString("\n")
	return buf.String()
}
