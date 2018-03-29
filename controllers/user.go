package controllers

import(
	"github.com/astaxie/beego"
	_ "go/types"
	"ihome/utils"
	"encoding/json"
	"github.com/astaxie/beego/orm"
	"ihome/models"
	"os"
	"github.com/astaxie/beego/httplib"
)

type UserController struct {
	beego.Controller
}

type RegInfo struct {
	Mobile string `json:"mobile"`
	Password string `json:"password"`
	Sms_code string `json:"sms_code"`
}

func (this *UserController) RetData(resp interface{}) {
	beego.Info("UserController....RetData is called")
	this.Data["json"] = resp
	this.ServeJSON()//回给浏览器
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
	err := json.Unmarshal(this.Ctx.Input.RequestBody,&reginfo)
	if err != nil {
		beego.Info("Unmarshal request body err", err)
		return
	}
	beego.Info(reginfo)
	//数据校验
	if reginfo.Mobile == "" || reginfo.Password == ""||reginfo.Sms_code == "" {
		beego.Info("request body data err")
		return
	}
	//插入到数据库
	o := orm.NewOrm()
	r := o.Raw("insert user(name,password_hash,mobile) values(?,?,?)",reginfo.Mobile,reginfo.Password,reginfo.Mobile)
	res,err := r.Exec()
	if err != nil {
		beego.Info("insert user err", err)
		resp.Errno = utils.RECODE_DBERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	userid,_ := res.LastInsertId()
	beego.Info("userid is ....", userid)
	//设置session
	this.SetSession("name",reginfo.Mobile)
	this.SetSession("user_id", userid)
	this.SetSession("mobile",reginfo.Mobile)
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
	err := json.Unmarshal(this.Ctx.Input.RequestBody,&mapRequest)
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
	r := o.Raw("select * from user where mobile = ? and password_hash = ?",mapRequest["mobile"],mapRequest["password_hash"])
	var user models.User
	err = r.QueryRow(&user)
	if err != nil {
		beego.Info("QueryRow user err", err)
		resp.Errno = utils.RECODE_DBERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	beego.Info("get user....",user)
	this.SetSession("name",user.Name)
	this.SetSession("user_id",user.Id)
	this.SetSession("mobile",user.Mobile)
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
	err := json.Unmarshal(this.Ctx.Input.RequestBody,&mapRequest)
	if err != nil {
		beego.Info("UpdateUserName Unmarshal err", err)
		resp.Errno = utils.RECODE_PARAMERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	beego.Info("get name ...",mapRequest["name"],"user_id===", userid)
	//更新数据库
	o := orm.NewOrm()
	r := o.Raw("update user set name= ? where id = ?",mapRequest["name"], userid)
	_,err = r.Exec()
	if err != nil {
		beego.Info("update user err", err)
		resp.Errno = utils.RECODE_DBERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	//更新session
	this.SetSession("name",mapRequest["name"])
}

//添加头像,获取上传文件
func (this *UserController) UploadUserPic() {
	beego.Info("UploadUserPic is called")
	var resp NormalResp
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(resp.Errno)
	defer this.RetData(&resp)
	//开始编写业务逻辑
	f,h,err := this.GetFile("avatar")
	if err != nil {
		beego.Info("getfile err", err)
		resp.Errno = utils.RECODE_REQERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	defer f.Close()
	beego.Info("get filename ===",h.Filename)
	this.SaveToFile("avatar",h.Filename)
	defer os.Remove(h.Filename)

	//利用go语言模拟客户端
	req := httplib.Post("http://up.imgapi.com")
	//伪装成浏览器
	req.Header("Accept-Encoding", "gzip,deflate,sdch")
	req.Header("Host", "up.imgapi.com")
	req.Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/31.0.1650.57 Safari/537.36")
	//设置token
	req.Param("Token", "c8e56d278e8bf78f6e203b4619bb153a3f07a98d:kRfdE5dNXHNC-c933rf4Y1xZ8VM=:eyJkZWFkbGluZSI6MTUyMjI4ODMwMiwiYWN0aW9uIjoiZ2V0IiwidWlkIjoiNjM1NzM2IiwiYWlkIjoiMTQyMzkxMiIsImZyb20iOiJmaWxlIn0=")
	hr := req.PostFile("file",h.Filename)//上传文件到服务器
	hrdata,err := hr.Bytes()
	if err != nil {
		beego.Info("hr.Bytes err", err)
		resp.Errno = utils.RECODE_REQERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	mapResp := make(map[string]interface{})
	err = json.Unmarshal(hrdata,&mapResp)
	if err != nil {
		beego.Info("Unmarshal err", err)
		resp.Errno = utils.RECODE_REQERR
		resp.Errmsg = utils.RecodeText(resp.Errno)
		return
	}
	beego.Info("get resp map ==========",mapResp)
	linkurl := mapResp["linkurl"]
	//更新数据库
	o := orm.NewOrm()
	userid := this.GetSession("user_id")
	r := o.Raw("update user set avatar_url= ? where id = ?",linkurl,userid)
	if _,err = r.Exec();err != nil {
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
