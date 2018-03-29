package controllers

import (
	"beeHome/models"
	"beeHome/utils"
	"encoding/json"
	_ "fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/orm"
	_ "go/types"
	"os"
	_ "os"
	_ "runtime/debug"
	"strconv"
	_ "strings"
	"time"
)

type UserController struct {
	beego.Controller
}

type RegInfo struct {
	Mobile   string `json:"mobile"`
	Password string `json:"password"`
	Sms_code string `json:"sms_code"`
}

func (this *UserController) RetData(resp interface{}) {
	beego.Info("UserController....RetData is called")
	this.Data["json"] = resp
	this.ServeJSON() //回给浏览器
}

//用户注册 --> /api/v1.0/users - port
func (this *UserController) UserReg() {
	beego.Info("UserReg() is called")
	var resp NormalResp
	resp.Errno = utils.RECODE_DATAERR
	resp.Errmsg = utils.RecodeText(resp.Errno)
	defer this.RetData(&resp)

	//获得注册信息 从请求里得到
	var reginfo RegInfo
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &reginfo)
	if err != nil {
		beego.Info("Unmarshal request body err", err)
		return
	}
	beego.Info(reginfo)
	//数据校验
	if reginfo.Mobile == "" || reginfo.Password == "" || reginfo.Sms_code == "" {
		beego.Info("request body data err")
		return
	}
	//插入到数据库
	o := orm.NewOrm()
	r := o.Raw("insert user(name,password_hash,mobile) values(?,?,?)", reginfo.Mobile, reginfo.Password, reginfo.Mobile)
	res, err := r.Exec()
	if err != nil {
		beego.Info("insert user err", err)
		resp.Errno = utils.RECODE_DBERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	userid, _ := res.LastInsertId()
	beego.Info("userid is ....", userid)
	//设置session
	this.SetSession("name", reginfo.Mobile)
	this.SetSession("user_id", userid)
	this.SetSession("mobile", reginfo.Mobile)
	//重新设置响应码
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(resp.Errno)
}

