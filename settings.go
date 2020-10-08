package main

import (
	"encoding/json"
	"io/ioutil"
)

type Settings struct {
	Servers   []Server    `json:"servers"`
	AdminId   int         `json:"adminId"`
	WhiteList []WhiteList `json:"whiteList"`
}

type Server struct {
	Name   string `json:"name"`
	Info   string `json:"info"`
	IsFree bool   `json:"isFree"`
	From   int64  `json:"from"`
	To     int64  `json:"to"`
	ByName string `json:"byName"`
	ById   int    `json:"byId"`
	Desc   string `json:"desc"`
}

type WhiteList struct {
	Id   int    `json:"id"`
	Desc string `json:"desc"`
}

const settingsFileName = "settings.json"

var settings Settings

//--------------------------------------------

func ReadSettings() error {
	b, err := ioutil.ReadFile(settingsFileName)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &settings)
	if err != nil {
		return err
	}

	return nil
}

func UpdateSettings() error {
	b, err := json.Marshal(settings)
	if err != nil {
		return nil
	}

	err = ioutil.WriteFile(settingsFileName, b, 0644)
	if err != nil {
		return nil
	}

	return err
}

func ReleaseServer(i int) {
	settings.Servers[i].IsFree = true
	settings.Servers[i].From = 0
	settings.Servers[i].To = 0
	settings.Servers[i].ById = 0
	settings.Servers[i].ByName = ""
	settings.Servers[i].Desc = ""
}
