package json

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

type customJSON struct {
	value string
}

func (c customJSON) MarshalJSON() ([]byte, error) {
	return []byte(`"` + c.value + `"`), nil
}

func (c *customJSON) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	c.value = s
	return nil
}

func TestCoderJSONBranches(t *testing.T) {
	c := GetCoder()
	if c.Name() != Name {
		t.Fatalf("name = %q", c.Name())
	}

	data, err := c.Marshal(customJSON{value: "custom"})
	if err != nil || string(data) != `"custom"` {
		t.Fatalf("custom marshal = %q, %v", data, err)
	}
	var custom customJSON
	if err := c.Unmarshal(data, &custom); err != nil || custom.value != "custom" {
		t.Fatalf("custom unmarshal = %#v, %v", custom, err)
	}

	pm := wrapperspb.String("hello")
	data, err = c.Marshal(pm)
	if err != nil {
		t.Fatalf("proto marshal: %v", err)
	}
	var out wrapperspb.StringValue
	if err := c.Unmarshal(data, &out); err != nil || out.Value != "hello" {
		t.Fatalf("proto unmarshal = %q, %v", out.Value, err)
	}

	var ptr *wrapperspb.StringValue
	if err := c.Unmarshal(data, &ptr); err != nil || ptr == nil || ptr.Value != "hello" {
		t.Fatalf("pointer proto unmarshal = %#v, %v", ptr, err)
	}

	var plain struct {
		A string `json:"a"`
	}
	if err := c.Unmarshal([]byte(`{"a":"b"}`), &plain); err != nil || plain.A != "b" {
		t.Fatalf("plain unmarshal = %#v, %v", plain, err)
	}
}
