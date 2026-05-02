package main

type Tournament struct {
	Id     int    `json:"id"`
	Date   string `json:"date"`
	Name   string `json:"name"`
	Loaded bool   `json:"loaded"`
}
