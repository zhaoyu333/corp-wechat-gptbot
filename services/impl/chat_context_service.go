package impl

import (
	xcache "github.com/baiyz0825/corp-webot/cache"
	"github.com/baiyz0825/corp-webot/dao"
	"github.com/baiyz0825/corp-webot/model"
	"github.com/baiyz0825/corp-webot/to"
	"github.com/baiyz0825/corp-webot/utils/xlog"
	"github.com/baiyz0825/corp-webot/xconst"
	"github.com/sirupsen/logrus"
)

type ContextCommand struct {
	Command string
}

func NewContextCommand() *ContextCommand {
	return &ContextCommand{Command: xconst.COMMAN_GPT_DELETE_CONTEXT}
}

// Exec
// @Description: 删除上下文信息
// @receiver c
// @param userData
// @return bool
func (c ContextCommand) Exec(userData to.MsgContent) bool {
	// 删除缓存上下文
	var msgContext model.MessageContext
	cache := xcache.GetDataFromCache(xcache.GetUserCacheKey(userData.FromUsername))
	if cache != nil {
		context, ok := cache.(model.MessageContext)
		if !ok {
			logrus.WithField("error", "上下文断言失败").
				WithField("userID", userData.FromUsername).
				Errorf("用户上下文数据断言失败")
			return false
		}
		msgContext = context
	}
	// 存储之前的context到db
	msgContextJson, err := MarshalMsgContextToJSon(userData, msgContext)
	err = dao.InsertUserContext(userData.FromUsername, string(msgContextJson))
	if err != nil {
		xlog.Log.WithError(err).WithField("插入数据是:", string(msgContextJson)).
			WithField("用户是:", userData.FromUsername).
			Error("保存过期缓存中的用户上下文数据->db错误")
		return false
	}
	// 删除缓存
	xcache.GetCacheDb().Delete(msgContext.Key)
	// 删除用户设置的prompt
	err = dao.UpdateUser("", userData.FromUsername)
	if err != nil {
		xlog.Log.WithError(err).WithField("用户:", userData.FromUsername).Error("删除用户此次设置prompt失败")
	}
	SendToWxByText(userData, xconst.AI_CLEAR_CONTEXT_SUCCESS)
	return true
}

