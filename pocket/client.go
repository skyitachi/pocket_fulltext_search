package pocket

import (
  "net/http"
  "net/url"
  "bytes"
  "log"
  "io/ioutil"
  "fmt"
  "encoding/json"
  "os/user"
  "os"
  "path"
  "bufio"
  "errors"
  "time"
)

const Auth_Request_Api = "https://getpocket.com/v3/oauth/request"
const Auth_Authorize_Api = "https://getpocket.com/v3/oauth/authorize"
const Auth_Request_Code = "https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s"
const POCKET_GET_API = "https://getpocket.com/v3/get"
const POCKETRC = ".pocketrc"

type Client struct {
  httpClient *http.Client
  ConsumerKey string
  RedirectUrl string
  accessToken string
  init bool
}

type config struct {
  AccessToken string `json:"access_token"`
}

type accessTokenPayLoad struct {
  ConsumerKey string `json:"consumer_key"`
  Code string `json:"code"`
}


func NewClient(consumerKey string) *Client {
  c := &Client{
    httpClient: &http.Client{},
    ConsumerKey: consumerKey,
    RedirectUrl: "https://www.skyitachi.cn",
  }
  return c
}

func checkError(err error) {
  if err != nil {
    log.Println(err.Error())
  }
}

func (c Client) storeAccessToken(accessToken string) {
  usr, err := user.Current()
  checkError(err)
  configPath := path.Join(usr.HomeDir, POCKETRC)
  file, err := os.OpenFile(configPath, os.O_RDWR | os.O_CREATE, 0744)
  defer file.Close()
  checkError(err)
  userConfig := config{
    AccessToken: accessToken,
  }
  configBytes, err := json.Marshal(userConfig)
  checkError(err)
  file.Write(configBytes)
  fmt.Printf("access_token get successfully\n")
}

func (c Client) readAccessToken() (string, error) {
  usr, err := user.Current()
  if err != nil {
    return "", err
  }
  configPath := path.Join(usr.HomeDir, POCKETRC)
  file, err := os.OpenFile(configPath, os.O_RDONLY, 0744)
  if err != nil {
    return "", err
  }
  defer file.Close()
  rl := bufio.NewReader(file)
  configBytes, err := ioutil.ReadAll(rl)
  if err != nil {
    return "", err
  }
  usrConfig := &config{}
  err = json.Unmarshal(configBytes, usrConfig)
  if err != nil {
    return "", err
  }
  return usrConfig.AccessToken, nil
}

func (c Client) getAccessToken(payLoad accessTokenPayLoad) {
  jsonStr, err := json.Marshal(payLoad)
  log.Printf("payload is %s\n", string(jsonStr))
  checkError(err)
  req, err := http.NewRequest("POST", Auth_Authorize_Api, bytes.NewBufferString(string(jsonStr)))
  req.Header.Set("Content-Type", "application/json; charset=UTF-8")
  req.Header.Set("X-Accept", "application/json")
  resp, err := c.httpClient.Do(req)
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    fmt.Println(resp.Status)
    os.Exit(1)
  }
  resBytes, err := ioutil.ReadAll(resp.Body)
  fmt.Printf("response is %s\n", string(resBytes))
  checkError(err)
  var body map[string]interface{}
  err = json.Unmarshal(resBytes, &body)
  checkError(err)
  c.storeAccessToken(body["access_token"].(string))
}

func (c *Client) fetchSimpleJSON(payLoad interface{}) ([]SimpleItem, error){
  payLoadBytes, err := json.Marshal(payLoad)
  fmt.Println(string(payLoadBytes))
  if err != nil {
    log.Fatal("unexpected payload: ", payLoad)
  }
  req, err := http.NewRequest("POST", POCKET_GET_API, bytes.NewBuffer(payLoadBytes))
  req.Header.Set("Content-Type", "application/json; charset=UTF-8")
  resp, err := c.httpClient.Do(req)
  if err != nil {
    return []SimpleItem{}, err
  } else if resp.StatusCode != http.StatusOK {
    log.Printf("fetchJSON error %s\n", resp.Status)
    return []SimpleItem{}, errors.New(resp.Status)
  }
  defer resp.Body.Close()
  contentBytes, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return []SimpleItem{}, err
  }
  fmt.Println(string(contentBytes))
  res := &Response{}
  err = json.Unmarshal(contentBytes, res)
  if err != nil {
    return []SimpleItem{}, err
  }
  ret := []SimpleItem{}
  for _, v := range res.ItemMap {
    ret = append(ret, v)
  }
  return ret, nil
}

