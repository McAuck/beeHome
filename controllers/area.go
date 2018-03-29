package controllers

import (
	"github.com/astaxie/beego"
	"beeHome/utils"
	"beeHome/models"
	"github.com/astaxie/beego/orm"
)

type AreaController struct {
	 beego.Controller
}

func (a *AreaController) RetData(resp interface{}) {
	beego.Info("AreaController .... RetData is called")
	a.Data["json"] = resp
	a.ServeJSON()
}

//获取营业区
func (a *AreaController) GetAreas() {
	beego.Info("GetAreas() is called")
	var resp NormalResp
	resp.Errno = "0"
	resp.Errmsg = "OK"
	defer a.RetData(&resp)
	//查询数据库的数据
	o := orm.NewOrm()
	r := o.Raw("select * from area")
	var areas []models.Area
	num,err := r.QueryRows(&areas)
	if err != nil || num <= 0 {
		beego.Info("query data err or not data found",err)
		resp.Errno = utils.RECODE_DBERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	beego.Info("num====", num)
	beego.Info(areas)
	resp.Data = &areas
}