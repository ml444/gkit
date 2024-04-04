package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/xuri/excelize/v2"

	"tests/storage"

	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/transport/httpx"
)

func main() {
	go RunServer()

	time.Sleep(1 * time.Second)

	httpxCli, err := httpx.NewClient(httpx.WithEndpoint(storage.ServiceEndpoint))
	if err != nil {
		log.Error(err.Error())
		return
	}
	cli := storage.NewStorageHTTPClient(httpxCli)
	testUploadV0BySDK(cli)
	testUploadV1BySDK(cli)
	testUploadV2BySDK(cli)
	testUploadV0ByHTTP()
	testUploadV1ByHTTP()
	testUploadV2ByHTTP()

	testDownloadV0ByHTTP()
	testDownloadV1ByHTTP()
	testDownloadV2ByHTTP()
}

func RunServer() {
	srv := httpx.NewServer(httpx.Address(storage.ServiceEndpoint))
	storage.RegisterStorageHTTPServer(srv, &storage.StroageService{})
	ctx := context.Background()
	if err := srv.Start(ctx); err != nil {
		panic(err.Error())
	}
}
func validateUploadRsp(uploadRsp *storage.UploadRsp) {
	filename := "test_upload.txt"
	checkUrl := fmt.Sprintf("http://%s/%s", storage.ServiceEndpoint, filename)
	if uploadRsp.Url != checkUrl {
		panic(fmt.Sprintf("uploadRsp.Url [%s] != %s", uploadRsp.Url, checkUrl))
	}
	if uploadRsp.Size != 18 {
		panic(fmt.Sprintf("uploadRsp.Size [%d] != 18", uploadRsp.Size))
	}
	log.Info("success: ", uploadRsp.Url)
	time.Sleep(1 * time.Second)
	// delete file
	if err := os.Remove(filename); err != nil {
		log.Error(err.Error())
	}
}

var uploadReq = &storage.UploadReq{
	FileInfo: &storage.FileInfo{
		FileName:   "test_upload",
		FileSuffix: ".txt",
	},
	FileData: []byte("1234567890-abcdefg"),
}

func testUploadV0BySDK(cli storage.StorageHTTPClient) {
	uploadRsp, err := cli.UploadV0(context.Background(), uploadReq)
	if err != nil {
		log.Error(err.Error())
		return
	}
	validateUploadRsp(uploadRsp)
}

func testUploadV1BySDK(cli storage.StorageHTTPClient) {
	uploadRsp, err := cli.UploadV1(context.Background(), uploadReq)
	if err != nil {
		log.Error(err.Error())
		return
	}
	validateUploadRsp(uploadRsp)
}

func testUploadV2BySDK(cli storage.StorageHTTPClient) {
	uploadRsp, err := cli.UploadV2(context.Background(), uploadReq)
	if err != nil {
		log.Error(err.Error())
		return
	}
	validateUploadRsp(uploadRsp)
}

func uploadReqAndDo(req *http.Request) {
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Error("Error sending HTTP request:", err)
		return
	}
	defer response.Body.Close()

	// Read response content
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error("Error reading response body:", err)
		return
	}
	if response.StatusCode != 200 {
		log.Error("Response status code:", response.StatusCode)
		log.Error("Response body:", string(responseData))
		return
	}
	var uploadRsp storage.UploadRsp
	if err := json.Unmarshal(responseData, &uploadRsp); err != nil {
		log.Error(err.Error())
		return
	}
	validateUploadRsp(&uploadRsp)
}

func testUploadV0ByHTTP() {
	url := fmt.Sprintf("http://%s/storage/upload/v0", storage.ServiceEndpoint)
	data, err := json.Marshal(uploadReq)
	if err != nil {
		log.Error(err.Error())
		return
	}
	body := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Error("Error creating HTTP request:", err)
		return
	}
	uploadReqAndDo(req)
}

func testUploadV1ByHTTP() {
	url := fmt.Sprintf("http://%s/storage/upload/v1", storage.ServiceEndpoint)
	data := []byte(`1234567890-abcdefg`)

	body := bytes.NewBuffer(data)

	// 创建一个新的 HTTP 请求对象
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Error("Error creating HTTP request:", err)
		return
	}

	// Set request header `File-Name` and `File-Suffix` to specify the file name and suffix
	req.Header.Set("File-Name", "test_upload")
	req.Header.Set("File-Suffix", ".txt")
	//NOTE: set header `Content-Type` to `application/octet-stream`
	req.Header.Set("Content-Type", "application/octet-stream")

	uploadReqAndDo(req)
}

func testUploadV2ByHTTP() {
	url := fmt.Sprintf("http://%s/storage/upload/v2", storage.ServiceEndpoint)
	data := []byte(`1234567890-abcdefg`)

	body := bytes.NewBuffer(data)

	// 创建一个新的 HTTP 请求对象
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Error("Error creating HTTP request:", err)
		return
	}

	// Set request header `File-Name` and `File-Suffix` to specify the file name and suffix
	req.Header.Set("File-Name", "test_upload")
	req.Header.Set("File-Suffix", ".txt")

	uploadReqAndDo(req)
}