func (c *Client) fetchCompleteJSON(payLoad interface{}) ([]CompleteItem, error){
  payLoadBytes, err := json.Marshal(payLoad)
  fmt.Println(string(payLoadBytes))
  if err != nil {
    log.Fatal("unexpected payload: ", payLoad)
  }
  req, err := http.NewRequest("POST", POCKET_GET_API, bytes.NewBuffer(payLoadBytes))
  req.Header.Set("Content-Type", "application/json; charset=UTF-8")
  resp, err := c.httpClient.Do(req)
  if err != nil {
    return []CompleteItem{}, err
  } else if resp.StatusCode != http.StatusOK {
    log.Printf("fetchJSON error %s\n", resp.Status)
    return []CompleteItem{}, errors.New(resp.Status)
  }
  defer resp.Body.Close()
  contentBytes, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return []CompleteItem{}, err
  }
  fmt.Println(string(contentBytes))
  res := CompleteResponse{}
  err = json.Unmarshal(contentBytes, &res)
  if err != nil {
    // list 字段为空时, 会变成[], 兼容下
    oRes := map[string]interface{}{}
    err = json.Unmarshal(contentBytes, &oRes)
    if err != nil {
      return []CompleteItem{}, err
    }
    _, ok := oRes["list"]
    if ok {
      return []CompleteItem{}, nil
    } else {
      return []CompleteItem{}, errors.New("unexpected response from pocket")
    }
  }
  ret := []CompleteItem{}
  for _, v := range res.ItemMap {
    // status: 2 - deleted
    if v.Status == "2" {
      continue
    }
    ret = append(ret, v)
  }
  return ret, nil
}

func (c *Client) GetArchiveList(count int, offset int) ([]SimpleItem, error){
  //payload := c.NewArchiveSimplePayload(count, offset)
  payload := c.NewArchiveCompletePayload(count, offset)
  return c.fetchSimpleJSON(payload)
}

func (c *Client) GetUnreadList(count int, offset int) ([]CompleteItem, error) {
  payload := c.NewUnreadCompletePayload(count, offset)
  return c.fetchCompleteJSON(payload)
}

func (c *Client) GetAllList(count int, offset int) ([]CompleteItem, error) {
  payload := c.NewAllPayload(count, offset)
  return c.fetchCompleteJSON(payload)
}

func (c *Client) GetLatestList(since time.Time) ([]CompleteItem, error) {
  payload := c.NewLatestPayload(since)
  return c.fetchCompleteJSON(payload)
}

func (c *Client) GetListAfter(count int, offset int, after time.Time) ([]CompleteItem, error) {
  payload := c.NewLatestPayload(after)
  payload.Count = count
  payload.Offset = offset
  payload.Sort = "oldest"
  return c.fetchCompleteJSON(payload)
}

func (c *Client) Init() {
  accessToken, err := c.readAccessToken()
  if err == nil && len(accessToken) > 0 {
    c.init = true
    c.accessToken = accessToken
    log.Println("read accesstoken from config successfully")
    return
  } else {
    log.Fatal(err.Error())
  }
  body := url.Values{}
  body.Set("consumer_key", c.ConsumerKey)
  body.Set("redirect_uri", c.RedirectUrl)
  req, err := http.NewRequest("POST", Auth_Request_Api, bytes.NewBufferString(body.Encode()))
  checkError(err)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
  resp, err := c.httpClient.Do(req)
  checkError(err)
  if resp.StatusCode != http.StatusOK {
    fmt.Println(resp.Status)
    os.Exit(1)
  }
  defer resp.Body.Close()
  resBody, err := ioutil.ReadAll(resp.Body)
  checkError(err)
  values, err := url.ParseQuery(string(resBody))
  fmt.Println(values)
  checkError(err)
  fmt.Printf("open the url in the browser: \n %s\n",
    fmt.Sprintf(Auth_Request_Code, values.Get("code"), c.RedirectUrl))
  fmt.Printf("authorization done? (Y/N): ")
  var choice string
  fmt.Scanf("%s\n", &choice)
  if choice == "Y" {
    payload := accessTokenPayLoad{c.ConsumerKey, values.Get("code")}
    c.getAccessToken(payload)
    c.init = true
  } else {
    fmt.Println("authorize failed")
    os.Exit(1)
  }
}


