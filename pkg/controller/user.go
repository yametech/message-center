package controller

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	requests "github.com/levigross/grequests"
	"github.com/yametech/devops/pkg/store"
	"github.com/yametech/message-center/pkg/common"
	"github.com/yametech/message-center/pkg/proc"
	"github.com/yametech/message-center/pkg/resource/message"
	"io/ioutil"
	"net/http"
	urlPkg "net/url"
	"strings"
	"time"
)

var _ Controller = &AppServiceController{}

type AppServiceController struct {
	store.IKVStore
	proc *proc.Proc
}

func NewUserController(store store.IKVStore) *AppServiceController {
	server := &AppServiceController{
		IKVStore: store,
		proc:     proc.NewProc(),
	}

	return server
}

func (a *AppServiceController) Run() error {
	a.proc.Add(a.recvUser)

	return <-a.proc.Start()
}

func (a *AppServiceController) recvUser(errC chan<- error) {
	token, err := getToken()
	if err != nil {
		errC <- err
		return
	}
	departments, err := GetDepartment(token)
	if err != nil {
		errC <- err
		return
	}
	if departments.ErrCode != 0 {
		errC <- errors.New("获取department失败")
	}
	departments.DeptIdList = append(departments.DeptIdList, 1)
	for _, department := range departments.DeptIdList {
		err = a.DepartmentChildrenList(token, department)
		if err != nil {
			errC <- err
			return
		}
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

func (a *AppServiceController) DepartmentChildrenList(token string, department int) error {
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
		return err
	}
	result := message.ReqUser{}
	err = resp.JSON(&result)
	if err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return errors.New("查询user失败")
	}
	for _, userData := range result.UserList {
		user := message.User{
			Spec: message.UserSpec{
				Name:   userData.Name,
				DingID: userData.Userid,
			},
		}
		user.GenerateVersion()
		_, _, err := a.IKVStore.Apply(common.DefaultNamespace, common.User, user.UUID, &user, false)
		if err != nil {
			return err
		}
	}
	return nil
}
