/**
 * @author 刘荣飞 yes@noxue.com
 * @date 2018/12/26 23:55
 */

package srv

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/noxue/ormgo.v1"
	"noxue/dao"
	"noxue/model"
	"noxue/utils"
)

var SrvUser UserService

type UserService struct {
}

func init() {
	initGroup()
}

func initGroup() {
	dao.UserDao.GroupInsert("普通用户", "")
	dao.UserDao.GroupInsert("实习版主", "")
	dao.UserDao.GroupInsert("版主", "")
	dao.UserDao.GroupInsert("管理员", "")
	dao.UserDao.GroupInsert("站长", "")
}

func (UserService) GroupExists(name string) (isExists bool, err error) {
	var n int
	n, err = dao.UserDao.GroupCount(map[string]interface{}{"name": name})
	if err != nil {
		return
	}
	isExists = n > 0
	return
}

func (UserService) GroupAdd(group model.UserGroup) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(utils.Error)
		}
	}()
	n, err := dao.UserDao.GroupCount(ormgo.M{"name": group.Name})
	utils.CheckErr(err)
	if n > 0 {
		utils.CheckErr(errors.New("该用户组已存在"))
	}
	err = dao.UserDao.GroupInsert(group.Name, group.Icon)
	return
}

func (UserService) GroupFindById(id string) (group model.UserGroup, err error) {
	group, err = dao.UserDao.GroupFindById(id)
	return
}

func (UserService) GroupFind(name string) (group model.UserGroup, err error) {
	group, err = dao.UserDao.GroupFindByName(name)
	return
}

func (UserService) GroupSelect(condition map[string]interface{}, fields map[string]bool, sorts []string) (groups []model.UserGroup, err error) {
	groups, err = dao.UserDao.GroupSelect(condition, fields, sorts, 0, 0)
	return
}

// 获取指定api能被哪些用户组访问
func (this *UserService) GroupSelectByApi(api string) (groups []model.UserGroup, total int, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(utils.Error)
		}
	}()

	// 获取所有允许访问此api的groupId
	rs, err := this.ResourceSelectByApi(api)
	utils.CheckErr(err)
	var ids []bson.ObjectId
	for _, v := range rs {
		ids = append(ids, v.Group)
	}

	// 根据groupId数组查询出满足条件的group文档
	dao.UserDao.GroupSelect(ormgo.M{
		"_id": ormgo.M{"$in": ids},
	}, nil, nil, 0, 0)
	return
}

func (UserService) GroupRemoveById(id string) (err error) {
	var n int
	n, err = dao.UserDao.UserCount(ormgo.M{
		"groups": ormgo.M{
			"$in": bson.ObjectIdHex(id),
		},
	}, ormgo.All)
	if err != nil {
		return
	}

	if n > 0 {
		err = errors.New("该用户组下存在用户，无法删除")
		return
	}

	err = dao.UserDao.GroupRemove(id)
	return
}

// ====================================================================================================

// 注册用户，根据用户信息和授权信息添加用户
func (this *UserService) UserRegister(user *model.User, auth *model.Auth) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(utils.Error)
		}
	}()

	// 检查用户是否存在
	exists, err := this.UserExists(user.Name)
	utils.CheckErr(err)
	if exists {
		utils.CheckErr(errors.New("用户名[" + user.Name + "]已被占用，请更换一个"))
	}

	// 检查授权信息是否存在
	exists, err = this.AuthExists(auth)
	utils.CheckErr(err)
	if exists {
		utils.CheckErr(errors.New("账号[" + auth.Name + "]已注册，请直接登陆"))
	}

	// 创建一个id，后面添加授权信息需要用到
	user.Id = bson.NewObjectId()
	// 添加用户信息
	err = dao.UserDao.UserInsert(user)
	utils.CheckErr(err)

	// 添加授权信息
	auth.User = user.Id
	err = dao.UserDao.AuthInsert(auth)
	if err != nil {
		// 如果添加授权失败，删除上面添加的用户信息
		// 防止用户名被占用缺无法登陆
		dao.UserDao.UserRemoveById(user.Id.Hex(), true)
	}

	return
}

