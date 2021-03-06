/**
 * @author 刘荣飞 yes@noxue.com
 * @date 2018/12/31 12:20
 */
package srv

import (
	"fmt"
	"github.com/pkg/errors"
	"noxue/config"
	"noxue/utils"
)

// 发送验证码
var ApiCode CodeApi

type CodeApi struct {
}

func (CodeApi) SendReg(number, numberType string) (key string, err error) {
	key, code := utils.GenerateVerifyCode(number)
	if config.Config.Debug {
		fmt.Println("发送的验证码是：", code)
	}

	if numberType == "phone" {
		err = utils.SendRegCodePhone(number, code)
	} else if numberType == "email" {
		err = utils.SendRegCodeEmail(number, code)
	} else {
		err = errors.New("目前只支持邮箱或手机注册，请确认账号格式是否正确")
	}

	return
}
