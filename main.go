package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

const (
	Post    string = "POST"
	Get     string = "GET"
	Put     string = "PUT"
	Patch   string = "PATCH"
	Delete  string = "DELETE"
	Options string = "OPTIONS"
)

type HarBody struct {
	Size     int32  `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
}

type HarQueryString struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HarHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HarResponse struct {
	Status          int32       `json:"status"`
	StatusText      string      `json:"statusText,omitempty"`
	HttpVersion     string      `json:"httpVersion"`
	Headers         []HarHeader `json:"headers,omitempty"`
	Cookies         []any       `json:"cookies,omitempty"`
	Content         HarBody     `json:"content,omitempty"`
	ServerIPAddress string      `json:"serverIPAddress,omitempty"`
	RedirectURL     string      `json:"redirectUrl,omitempty"`
}

type HarRequest struct {
	Method      string           `json:"method"` // enum
	Url         string           `json:"url"`
	Headers     []HarHeader      `json:"headers,omitempty"`
	QueryString []HarQueryString `json:"queryString,omitempty"`
	Cookies     []any            `json:"cookies,omitempty"`
	PostData    HarBody          `json:"postData,omitempty"`
}

type HarEntry struct {
	Request  HarRequest  `json:"request,omitempty"`
	Response HarResponse `json:"response,omitempty"`
}

type HarCreator struct {
	Name    string  `json:"name,omitempty"`
	Version float64 `json:"version,string"`
}

type HarLog struct {
	Version float32    `json:"version,string"`
	Creator HarCreator `json:"creator,omitempty"`
	Entries []HarEntry `json:"entries,omitempty"`
}

type HarFile struct {
	Log HarLog
}

func main() {
	var exclude_params string
	var output string

	flag.StringVar(&output, "o", "output.txt", "Send Output to File (filename)")
	flag.StringVar(&exclude_params, "e", "", "Exclude Params Values (name=value)")

	flag.Parse()

	// if true {
	// 	var exclude_values []struct {string; string}
	// 	// test_value := make(chan struct {string; string})
	//
	// 	for _, param := range strings.Split(exclude_params, ",") {
	// 		if param_value := strings.Split(param, ":"); len(param_value) != 0 {
	// 			var name string = param_value[0]
	// 			var value string = param_value[1]
	// 			// a := types.NewTuple(&name, &value)
	// 		// 	exclude_values
	// 		// struct {Name; Value}{name, value}
	// 		}
	//
	// 		// append(exclude_values, types.*NewTuple(x ...*types.Var)())
	// 	}
	//
	// 	log.Fatal(exclude_values)
	// }

	filename := flag.Arg(0)
	if filename := flag.Arg(0); filename == "" {
		log.Fatalln("Must provide a '.har' file.")
	}

	if filepath.Ext(filename) != ".har" {
		log.Fatalln("File extension error.", filepath.Ext(filename), "is not valid file.", "Must pass a valid '.har' file")
	}

	filedir := filepath.Dir(filename)
	filename = filepath.Base(filename)

	thepath, err := filepath.Abs(filedir)

	if err != nil {
		log.Fatal(err)
	}

	contents, err := os.ReadFile(filepath.Join(thepath, filename))
	if err != nil {
		log.Fatalln("File reading error", err)
	}

	var result HarFile
	json.Unmarshal([]byte(contents), &result)

	// check dir exists, if not create
	if _, err := os.Stat(output); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(output), 0700)
	}

	// create and get file
	file_output, err := os.Create(output)
	if err != nil {
		log.Fatal(err)
	}
	defer file_output.Close()

	file_output.WriteString(time.Now().Format("2006-1-2 15:4"))

	for _, entry := range result.Log.Entries {

		if entry.Request.Method == Options || entry.Request.Method == Get {
			continue
		}

		if strings.Contains(entry.Request.Url, "businessHierarch") || strings.Contains(entry.Request.Url, "program") {
			printToFile(*file_output, entry)
		}

	}

	fmt.Println("----")
	fmt.Println("Parsed to:", output)
}

// fmt.Println(entry.Request.Method, ":", entry.Request.Url)
//
//	if entry.Request.PostData.Text != "" {
//		fmt.Println("Request Body:")
//		fmt.Println(entry.Request.PostData.Text)
//	}
//
//	if entry.Response.Content.Text != "" {
//		fmt.Println("Response Body:")
//		fmt.Println(entry.Response.Content.Text)
//	}

func jsonToLine(src string) string {
	line := &bytes.Buffer{}
	if err := json.Compact(line, []byte(src)); err != nil {
		panic(err)
	}
	return line.String()
}

func writeSpace(file os.File, length int32) {
	for range make([]int, length) {
		file.WriteString("\n")
	}
}

func printToFile(file os.File, entry HarEntry) {
	line := strings.Join([]string{entry.Request.Method, ":", entry.Request.Url}, " ")

	writeSpace(file, 2)

	file.WriteString(line)
	for _, query := range entry.Request.QueryString {
		writeSpace(file, 1)
		file.WriteString(strings.Join([]string{"Query String: ", query.Name, "->", query.Value}, " "))
	}
	writeSpace(file, 2)

	if entry.Request.PostData.Text != "" {
		file.WriteString("Request Body:")
		writeSpace(file, 2)
		file.WriteString(jsonToLine(entry.Request.PostData.Text))
		writeSpace(file, 2)
	}
	if entry.Response.Content.Text != "" {
		file.WriteString("Response Body:")
		writeSpace(file, 2)
		file.WriteString(jsonToLine(entry.Response.Content.Text))
		writeSpace(file, 2)
	}

	file.WriteString("----")
	writeSpace(file, 3)
}
