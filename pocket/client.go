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
)

const Auth_Request_Api = "https://getpocket.com/v3/oauth/request"
const Auth_Authorize_Api = "https://getpocket.com/v3/oauth/authorize"
const Auth_Request_Code = "https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s"
const POCKETRC = ".pocketrc"

type Client struct {
  httpClient *http.Client
  ConsumerKey string
  RedirectUrl string
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

func (c Client) readAccessToken() string {
  usr, err := user.Current()
  checkError(err)
  configPath := path.Join(usr.HomeDir, POCKETRC)
  file, err := os.OpenFile(configPath, os.O_RDONLY, 0744)
  checkError(err)
  defer file.Close()
  rl := bufio.NewReader(file)
  configBytes, err := ioutil.ReadAll(rl)
  checkError(err)
  usrConfig := config{}
  err = json.Unmarshal(configBytes, usrConfig)
  checkError(err)
  return usrConfig.AccessToken
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

func (c *Client) Init() {
  body := url.Values{}
  body.Set("consumer_key", c.ConsumerKey)
  body.Set("redirect_uri", c.RedirectUrl)
  fmt.Println(body.Encode())
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
  } else {
    fmt.Println("authorize failed")
    os.Exit(1)
  }
}


