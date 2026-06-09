package stream

import "testing"

func TestCoderStream(t *testing.T) {
	c := GetCoder()
	if c.Name() != Name {
		t.Fatalf("name = %q", c.Name())
	}
	data, err := c.Marshal([]byte("hello"))
	if err != nil || string(data) != "hello" {
		t.Fatalf("marshal = %q, %v", data, err)
	}
	var out []byte
	if err := c.Unmarshal([]byte("world"), &out); err != nil || string(out) != "world" {
		t.Fatalf("unmarshal = %q, %v", out, err)
	}
}