//处理用户登陆
func (this *UserController) UserLogin() {
	beego.Info("user loign is called")
	var resp NormalResp
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(resp.Errno)
	defer this.RetData(&resp)
	//获取登陆请求信息 用户手机号和密码
	mapRequest := make(map[string]interface{})
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &mapRequest)
	if err != nil {
		beego.Info("UserLogin Unmarshal err", err)
		resp.Errno = utils.RECODE_PARAMERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	beego.Info("get request map ", mapRequest)

	//数据校验
	if mapRequest["mobile"] == nil || mapRequest["password"] == nil {
		beego.Info("data err", err)
		resp.Errno = utils.RECODE_PARAMERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}

	//查询数据库看是否由结果 也就是验证用户名和密码是否ok
	o := orm.NewOrm()
	r := o.Raw("select * from user where mobile = ? and password_hash = ?", mapRequest["mobile"], mapRequest["password"])
	var user models.User
	err = r.QueryRow(&user)
	if err != nil {
		beego.Info("QueryRow user err", err)
		beego.Info(mapRequest["mobile"], mapRequest["password"])
		resp.Errno = utils.RECODE_DBERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	beego.Info("get user....", user)
	this.SetSession("name", user.Name)
	this.SetSession("user_id", user.Id)
	this.SetSession("mobile", user.Mobile)
}

//更新用户名
func (this *UserController) UpdateUserName() {
	beego.Info("UpdateUserName is called")
	var resp NormalResp
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(resp.Errno)
	defer this.RetData(&resp)
	//从session中获得user_id
	userid := this.GetSession("user_id")
	//获取用户名
	//获取登陆请求信息 用户手机号和密码
	mapRequest := make(map[string]interface{})
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &mapRequest)
	if err != nil {
		beego.Info("UpdateUserName Unmarshal err", err)
		resp.Errno = utils.RECODE_PARAMERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	beego.Info("get name ...", mapRequest["name"], "user_id===", userid)
	//更新数据库
	o := orm.NewOrm()
	r := o.Raw("update user set name= ? where id = ?", mapRequest["name"], userid)
	_, err = r.Exec()
	if err != nil {
		beego.Info("update user err", err)
		resp.Errno = utils.RECODE_DBERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	//更新session
	this.SetSession("name", mapRequest["name"])
}

//添加头像,获取上传文件
func (this *UserController) UploadUserPic() {
	beego.Info("UploadUserPic is called")
	var resp NormalResp
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(resp.Errno)
	defer this.RetData(&resp)
	//开始编写业务逻辑
	f, h, err := this.GetFile("avatar")
	if err != nil {
		beego.Info("getfile err", err)
		resp.Errno = utils.RECODE_REQERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	defer f.Close()
	beego.Info("get filename ===", h.Filename)
	picname := strconv.Itoa(int(time.Now().UnixNano())) + ".jpg"
	this.SaveToFile("avatar", picname)
	beego.Info("picname=", picname)
	defer os.Remove(picname)

	//利用go语言模拟客户端
	req := httplib.Post("http://up.imgapi.com")
	//伪装成浏览器
	req.Header("Accept-Encoding", "gzip,deflate,sdch")
	req.Header("Host", "up.imgapi.com")
	req.Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/31.0.1650.57 Safari/537.36")
	//设置token
	//59acb1be209c3e94204f884fac2e7686afa67764:J7i81cGB-dZj9a9wJGlIqBO1B0o=:eyJkZWFkbGluZSI6MTUyMjMyMzM0MiwiYWN0aW9uIjoiZ2V0IiwidWlkIjoiNjM4ODU4IiwiYWlkIjoiMTQyNDI5MSIsImZyb20iOiJ3ZWIifQ==
	//59acb1be209c3e94204f884fac2e7686afa67764:Ck5kCcVOEwROjbqkJH_GD7d13gM=:eyJkZWFkbGluZSI6MTUyMjMyNjYzNSwiYWN0aW9uIjoiZ2V0IiwidWlkIjoiNjM4ODU4IiwiYWlkIjoiMTQyNDMwNSIsImZyb20iOiJmaWxlIn0=
	req.Param("Token", "59acb1be209c3e94204f884fac2e7686afa67764:Ck5kCcVOEwROjbqkJH_GD7d13gM=:eyJkZWFkbGluZSI6MTUyMjMyNjYzNSwiYWN0aW9uIjoiZ2V0IiwidWlkIjoiNjM4ODU4IiwiYWlkIjoiMTQyNDMwNSIsImZyb20iOiJmaWxlIn0=")
	hr := req.PostFile("file", picname) //上传文件到服务器
	hrdata, err := hr.Bytes()
	if err != nil {
		beego.Info("hr.Bytes err", err)
		resp.Errno = utils.RECODE_REQERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	mapResp := make(map[string]interface{})
	err = json.Unmarshal(hrdata, &mapResp)
	if err != nil {
		beego.Info("Unmarshal err", err)
		resp.Errno = utils.RECODE_REQERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	beego.Info("get resp map ==========", mapResp)
	linkurl := mapResp["linkurl"]
	//更新数据库
	o := orm.NewOrm()
	userid := this.GetSession("user_id")
	beego.Info(linkurl, userid)
	r := o.Raw("update user set avatar_url= ? where id = ?", linkurl, userid)
	if _, err = r.Exec(); err != nil {
		beego.Info("Unmarshal err", err)
		resp.Errno = utils.RECODE_DBERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	//设置返回
	type urlinfo struct {
		Avatar_url string `json:"avartar_url"`
	}
	var info urlinfo
	info.Avatar_url = linkurl.(string)
	resp.Data = &info
}

//请求用户信息
func (this *UserController) GetUserInfo() {
	beego.Info("GetUserInfo is called")
	resp := NormalResp{Errno: utils.RECODE_OK, Errmsg: utils.RecodeText(utils.RECODE_OK)}
	defer this.RetData(&resp)
	//获得user_id 通过session获得
	userid := this.GetSession("user_id")
	//查询数据库
	o := orm.NewOrm()
	r := o.Raw("select * from user where id = ?", userid)
	var user models.User
	if err := r.QueryRow(&user); err != nil {
		beego.Info("query user err ", err)
		resp.Errno = utils.RECODE_DBERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	mapUserInfo := map[string]interface{}{"user_id": user.Id,
		"name": user.Name, "password": user.Password_hash,
		"mobile": user.Mobile, "real_name": user.Real_name,
		"id_card": user.Id_card, "avatar_url": user.Avatar_url}
	resp.Data = mapUserInfo
}

//更新实名认证
func (this *UserController) UpdateUserAuth() {
	beego.Info("UpdateUserAuth is called")
	var resp NormalResp
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(resp.Errno)
	defer this.RetData(&resp)
	//开始业务逻辑
	//1,获得验证数据
	mapRequest := make(map[string]interface{})

	err := json.Unmarshal(this.Ctx.Input.RequestBody, &mapRequest)
	if err != nil {
		beego.Info("UpdateUserAuth Unmarshal err", err)
		resp.Errno = utils.RECODE_PARAMERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	if mapRequest["real_name"] == "" || mapRequest["id_card"] == "" {
		beego.Info("UpdateUserAuth request data err")
		resp.Errno = utils.RECODE_PARAMERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	beego.Info(mapRequest)
	//2,更新数据库
	userid := this.GetSession("user_id")
	o := orm.NewOrm()
	r := o.Raw("update user set id_card = ?,real_name = ? where id = ?", mapRequest["id_card"],
		mapRequest["real_name"], userid)
	if _, err = r.Exec(); err != nil {
		beego.Info("update user err", err)
		resp.Errno = utils.RECODE_USERERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	//设置session
	this.SetSession("user_id", userid)
}
