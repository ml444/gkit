package account

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ml444/gkit/transport/httpx"
)

func TestNewAccountHTTPClient(t *testing.T) {
	svr := httpx.NewServer(httpx.Timeout(10 * time.Second))
	go func() {
		ctx := context.Background()
		defer svr.Stop(ctx)
		if err := svr.Start(ctx); err != nil {
			panic(err)
		}
	}()
	time.Sleep(2 * time.Second)
	RegisterAccountHTTPServer(svr, NewAccountService())
	client, err := httpx.NewClient(httpx.WithEndpoint("127.0.0.1:5050"), httpx.WithTimeout(20*time.Second))
	if err != nil {
		t.Errorf("err: %v", err)
		return
	}
	//t.Log(json.Name)
	e, err := svr.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	//reqURL := e.String() + "/account/GetAccountInfo/cml"
	//reqURL := e.String() + "/account/ListAccount?account=cml&name=123"
	reqURL := e.String() + "/account/Register"

	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer([]byte(`{"account":"cml", "password":"123"}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header = http.Header{
		//"Content-Type": {"application/x-www-form-urlencoded; param=value"},
		"Content-Type": {"application/json; param=value"},
		"Accept":       {"application/json; charset=utf-8"}}
	resp, err := client.Do(req)
	if err != nil {
		t.Log(err.Error())
	}
	t.Log(resp)
}
