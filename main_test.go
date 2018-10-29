package app

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestDecodeItem(t *testing.T) {
	name := "milk"
	supermarket := "fakta"
	price := float32(10)
	s := fmt.Sprintf(`
	{
		"name" : %s,
		"supermarket" : %s,
		"price": %f
	}
	`, name, supermarket, price)
	rc := ioutil.NopCloser(strings.NewReader(s))
	item, err := decodeItem(rc)
	if err != nil {
		t.Errorf(err.Error())
	}
	if item.Name != name {
		t.Errorf("Was %s, expected %s", item.Name, name)
	}
	if item.Supermarket != supermarket {
		t.Errorf("Was %s, expected %s", item.Supermarket, supermarket)
	}
	if item.Price != price {
		t.Errorf("Was %f, expected %f", item.Price, price)
	}
}
