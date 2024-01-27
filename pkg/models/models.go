package models

type IPrecord struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Size int    `json:"size"`
	Src  string `json:"src"`
	Dsc  string `json:"description"`
}

type FullData struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Detail string `json:"detailText"`
	Docs   []Docs `json:"docs"`
}

type Docs struct {
	ID     int        `json:"id"`
	Name   string     `json:"name"`
	Recods []IPrecord `json:"files"`
}
