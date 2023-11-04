package marlen

import (
	"io"
	"fmt"
	"os"
	"net"
	"net/http"
	"time"
)
var client *http.Client = &http.Client{
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   2 * time.Second,
			KeepAlive: 2 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   2 * time.Second,
		ResponseHeaderTimeout: 2 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

func NewRequest(url string) string {
	var answer string = ""
	//	t := time.Now().Add(2 * time.Second)
	//	ctx, cancel := context.WithCancel(context.TODO())
	reqm, _ := http.NewRequest("GET", url, nil)
	reqm.Header.Set("User-Agent", "Mozilla")
	reqm.Header.Set("Content-Type", "text/html")
	content, err := client.Do(reqm)
	if err != nil {
		fmt.Println(err)
		if content != nil {
			fmt.Println("statusCode: ", content.StatusCode)
		}
		return answer
	} else if content.StatusCode >= 400 {
		return answer
	}
	defer content.Body.Close()

	value, err := io.ReadAll(content.Body)
	if err != nil {
		fmt.Println(err)
		return answer
	}
	answer = string(value)
	return answer
}

func NewImgRequest(url, filename string) string {
	var answer string = ""
	reqm, _ := http.NewRequest("GET", url, nil)
	reqm.Header.Set("User-Agent", "Mozilla")
	reqm.Header.Set("Content-Type", "image/*")
	content, err := client.Do(reqm)
	if err != nil {
		fmt.Println(err)
		if content != nil {
			fmt.Println("statusCode: ", content.StatusCode)
		}
		return answer
	} else if content.StatusCode >= 400 {
		return answer
	}
	defer content.Body.Close()

	// Create a new file to save the image
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return answer
	}
	defer file.Close()

	// Copy the image data from the response body to the file
	_, err = io.Copy(file, content.Body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return answer
	}
	answer = "OK"
	fmt.Println("Image downloaded successfully.")
	return answer
}