func downloadReqAndDo(req *http.Request, validate func([]byte)) {
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Error("Error sending HTTP request:", err)
		return
	}
	defer response.Body.Close()

	// Read response content
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error("Error reading response body:", err)
		return
	}
	if response.StatusCode != 200 {
		log.Error("Response status code:", response.StatusCode)
		log.Error("Response body:", string(responseData))
		return
	}
	validate(responseData)
}
func validateDownloadV0(responseData []byte) {
	var downloadRsp storage.DownloadRsp
	err := json.Unmarshal(responseData, &downloadRsp)
	if err != nil {
		log.Error("Error unmarshalling response data:", err)
		return
	}
	if downloadRsp.Headers["Content-Type"] != "application/vnd.openxmlformats" {
		log.Errorf("Content-Type: '%s' != 'application/vnd.openxmlformats'\n", downloadRsp.Headers["Content-Type"])
		return
	}
	if downloadRsp.Headers["Content-Disposition"] != "attachment; filename=test.txt" {
		log.Errorf("Content-Disposition: '%s' != 'attachment; filename=test.txt'\n", downloadRsp.Headers["Content-Disposition"])
		return
	}
	if downloadRsp.Headers["Access-Control-Expose-Headers"] != "Content-Disposition" {
		log.Errorf("Access-Control-Expose-Headers: '%s' != 'Content-Disposition'\n", downloadRsp.Headers["Access-Control-Expose-Headers"])
		return
	}
	if string(downloadRsp.Data) != "1234567890-abcdefg" {
		log.Errorf("Response data: '%s' != '1234567890-abcdefg'\n", responseData)
		return
	}
	log.Info("success: ", downloadRsp.Headers)

}
func validateDownloadV12(responseData []byte) {
	// save Excel file
	file, err := os.OpenFile("test.xlsx", os.O_CREATE|os.O_WRONLY, 0o776)
	if err != nil {
		log.Error("Error creating file:", err)
		return

	}
	defer file.Close()
	if _, err := file.Write(responseData); err != nil {
		log.Error("Error writing file:", err)
		return
	}

	// open Excel file
	excelFile, err := excelize.OpenFile("test.xlsx")
	if err != nil {
		log.Error("Error opening file:", err)
		return
	}
	// Get value from cell by given worksheet and axis.
	cellValue, err := excelFile.GetCellValue("Sheet1", "A1")
	if err != nil {
		log.Error("Error getting cell value:", err)
		return
	}
	if cellValue != "Hello world." {
		log.Errorf("Cell value: '%s' != 'Hello world.'\n", cellValue)
		return

	}
	if err := os.Remove("test.xlsx"); err != nil {
		log.Error("Error deleting file:", err)
	}
	log.Info("success: file saved and opened successfully")
}

func testDownloadV0ByHTTP() {
	url := fmt.Sprintf("http://%s/storage/download/v0", storage.ServiceEndpoint)

	buf, err := json.Marshal(&storage.DownloadReq{Filename: "test.xlsx"})
	if err != nil {
		log.Error(err.Error())
		return
	}
	// 创建一个新的 HTTP 请求对象
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		log.Error("Error creating HTTP request:", err)
		return
	}
	downloadReqAndDo(req, validateDownloadV0)
}

func testDownloadV1ByHTTP() {
	url := fmt.Sprintf("http://%s/storage/download/v1", storage.ServiceEndpoint)

	buf, err := json.Marshal(&storage.DownloadReq{Filename: "test.xlsx"})
	if err != nil {
		log.Error(err.Error())
		return
	}
	// 创建一个新的 HTTP 请求对象
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		log.Error("Error creating HTTP request:", err)
		return
	}
	downloadReqAndDo(req, validateDownloadV12)
}

func testDownloadV2ByHTTP() {
	url := fmt.Sprintf("http://%s/storage/download/v2", storage.ServiceEndpoint)

	buf, err := json.Marshal(&storage.DownloadReq{Filename: "test.xlsx"})
	if err != nil {
		log.Error(err.Error())
		return
	}
	// 创建一个新的 HTTP 请求对象
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		log.Error("Error creating HTTP request:", err)
		return
	}
	downloadReqAndDo(req, validateDownloadV12)
}

var checkHeader = map[string][]string{
	"File-Path":   {"/test/my_path"}, // not used
	"Expired-In":  {"30"},            // not used
	"File-Name":   {"test_upload"},
	"File-Suffix": {".txt"},
	"Detail":      {`{"name": "foo", "age": "18", "male": "true"}`}, // not used
	"Id-List":     {"123", "456"},                                   // not used
}
