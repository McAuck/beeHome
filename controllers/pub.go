package controllers

type NormalResp struct {
	Errno string `json:"errno"`
	Errmsg string `json:"errmsg"`
	Data interface{} `json:"data"`
}