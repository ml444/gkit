package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/xuri/excelize/v2"
)

const ServiceEndpoint = "localhost:5050"

type StroageService struct{}

func (s *StroageService) UploadV0(ctx context.Context, req *UploadReq) (*UploadRsp, error) {
	log.Println("=== Upload v0 ===")
	info := req.FileInfo
	if info == nil {
		return nil, errors.New("req.FileInfo is nil")
	}
	relativePath := fmt.Sprintf("%s%s", info.FileName, info.FileSuffix)
	file, err := os.OpenFile(relativePath, os.O_CREATE|os.O_WRONLY, 0o776)
	if err != nil {
		return nil, err
	}
	n, err := file.Write(req.FileData)
	if err != nil {
		return nil, err
	}
	return &UploadRsp{Url: fmt.Sprintf("http://%s/%s", ServiceEndpoint, relativePath), Size: uint32(n)}, nil
}

func (s *StroageService) UploadV1(ctx context.Context, req *UploadReq) (*UploadRsp, error) {
	log.Println("=== Upload v1 ===")
	log.Printf("===>req: %v\n", req)
	return s.UploadV0(ctx, req)
}

func (s *StroageService) UploadV2(ctx context.Context, req *UploadReq) (*UploadRsp, error) {
	log.Println("=== Upload v2 ===")
	log.Printf("===>req: %v\n", req)
	return s.UploadV0(ctx, req)
}

func (s *StroageService) DownloadV0(ctx context.Context, req *DownloadReq) (*DownloadRsp, error) {
	log.Println("=== Download v0 ===")
	return &DownloadRsp{
		Headers: map[string]string{
			"Content-Type":                  "application/vnd.openxmlformats",
			"Content-Disposition":           "attachment; filename=test.txt",
			"Access-Control-Expose-Headers": "Content-Disposition",
		},
		Data: []byte("1234567890-abcdefg"),
	}, nil
}

func (s *StroageService) DownloadV1(ctx context.Context, req *DownloadReq) (*DownloadRsp, error) {
	var rsp DownloadRsp
	if req.Filename == "" {
		return nil, errors.New("req.Filename is empty")
	}
	log.Printf("===> Download v1 Req: %v\n", req)
	buf, err := createTestExcelFile()
	if err != nil {
		return nil, err
	}
	rsp.Data = buf
	rsp.Headers = map[string]string{
		"Content-Type":                  "application/vnd.openxmlformats",
		"Content-Disposition":           fmt.Sprintf("attachment; filename=%s", req.Filename),
		"Access-Control-Expose-Headers": "Content-Disposition",
	}
	return &rsp, nil
}

func (s *StroageService) DownloadV2(ctx context.Context, req *DownloadReq) (*DownloadRsp, error) {
	var rsp DownloadRsp
	if req.Filename == "" {
		return nil, errors.New("req.Filename is empty")
	}
	log.Printf("===> Download v2 Req: %v\n", req)
	buf, err := createTestExcelFile()
	if err != nil {
		return nil, err
	}
	rsp.Data = buf
	// NOTE: Content-Type and Access-Control-Expose-Headers are not set in this version
	rsp.Headers = map[string]string{
		"Content-Disposition": fmt.Sprintf("attachment; filename=%s", req.Filename),
	}
	return &rsp, nil
}

func createTestExcelFile() ([]byte, error) {
	// Create a new Excel file and add a sheet.
	file := excelize.NewFile()
	index, err := file.NewSheet("Sheet1")
	if err != nil {
		return nil, err
	}
	// Set value of a cell.
	err = file.SetCellValue("Sheet1", "A1", "Hello world.")
	if err != nil {
		return nil, err
	}
	// Set active sheet of the workbook.
	file.SetActiveSheet(index)
	// Save xlsx file by the given path.
	//if err := file.SaveAs(req.Filename); err != nil {
	//	return nil, err
	//}
	buf, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
