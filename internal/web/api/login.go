package api

import (
	"fmt"
	"net/http"
	"strings"

	"go-mahjong-server/db"
	"go-mahjong-server/protocol"

	"go-mahjong-server/db/model"

	"github.com/gorilla/mux"
	"github.com/lonng/nex"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	host     string                // 服务器地址
	port     int                   // 服务器端口
	config   protocol.ClientConfig // 远程配置
	messages []string              // 广播消息
	logger   = log.WithFields(log.Fields{"component": "http", "service": "login"})

	// 游客登陆
	enableGuest   = false
	guestChannels = []string{}

	enableDebug = false
)

const defaultCoin = 10

// 查询是否使用游客登陆
type (
	queryRequest struct {
		AppId     string `json:"appId"`
		ChannelId string `json:"channelId"`
	}

	queryResponse struct {
		Code  int  `json:"code"`
		Guest bool `json:"guest"`
	}
)

var (
	forbidGuest  = &queryResponse{Guest: false}
	accepetGuest = &queryResponse{Guest: true}
)

func MakeLoginService() http.Handler {
	host = viper.GetString("game-server.host")
	port = viper.GetInt("game-server.port")

	// 更新相关配置
	config.Version = viper.GetString("update.version")
	config.Android = viper.GetString("update.android")
	config.IOS = viper.GetString("update.ios")
	config.Heartbeat = viper.GetInt("core.heartbeat")

	// 分享相关配置
	config.Title = viper.GetString("share.title")
	config.Desc = viper.GetString("share.desc")

	// 客服相关配置
	config.Daili1 = viper.GetString("contact.daili1")
	config.Daili2 = viper.GetString("contact.daili2")
	config.Kefu1 = viper.GetString("contact.kefu1")

	// 游客相关配置
	enableGuest = viper.GetBool("login.guest")
	guestChannels = viper.GetStringSlice("login.lists")
	logger.Infof("是否开启游客登陆: %t, 渠道列表: %v", enableGuest, guestChannels)

	// 语音相关配置
	config.AppId = viper.GetString("voice.appid")
	config.AppKey = viper.GetString("voice.appkey")

	if config.Heartbeat < 5 {
		config.Heartbeat = 5
	}

	messages = viper.GetStringSlice("broadcast.message")

	logger.Debugf("version infomation: %+v", config)
	logger.Debugf("广播消息: %v", messages)

	fu := viper.GetBool("update.force")
	logger.Infof("是否强制更新: %t", fu)
	config.ForceUpdate = fu

	router := mux.NewRouter()
	router.Handle("/v1/user/login/query", nex.Handler(queryHandler)).Methods("POST") //三方登录
	// router.Handle("/v1/user/login/guest", nex.Handler(guestLoginHandler)).Methods("POST") //来宾登录
	router.Handle("/v1/user/login/guest", func() http.Handler {
		// 简单判断一下
		if !enableGuest {
			return nex.Handler(queryHandler)
		}
		return nex.Handler(guestLoginHandler)
	}()).Methods("POST")
	return router
}

// 游客登录
func guestLoginHandler(r *http.Request, data *protocol.LoginRequest) (*protocol.LoginResponse, error) {
	data.Device.IMEI = data.IMEI
	logger.Infof("%v", data)
	logger.Infof("游客登录IEMEI: %s", data.Device.IMEI)

	user, err := db.QueryGuestUser(data.AppID, data.Device.IMEI)
	if err != nil {
		// 生成一个新用户
		user = &model.User{
			Status:   db.StatusNormal,
			IsOnline: db.UserOffline,
			Role:     db.RoleTypeThird,
			Coin:     defaultCoin,
		}

		if err := db.InsertUser(user); err != nil {
			logger.Error(err.Error())
			return nil, err
		}

		db.RegisterUserLog(user, data.Device, data.AppID, data.ChannelID, protocol.RegTypeThird) //注册记录
	}

	// checkSession(user.Id)

	resp := &protocol.LoginResponse{
		Uid:      user.Id,
		HeadUrl:  "http://wx.qlogo.cn/mmopen/s962LEwpLxhQSOnarDnceXjSxVGaibMRsvRM4EIWic0U6fQdkpqz4Vr8XS8D81QKfyYuwjwm2M2ibsFY8mia8ic51ww/0",
		Sex:      1,
		IP:       host,
		Port:     port,
		FangKa:   user.Coin,
		PlayerIP: ip(r.RemoteAddr),
		Config:   config,
		Messages: messages,
		ClubList: clubs(user.Id),
		Debug:    0, //user.Debug,
	}
	resp.Name = fmt.Sprintf("G%d", resp.Uid)

	// 插入登陆记录
	device := protocol.Device{
		IP:     ip(r.RemoteAddr),
		Remote: r.RemoteAddr,
	}
	db.InsertLoginLog(user.Id, device, data.AppID, data.ChannelID)

	return resp, nil
}

func ip(addr string) string {
	addr = strings.TrimSpace(addr)
	deflt := "127.0.0.1"
	if addr == "" {
		return deflt
	}

	if parts := strings.Split(addr, ":"); len(parts) > 0 {
		return parts[0]
	}

	return deflt
}

func clubs(uid int64) []protocol.ClubItem {
	list, err := db.ClubList(uid)
	if err != nil {
		return []protocol.ClubItem{}
	}

	ret := make([]protocol.ClubItem, len(list))
	for i := range list {
		ret[i] = protocol.ClubItem{
			Id:        list[i].ClubId,
			Name:      list[i].Name,
			Desc:      list[i].Desc,
			Member:    list[i].Member,
			MaxMember: list[i].MaxMember,
		}
	}
	return ret
}

func queryHandler(query *queryRequest) (*queryResponse, error) {
	logger.Infof("%v", query)
	if !enableGuest {
		return forbidGuest, nil
	}

	for _, s := range guestChannels {
		if query.ChannelId == s {
			return accepetGuest, nil
		}
	}

	return forbidGuest, nil
}
