package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/hooksie1/piggybank/server"
)

type PUser struct {
	*server.User
}

type PApp struct {
	*server.Application
}

type Printer interface {
	String() string
}

func (u *PUser) String() string {
	return fmt.Sprintf("User Created\nUsername: %s\nPassword: %s", u.Username, u.Pass.PlainText)
}

func (p *PApp) String() string {
	return fmt.Sprintf("Application: %s\nUsername: %s\nPassword: %s", p.Application.Application, p.Username, p.Password)
}

func PrintData(p Printer, r io.ReadCloser) error {
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
