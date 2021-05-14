package controller

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	requests "github.com/levigross/grequests"
	"github.com/yametech/devops/pkg/store"
	"github.com/yametech/devops/pkg/utils"
	"github.com/yametech/message-center/pkg/common"
	"github.com/yametech/message-center/pkg/proc"
	"github.com/yametech/message-center/pkg/resource/message"
	"io/ioutil"
	"log"
	"net/http"
	urlPkg "net/url"
	"strings"
	"time"
)

var _ Controller = &messageController{}

var Token string

type messageController struct {
	store.IKVStore
	proc *proc.Proc
}

func NewUserController(store store.IKVStore) *messageController {
	server := &messageController{
		IKVStore: store,
		proc:     proc.NewProc(),
	}

	return server
}

func (a *messageController) Run() error {
	go a.recvtoken(a.proc.Error())
	time.Sleep(time.Second * 1)
	a.proc.Add(a.recvUser)
	a.proc.Add(a.recvMessage)

	return <-a.proc.Start()
}

func (a *messageController) recvUser(errC chan<- error) {
	log.Printf("start recv user\n")

	for {
		departments, err := GetDepartment(Token)
		if err != nil {
			errC <- err
			return
		}
		if departments.ErrCode != 0 {
			errC <- errors.New("获取department失败")
		}
		departments.DeptIdList = append(departments.DeptIdList, 1)
		for _, department := range departments.DeptIdList {
			go a.DepartmentChildrenList(Token, department)
		}
		log.Printf("save user finish\n")
		time.Sleep(time.Minute * 10)
	}

}

func getToken() (string, error) {
	tokenURL := fmt.Sprintf("%s/gettoken?appkey=%s&appsecret=%s", common.DingURL, common.AppKey, common.AppSecret)
	data := urlPkg.Values{}
	req, err := http.NewRequest("GET", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 30 * time.Second, Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	respData, err := ioutil.ReadAll(resp.Body)
	structData := make(map[string]interface{})
	if err = json.Unmarshal(respData, &structData); err != nil {
		return "", err
	}

	token := structData["access_token"]
	return token.(string), nil
}

func GetDepartment(token string) (*message.ReqDepart, error) {
	ro := &requests.RequestOptions{
		Headers: map[string]string{
			"Content-type": "application/json",
			"Accept":       "application/json",
			"User-Agent":   "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
		},
	}
	url := fmt.Sprintf("https://oapi.dingtalk.com/department/list_ids?access_token=%s&id=1", token)
	resp, err := requests.Get(url, ro)
	if err != nil {
		return nil, err
	}
	result := message.ReqDepart{}
	err = resp.JSON(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (a *messageController) DepartmentChildrenList(token string, department int) {
	url := fmt.Sprintf("https://oapi.dingtalk.com/user/simplelist?access_token=%s&department_id=%d", token, department)
	ro := &requests.RequestOptions{
		Headers: map[string]string{
			"Content-type": "application/json",
			"Accept":       "application/json",
			"User-Agent":   "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
		},
	}
	resp, err := requests.Get(url, ro)
	if err != nil {
		log.Printf("department 获取user失败 %s", err.Error())
	}
	result := message.ReqUser{}
	err = resp.JSON(&result)
	if err != nil {
		log.Printf("department 获取user失败 %s", err.Error())
	}
	if result.ErrCode != 0 {
		log.Printf("department 获取user失败 %s", err.Error())
	}
	for _, userData := range result.UserList {
		user := message.User{
			Spec: message.UserSpec{
				Name:   userData.Name,
				DingID: userData.Userid,
			},
		}
		user.UUID = userData.Userid
		log.Printf("save one user name: %s\n", user.Spec.Name)
		user.GenerateVersion()
		_, _, err := a.IKVStore.Apply(common.DefaultNamespace, common.User, user.UUID, &user, false)
		if err != nil {
			log.Printf("department 保存 user %s", err.Error())
		}
	}
}

func (a *messageController) recvMessage(errC chan<- error) {
	msgObjs, err := a.List(common.DefaultNamespace, common.MessageCenter, "", map[string]interface{}{}, 0, 0)
	if err != nil {
		errC <- err
	}
	msgCoder := store.GetResourceCoder(string(message.Kind))
	if msgCoder == nil {
		errC <- fmt.Errorf("(%s) %s", message.Kind, "coder not exist")
	}
	var version int64

	mgsWatchChan := store.NewWatch(msgCoder)
	for _, item := range msgObjs {
		msgObj := &message.Message{}
		if err := utils.UnstructuredObjectToInstanceObj(&item, msgObj); err != nil {
			log.Printf("unmarshal message error %s\n", err)
		}
		if msgObj.GetResourceVersion() > version {
			version = msgObj.GetResourceVersion()
		}
		go a.handleMessage(Token, msgObj)
	}
	for {
		select {
		case item, ok := <-mgsWatchChan.ResultChan():
			if !ok {
				errC <- fmt.Errorf("recvMsg watch channal close")
			}
			if item.GetUUID() == "" {
				continue
			}
			msgObj := &message.Message{}
			if err := utils.UnstructuredObjectToInstanceObj(&item, msgObj); err != nil {
				log.Printf("receive message UnmarshalInterfaceToResource error %s\n", err)
				continue
			}
			go a.handleMessage(Token, msgObj)
		}
	}

}

func (a *messageController) handleMessage(token string, msgObj *message.Message) {
	if msgObj.Spec.Status == message.Success {
		return
	}
	var sendUserList []string
	for _, userName := range msgObj.Spec.SendUser {
		userObj := message.User{}
		err := a.GetByFilter(common.DefaultNamespace, common.User, &userObj, map[string]interface{}{"spec.name": userName})
		if err != nil {
			log.Printf("find user error %s\n", err.Error())
			return
		}
		sendUserList = append(sendUserList, userObj.Spec.DingID)
	}
	sendUser := strings.Join(sendUserList, ",")

	ro := &requests.RequestOptions{
		Headers: map[string]string{
			"Content-type": "application/json",
			"Accept":       "application/json",
			"User-Agent":   "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
		},
		JSON: map[string]interface{}{
			"agent_id":    common.Agent,
			"userid_list": sendUser,
			"msg": map[string]interface{}{
				"msgtype": "text",
				"text": map[string]interface{}{
					"content": msgObj.Spec.Content,
				},
			},
		},
	}
	url := fmt.Sprintf("https://oapi.dingtalk.com/topapi/message/corpconversation/asyncsend_v2?access_token=%s", token)
	resp, err := requests.Post(url, ro)
	if err != nil {
		log.Println("发送echoer错误: ", err.Error())
		msgObj.Spec.Status = message.Fail
		a.modifyMsgObjStatus(msgObj)
		return
	}
	msgObj.Spec.Status = message.Success
	a.modifyMsgObjStatus(msgObj)
	var result interface{}
	err = resp.JSON(&result)
	if err != nil {
		return
	}
	log.Println(result)
}

func (a *messageController) modifyMsgObjStatus(obj *message.Message) {
	_, _, err := a.Apply(common.DefaultNamespace, common.MessageCenter, obj.UUID, obj, false)
	if err != nil {
		log.Printf("更新msg状态失败 err:%s ", err.Error())
	}

}

func (a *messageController) recvtoken(errC chan<- error) {
	for {
		err := errors.New("")
		Token, err = getToken()
		if err != nil {
			errC <- err
		}
		time.Sleep(time.Hour * 1)
	}
}
