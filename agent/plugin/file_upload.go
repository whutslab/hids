package plugin

import (
	"bytes"
	"encoding/json"
	"go.uber.org/config"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Res struct {
	Status_code string `json:"status_code"`
	Hostid      string `json:"hostid"`
}
type remote struct {
	Baseurl string
}

var remoteserver remote
var Hostid string

func ParserConf() {
	f, err := os.Open("config.yaml")
	if err != nil {
		log.Fatalf("Open yaml error:%s\n", err)
		return
	}
	defer f.Close()
	config, err := config.NewYAML(config.Source(f))
	if err != nil {
		log.Fatalf("Error:%s\n", err)
		return
	}
	err = config.Get("remoteserver").Populate(&remoteserver)
	if err != nil {
		log.Fatalf("Parser error:%s\n", err)
		return
	}
}

func postFile(filename string, target_url string, params map[string]string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("file", filename)
	if err != nil {
		log.Println("Error writing to buffer")
		return err
	}
	fh, err := os.Open(filename)
	if err != nil {
		log.Println("error opening file")
		return err
	}
	defer fh.Close()
	// iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}
	contentType := bodyWriter.FormDataContentType()
	for key, val := range params {
		_ = bodyWriter.WriteField(key, val)
	}
	bodyWriter.Close()

	resp, err := http.Post(target_url, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Printf("Post file to server ,receive %s\n", string(resp_body))
	return nil
}

func Http_postFile(filename string, target_url string, hostID string) {
	if hostID == "" {
		log.Println("hostID is NIL")
		return
	}
	params := map[string]string{
		"hostid": hostID,
	}
	err := postFile(filename, target_url, params)
	if err != nil {
		log.Println("Post file failed with %s", err)
	}
}

func Http_postPcap(filename string, target_url string, hostID string, cpurate float64, memrate float64) {
	if hostID == "" {
		log.Println("hostID is NIL")
		return
	}
	params := map[string]string{
		"hostid": hostID,
		"cpu":    strconv.FormatFloat(cpurate, 'E', -1, 64),
		"mem":    strconv.FormatFloat(memrate, 'E', -1, 64),
	}
	err := postFile(filename, target_url, params)
	if err != nil {
		log.Println("Post file failed with %s", err)
	}
}

func Gethostnamereq(url string) string {
	timeout := time.Duration(10 * time.Second)
	clent := http.Client{
		Timeout: timeout,
	}
	resp, err := clent.Get(url)
	//resp, err := http.Get(url)
	var res Res
	if err != nil {
		log.Println(err)
		return ""
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		err = json.Unmarshal(body, &res)
		if err != nil {
			log.Println("Json parser failed with s%", err.Error())
			return ""
		}
		Hostid = res.Hostid
		log.Printf("Get hostID{%s} successfully!", res.Hostid)
	}
	return res.Hostid
}
