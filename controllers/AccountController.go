package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"wahaha/utils"
	"wahaha/module/rbac"
	"regexp"
	"wahaha/conf"
	"strings"
	"wahaha/base"
	"wahaha/service/impl"
	"wahaha/utils/jwt"
	"wahaha/utils/gocaptcha"
	"wahaha/constant/httpcode"
)

//注册页面
func RegisteredHtml(context *gin.Context) {
	baseName, _ := utils.AppConfigMap[utils.BASE_NAME]
	htmlTitle, _ := utils.AppConfigMap[utils.HTML_TITLE]
	baseUrl, _ := utils.AppConfigMap[utils.BASE_URL]

	context.HTML(http.StatusOK, "register.html", gin.H{
		"baseName":            baseName,
		"registeredHTMLTitle": htmlTitle,
		"baseUrl":             baseUrl,
		"captcha":"http://"+conf.GetEnv().ServerConfig.SERVER_IP+":"+conf.GetEnv().ServerConfig.SERVER_PORT+"/login/captcha",
	})

}

//注册
func Register(context *gin.Context) {
	var m rbac.Member
	if err := context.BindJSON(&m); err != nil {
		baseResult := base.BaseReturnJson{
			Code:    httpcode.BASE_SYS_ERROR_CODE,
			Message: httpcode.MemberHttpCodes[httpcode.BASE_SYS_ERROR_CODE],
		}
		context.JSON(http.StatusOK, baseResult)
		return
	}
	var code string
	if code = context.PostForm("code"); code == "" {
		baseResult := base.BaseReturnJson{
			Code:    httpcode.CODE_IS_EMPTY,
			Message: httpcode.MemberHttpCodes[httpcode.CODE_IS_EMPTY],
		}
		context.JSON(http.StatusOK, baseResult)
		return
	}
	errMsg, ok := checkMember(&m)
	if value, exists := context.Get(conf.CaptchaSessionName);!exists ||  value != code {
		baseResult := base.BaseReturnJson{
			Code:    httpcode.CODE_IS_NOT_EQUAL,
			Message: httpcode.MemberHttpCodes[httpcode.CODE_IS_NOT_EQUAL],
		}
		context.JSON(http.StatusOK, baseResult)
		return
	}
	if !ok {
		b := base.ReturnCode(http.StatusOK, errMsg, nil)
		context.JSON(http.StatusOK, b)
	} else {
		member := impl.Member{}
		e := member.AddMember(&m)
		context.JSON(http.StatusOK, e)
	}
}

func checkMember(m *rbac.Member) (errMsg string, flg bool) {
	var confirmPassword string
	if ok, err := regexp.MatchString(conf.RegexpAccount, m.Account); !ok || err != nil {
		errMsg = "账号只能由英文字母数字组成，且在3-50个字符"
		return
	}
	if l := strings.Count(m.RealName, ""); l > 50 {
		errMsg = "读者姓名不能为空"
		return
	}

	if l := strings.Count(m.Password, ""); l > 50 || l < 6 {
		errMsg = "密码必须在6-50个字符之间"
		return
	}
	if confirmPassword != m.Password {
		errMsg = "两次密码不一致"
		return
	}
	if ok, err := regexp.MatchString(conf.RegexpEmail, m.Email); !ok || err != nil || m.Email == "" {
		errMsg = "邮箱格式不正确"
		return
	}

	flg = true
	return
}

func Login(context *gin.Context) {
	var m rbac.Member
	if err := context.BindJSON(&m); err != nil {
		panic(err)
		return
	}
	if errMsg, flg := CheckLoginParams(&m); !flg {
		b := base.ReturnCode(http.StatusOK, errMsg, nil)
		context.JSON(http.StatusOK, b)
		return
	}
	member := impl.Member{}

	if e := member.Login(&m); !e.ExecuteStatus {
		context.JSON(http.StatusOK, e)
		return
	}
	//放到session
	if token, err := jwt.GenerateToken(m.Account, m.Password); err != nil {
		panic(err)
	} else {
		slice := strings.Split(context.Request.URL.String(), "redirect_url")
		var redictUrl string
		if len(slice) > 1 {
			redictUrl = slice[1]
		} else {
			redictUrl = "minidoc"
		}
		context.JSON(http.StatusOK, gin.H{
			"code":       http.StatusOK,
			"data":       token,
			"redict_url": redictUrl,
		})
	}
}

func CheckLoginParams(m *rbac.Member) (errMsg string, flg bool) {
	if m.Account == "" {
		errMsg = "账号不能为空!"
		return
	}
	if m.Password == "" {
		errMsg = "密码不能为空!"
		return
	}
	flg = true
	return
}

// 验证码
func Captcha(context *gin.Context) {

	captchaImage, err := gocaptcha.NewCaptchaImage(140, 40, gocaptcha.RandLightColor())

	if err != nil {
		panic(err)
	}

	captchaImage.DrawNoise(gocaptcha.CaptchaComplexLower)

	// captchaImage.DrawTextNoise(gocaptcha.CaptchaComplexHigh)
	txt := gocaptcha.RandText(4)

	context.Set(conf.CaptchaSessionName, txt)
	//c.SetSession(conf.CaptchaSessionName, txt)

	captchaImage.DrawText(txt)
	// captchaImage.Drawline(3);
	captchaImage.DrawBorder(gocaptcha.ColorToRGB(0x17A7A7A))
	// captchaImage.DrawHollowLine()

	captchaImage.SaveImage(context.Writer, gocaptcha.ImageFormatJpeg)

}