// 根据授权信息登陆，登陆成功，返回用户信息和授权信息
func (this *UserService) UserLogin(auth *model.Auth) (user model.User, authRet model.Auth, err error) {

	defer func() {
		if e := recover(); e != nil {
			err = e.(utils.Error)
		}
	}()

	// 查询授权信息
	authRet, err = this.AuthCheck(auth)
	if err != nil {
		return
	}

	user, err = dao.UserDao.UserFindById(authRet.User.Hex())
	return
}

// 获取用户拥有的用户组列表
func (UserService) UserGetGroups(uid string) (groups []model.UserGroup, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(utils.Error)
		}
	}()

	u, err := dao.UserDao.UserFindById(uid)
	utils.CheckErr(err)

	groups, err = dao.UserDao.GroupSelect(ormgo.M{
		"_id": ormgo.M{
			"$in": u.Groups,
		},
	}, nil, nil, 0, 0)

	return
}

// 检查用户名是否存在
func (UserService) UserExists(name string) (isExists bool, err error) {
	var n int
	n, err = dao.UserDao.UserCount(ormgo.M{"name": name}, ormgo.All)
	if err != nil {
		return
	}
	isExists = n > 0
	return
}

// 根据名称查找用户
func (UserService) UserFindByName(name string) (user model.User, err error) {
	user, err = dao.UserDao.UserFind(ormgo.M{
		"name": name,
	})
	return
}
func (UserService) UserFindById(id string) (user model.User, err error) {
	user, err = dao.UserDao.UserFindById(id)
	return
}

// 编辑用户资料
// 会用整个user对象替代数据库中的数据
// 警告：如果只赋值了部分字段，其他值将丢失
// 如果要修改部分字段，请使用 UserUpdateFieldsById
func (UserService) UserUpdateById(id string, user model.User) (err error) {
	err = dao.UserDao.UserEditById(id, user)
	return
}

// 编辑用户部分资料
func (UserService) UserUpdateFieldsById(id string, fields map[string]interface{}) (err error) {
	err = dao.UserDao.UserEditById(id, fields)
	return
}

// 把用户添加到指定的用户组中
func (UserService) UserAddToGroups(uid string, groupIds []string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(utils.Error)
		}
	}()

	var ids []bson.ObjectId
	for _, v := range groupIds {
		ids = append(ids, bson.ObjectIdHex(v))
	}

	err = dao.UserDao.UserEditById(uid, ormgo.M{
		"$addToSet": ormgo.M{
			"groups": ormgo.M{
				"$each": ids,
			},
		},
	})

	utils.CheckErr(err)

	return
}

func (UserService) UserRemoveFromGroup(uid, groupId string) (err error) {
	err = dao.UserDao.UserEditById(uid, ormgo.M{
		"$pull": ormgo.M{
			"groups": bson.ObjectIdHex(groupId),
		},
	})
	return
}

//==============================================================================================================

// 判断授权信息是否存在
// auth.Type = 0 则表示不限制查询范围，一般不限制，除非确定要查找什么类型
func (UserService) AuthExists(auth *model.Auth) (isExists bool, err error) {
	var n int
	n, err = dao.UserDao.AuthCount(ormgo.M{
		"type": auth.Type,
		"name": auth.Name,
	})

	if err != nil {
		return
	}
	isExists = n > 0
	return
}

// 根据用户ID查询所有授权信息
func (UserService) AuthSelectByUid(uid string) (auths []model.Auth, err error) {
	auths, err = dao.UserDao.AuthSelectByUid(uid)
	return
}

// 添加授权信息
func (this *UserService) AuthAdd(auth *model.Auth) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(utils.Error)
		}
	}()

	// 判断授权信息是否存在
	exists, err := this.AuthExists(auth)
	utils.CheckErr(err)
	if exists {
		utils.CheckErr(errors.New("账号[" + auth.Name + "]已注册，请直接登陆"))
	}

	err = dao.UserDao.AuthInsert(auth)
	return
}

