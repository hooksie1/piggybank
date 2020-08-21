package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/spf13/viper"
)

type Printer interface {
	String() string
}

func PrintData(p Printer, r io.ReadCloser) error {
	jsonTrue := viper.GetBool("jsonTrue")
	if jsonTrue {
		data, err := GetJson(r)
		if err != nil {
			return err
		}

		fmt.Println(string(data))
	}

	if !jsonTrue {
		GetData(r, p)
		fmt.Println(p.String())
	}

	return nil
}

func GetData(r io.ReadCloser, v interface{}) error {
	err := json.NewDecoder(r).Decode(v)
	if err != nil {
		return err
	}

	return nil

}

func GetJson(r io.ReadCloser) (string, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
