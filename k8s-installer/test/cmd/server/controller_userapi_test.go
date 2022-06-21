package server

import (
	"encoding/json"
	"fmt"
	inSchema "k8s-installer/schema"
	testUtil "k8s-installer/test/testutil"
	"net/http"
	"testing"
)

func TestControllerUserApi(t *testing.T) {
	//wait 3 seconds for the server to starts
	//time.Sleep(time.Second * 3)
	t.Run("API User Login", testAPIUserLogin)
	t.Run("API User Create", testAPIUserCreate)
	t.Run("API Users List", testAPIUsersList)
	t.Run("API User Get", testAPIUserGet)
	t.Run("API User Update", testAPIUserUpdate)
	t.Run("API User Delete", testAPIUserDelete)
}

//用户删除测试
func testAPIUserDelete(t *testing.T) {
	token := testUtil.GetToken()
	testhost, _ := testUtil.GetHost()
	deleteUsername := "admin1"
	requestURL := fmt.Sprintf("http://%v/user/v1/%v", testhost, deleteUsername)
	body, httpStatus := testUtil.RequestTest(t, requestURL, http.MethodDelete, nil, map[string]string{"token": token}, true, true)

	if !testUtil.HttpStatusIsExpected(t, []int{http.StatusOK}, httpStatus) {
		return
	}
	UsersListRequestURL := fmt.Sprintf("http://%v/user/v1/users", testhost)
	body, _ = testUtil.RequestTest(t, UsersListRequestURL, http.MethodGet, nil, map[string]string{"token": token}, true, true)
	usersList := []inSchema.User{}
	json.Unmarshal(body, &usersList)
	testUtil.IntIsExpected(t, "testAPIUserDelete", []int{1}, []int{len(usersList)})

}

//用户更新测试
func testAPIUserUpdate(t *testing.T) {
	token := testUtil.GetToken()
	testhost, _ := testUtil.GetHost()
	updateUsername := "admin1"
	requestURL := fmt.Sprintf("http://%v/user/v1/%v", testhost, updateUsername)
	postBody := `
	{
		"username": "admin1",
		"password": "123456"
	}
	`
	json.Marshal(postBody)
	body, httpStatus := testUtil.RequestTest(t, requestURL, http.MethodPut, []byte(postBody), map[string]string{"token": token}, true, true)
	if !testUtil.HttpStatusIsExpected(t, []int{http.StatusOK}, httpStatus) {
		return
	}
	t.Log()
	ret := inSchema.User{}
	json.Unmarshal(body, &ret)
	testUtil.StringExpected(t, "testAPIUserUpdate", []string{"123456"}, []string{ret.Password})
}

//用户登录测试
func testAPIUserLogin(t *testing.T) {
	testhost, _ := testUtil.GetHost()
	requestURL := fmt.Sprintf("http://%v/user/v1/login", testhost)

	type T struct {
		Token       string        `json:"token"`
		ExpiredDate string        `json:"expired_date"`
		Permission  uint64        `json:"permission"`
		User        inSchema.User `json:"user"`
	}

	var postBody = `
	{
		"username": "admin",
		"password": "123"
	}
	`

	body, httpStatus := testUtil.RequestTest(t, requestURL, http.MethodPost, []byte(postBody), nil, true, true)
	testUtil.HttpStatusIsExpected(t, []int{http.StatusOK}, httpStatus)
	if !testUtil.HttpStatusIsExpected(t, []int{http.StatusOK}, httpStatus) {
		return
	}

	var ret T
	json.Unmarshal(body, &ret)
	testUtil.StringExpected(t, "testAPIUserLogin", []string{"admin"}, []string{ret.User.Username})
}

//创建用户测试
func testAPIUserCreate(t *testing.T) {
	token := testUtil.GetToken()
	testhost, _ := testUtil.GetHost()
	requestURL := fmt.Sprintf("http://%v/user/v1/", testhost)

	var postBody = `
	{
		"username": "admin1",
		"password": "123"
	}
	`
	body, httpStatus := testUtil.RequestTest(t, requestURL, http.MethodPost, []byte(postBody), map[string]string{"token": token}, true, true)
	if !testUtil.HttpStatusIsExpected(t, []int{http.StatusOK}, httpStatus) {
		return
	}
	var ret inSchema.User
	json.Unmarshal(body, &ret)
	testUtil.StringExpected(t, "testAPIUserUpdate", []string{"admin1"}, []string{ret.Username})
}

//单个用户测试
func testAPIUserGet(t *testing.T) {
	token := testUtil.GetToken()
	testhost, _ := testUtil.GetHost()
	testUser := "admin"
	requestURL := fmt.Sprintf("http://%v/user/v1/%v", testhost, testUser)
	body, httpStatus := testUtil.RequestTest(t, requestURL, http.MethodGet, nil, map[string]string{"token": token}, true, true)
	if !testUtil.HttpStatusIsExpected(t, []int{http.StatusOK}, httpStatus) {
		return
	}
	var userResponse inSchema.User
	err := json.Unmarshal(body, &userResponse)
	if err != nil {
		t.Errorf("Failed to parse body data to obj User due to %s", err)
		return
	}
	testUtil.StringExpected(t, "testAPIUserGet", []string{testUser}, []string{userResponse.Username})
}

//用户列表测试
func testAPIUsersList(t *testing.T) {
	token := testUtil.GetToken()
	testhost, _ := testUtil.GetHost()
	requestURL := fmt.Sprintf("http://%v/user/v1/users", testhost)
	body, httpStatus := testUtil.RequestTest(t, requestURL, http.MethodGet, nil, map[string]string{"token": token}, true, true)
	if !testUtil.HttpStatusIsExpected(t, []int{http.StatusOK}, httpStatus) {
		return
	}
	usersResponse := []inSchema.User{}
	err := json.Unmarshal(body, &usersResponse)
	if err != nil {
		t.Errorf("Failed to parse body data to obj User due to %s", err)
		return
	}
	testUtil.IntIsExpected(t, "testAPIUsersList", []int{2}, []int{len(usersResponse)})
}
