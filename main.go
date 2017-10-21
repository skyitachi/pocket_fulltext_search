package main

import (
  "log"
  "fmt"
  "skyitachi/pocket_fulltext_search/pocket"
  "skyitachi/pocket_fulltext_search/collector"
  "time"
  "strconv"
  "flag"
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

func main() {
  // Init Pocket Client
  client := pocket.NewClient(consumer_key)
  client.Init()
  log.Println("pocket client init successfully")
  // Init ElasticSearch Client
  es, err := pocket.NewElasticClient()
  check(err)
  es.Init()
  log.Println("elasticsearch client init successfully")

  // parse command line
  syncPtr := flag.Bool("sync", false, "同步数据")
  flag.Parse()

  if *syncPtr {
    fmt.Println("start sync data")
    collector := collector.NewCollector(client, es, time.Second * 10, time.Second * 5)
    go collector.Sync()
    time.Sleep(20 * time.Second)
    collector.Exit <- 1
  } else {
    cList, err := client.GetAllList(1, 1)
    check(err)
    fmt.Printf("get %d items from pocket\n", len(cList))
    for _, v := range cList {
      fmt.Println(v.Create)
      ret, err := strconv.ParseInt(v.Create, 10, 64)
      if err != nil {
        continue
      }
      fmt.Println(time.Unix(ret, 0))
    }

  }

}
