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
type tagFlags []string

func (i *tagFlags) String() string {
  return "tagFlag"
}
func (i *tagFlags) Set(value string) error {
  *i = append(*i, value)
  return nil
}

type AccessTokenPayLoad struct {
    ConsumerKey string `json:"consumer_key"`
    Code string `json:"code"`
}

var consumer_key = "40534-becce4b35a568bb14eed0fe7"

func main() {
  // parse command line
  var tags tagFlags
  syncPtr := flag.Bool("sync", false, "同步数据")
  rmIndex := flag.Bool("rmIndex", false, "删除现有index")
  searchPtr := flag.Bool("search", false, "搜索文档")
  txtPtr := flag.String("text", "", "搜索文本")
  flag.Var(&tags, "tag", "pocket item tag")
  flag.Parse()

  // Init Pocket Client
  client := pocket.NewClient(consumer_key)
  client.Init()
  log.Println("pocket client init successfully")
  // Init ElasticSearch Client
  es, err := pocket.NewElasticClient()
  check(err)
  es.Init()
  log.Println("elasticsearch client init successfully")

  if *searchPtr {
    if len(tags) > 0 {
      es.SearchByTags(tags)
    } else if len(*txtPtr) > 0 {
      fmt.Printf("search string is %s\n", *txtPtr)
      es.Search(*txtPtr)
    }
  } else if *syncPtr {
    fmt.Println("start sync data")
    collector := collector.NewCollector(client, es, time.Second * 10, time.Second * 5)
    go collector.Sync()
    time.Sleep(20 * time.Second)
    collector.Exit <- 1
  } else if *rmIndex {
    fmt.Println("deleting elastic search index")
    es.RemoveIndex()
    fmt.Println("delete elastic search index successfully")
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
