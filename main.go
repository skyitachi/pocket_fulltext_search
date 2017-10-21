package main

import (
    "net/http"
    //"net/url"
    "bytes"
    "log"
    "fmt"
    "io/ioutil"
    "encoding/json"
    "skyitachi/pocket_fulltext_search/pocket"
)

func check(err error) {
    if err != nil {
        panic(err)
        log.Fatal(err)
    }
}
type AccessTokenPayLoad struct {
    ConsumerKey string `json:"consumer_key"`
    Code string `json:"code"`
}

var consumer_key = "40534-becce4b35a568bb14eed0fe7"
var pocket_api = "https://getpocket.com/v3/oauth/request"
var access_token = "b6511411-f6c5-e220-4be6-d67af3"

func getAccessToken(payLoad interface{}) {
    var tokenApi = "https://getpocket.com/v3/oauth/authorize"
    jsonStr, err := json.Marshal(payLoad)
    fmt.Printf("payload is %s\n", string(jsonStr))
    check(err)
    client := &http.Client{}
    req, err := http.NewRequest("POST", tokenApi, bytes.NewBufferString(string(jsonStr)))
    req.Header.Set("Content-Type", "application/json; charset=UTF-8")
    req.Header.Set("X-Accept", "application/json")
    resp, err := client.Do(req)
    defer resp.Body.Close()
    resBytes, err := ioutil.ReadAll(resp.Body)
    fmt.Printf("response is %s\n", string(resBytes))
    check(err)
    var body map[string]interface{}
    err = json.Unmarshal(resBytes, &body)
    check(err)
    fmt.Printf("access_token is %s\n", body["access_token"])
}

func get(payLoad interface{}) {
    var url = "https://getpocket.com/v3/get"
    jsonBytes, err := json.Marshal(payLoad)
    check(err)
    client := &http.Client{}
    req, err := http.NewRequest("GET", url, bytes.NewBufferString(string(jsonBytes)))
    req.Header.Set("Content-Type", "application/json; charset=UTF-8")
    req.Header.Set("X-Accept", "application/json")
    resp, err := client.Do(req)
    defer resp.Body.Close()
    resBytes, err := ioutil.ReadAll(resp.Body)
    fmt.Printf("response is %s\n", string(resBytes))
    check(err)
    var body map[string]interface{}
    err = json.Unmarshal(resBytes, &body)
    check(err)
}

func main() {
    //body := url.Values{}
    //body.Set("consumer_key", consumer_key)
    //body.Set("redirect_uri", "https://www.skyitachi.cn")
    //client := &http.Client{}
    //req, err := http.NewRequest("POST", pocket_api, bytes.NewBufferString(body.Encode()))
    //check(err)
    //req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
    //resp, err := client.Do(req)
    //check(err)
    //log.Print(resp.Header.Get("Content-Type"))
    //defer resp.Body.Close()
    //resBody, err := ioutil.ReadAll(resp.Body)
    //check(err)
    //fmt.Println(string(resBody))
    //values, err := url.ParseQuery(string(resBody))
    //check(err)
    //fmt.Println(values.Get("code"))
    //fmt.Printf("open the url in your browser \n%s\n",
    //    fmt.Sprintf("https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s", values.Get("code"), "https://www.skyitachi.cn"))
    //var choice string
    //fmt.Scanf("%s\n", &choice)
    //if choice == "Y" {
    //    getAccessToken(AccessTokenPayLoad{ consumer_key, values.Get("code")})
    //}
    //var payload = make(map[string]interface{})
    //payload["consumer_key"] = consumer_key
    //payload["access_token"] = access_token
    //payload["state"] = "all"
    //payload["sort"] = "newest"
    //payload["count"] = 1
    //payload["offset"] = 0
    //payload["detailType"] = "simple"
    //jsonBytes, err := json.Marshal(payload)
    //check(err)
    //fmt.Printf("%s\n", string(jsonBytes))
    //get(payload)
    client := pocket.NewClient(consumer_key)
    client.Init()
}