// 根据旧用户名修改用户名或密码
func (UserService) AuthEdit(oldAuth, auth *model.Auth) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(utils.Error)
		}
	}()

	v := ormgo.M{
	}
	// 如果用户名不为空，表示要修改用户名
	if auth.Name != "" {
		v["name"] = auth.Name
	}

	// 如果不是第三方登陆，密码需要加密
	if !auth.Third {
		v["secret"] = utils.EncodePassword(auth.Secret)
	}

	dao.UserDao.AuthUpdate(ormgo.M{
		"type":  oldAuth.Type,
		"name":  oldAuth.Name,
		"third": oldAuth.Third,
	}, v)
	return
}

// 判断账号密码是否正确
func (UserService) AuthCheck(auth *model.Auth) (authRet model.Auth, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(utils.Error)
		}
	}()
	// 查询授权信息
	authRet, err = dao.UserDao.AuthFind(auth.Type, auth.Name, auth.Third)
	if err != nil {
		utils.CheckErr(errors.New("账号不存在"))
	}

	// 不是第三方，就验证密码。验证密码
	if !auth.Third {
		// 密码在model.User.BeforeSave() hook方法中加密
		err = bcrypt.CompareHashAndPassword([]byte(authRet.Secret), []byte(auth.Secret))
		if err != nil {
			utils.CheckErr(errors.New("密码错误"))
		}
	}
	return
}

// 根据用户编号修改所有非第三方登陆的密码，防止出现手机和邮箱登陆密码不一致的问题
func (UserService) AuthChangePassByUid(uid, password string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(utils.Error)
		}
	}()
	err = dao.UserDao.AuthUpdateAll(ormgo.M{
		"user":  bson.ObjectIdHex(uid),
		"third": false,
	}, ormgo.M{
		"secret": utils.EncodePassword(password),
	})
	return
}

// 删除指定用户的所有授权信息，用于删除用户的时候
func (UserService) AuthRemoveByUid(uid string, really bool) (err error) {
	cond := ormgo.M{
		"user": bson.ObjectIdHex(uid),
	}
	err = dao.UserDao.AuthRemoveAll(cond, really)
	return
}

// 根据id删除第三方授权信息，用于解绑第三方账号
func (UserService) AuthRemoveBId(id string, really bool) (err error) {
	err = dao.UserDao.AuthRemoveById(id, really)
	return
}

//==============================================================================================================

// 授权给指定用户组
func (UserService) ResourceAdd(r *model.Resource) (err error) {
	err = dao.UserDao.ResourceInsert(r)
	return
}

// 根据用户组Id获取授权规则
func (UserService) ResourceSelectByGroupId(groupId string) (rs []model.Resource, err error) {
	rs, _, err = dao.UserDao.ResourceSelect(ormgo.M{
		"group": bson.ObjectIdHex(groupId),
	}, nil, nil, 0, 0)
	return
}

// 根据api获取授权规则
func (UserService) ResourceSelectByApi(api string) (rs []model.Resource, err error) {
	rs, _, err = dao.UserDao.ResourceSelect(ormgo.M{
		"api": api,
	}, nil, nil, 0, 0)
	return
}

// 根据用户Id获取拥有的授权规则
func (UserService) ResourceSelectByUserId(uid string) (rs []model.Resource, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(utils.Error)
		}
	}()

	user, err := dao.UserDao.UserFindById(uid)
	utils.CheckErr(err)

	rs, _, err = dao.UserDao.ResourceSelect(ormgo.M{
		"group": ormgo.M{
			"$in": user.Groups,
		},
	}, nil, nil, 0, 0)

	return
}

func (UserService) ResourceUpdateById(id string, r model.Resource) (err error) {
	err = dao.UserDao.ResourceEditById(id, r)
	return
}

func (UserService) ResourceRemoveById(id string) (err error) {
	err = dao.UserDao.ResourceRemoveById(id)
	return
}
