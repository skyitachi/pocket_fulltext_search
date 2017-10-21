package pocket

import (
  "time"
  "skyitachi/pocket_fulltext_search/util"
)

type TagItem struct {
  ItemId string `json:"item_id"`
  TagName string `json:"tag"`
}

type SimpleItem struct {
  Id string `json:"item_id"`
  ResolvedId string `json:"resolved_id"`
  Favorite string `json:"favorite,omitempty"`
  Title string `json:"resolved_title"`
  Url string `json:"resolved_url"`
  Excerpt string `json:"excerpt,omitempty"`
}


type CompleteItem struct {
  Id string `json:"item_id"`
  Status string `json:"status"`
  ResolvedId string `json:"resolved_id"`
  Favorite string `json:"favorite,omitempty"`
  Title string `json:"resolved_title"`
  Url string `json:"resolved_url"`
  Excerpt string `json:"excerpt,omitempty"`
  Create string `json:"time_added,omitempty"`
  Update string `json:"time_updated,omitempty"`
  Tags map[string]TagItem `json:"tags,omitempty"`
}

type Response struct {
  Status int `json:"status"`
  Complete int `json:"complete"`
  Since int `json:"since"`
  Error string `json:"error,omitempty"`
  ItemMap map[string]SimpleItem `json:"list,omitempty"`
}

type CompleteResponse struct {
  Status int `json:"status"`
  Complete int `json:"complete"`
  Since int `json:"since"`
  Error string `json:"error,omitempty"`
  ItemMap map[string]CompleteItem `json:"list,omitempty"`
}


func GetTagList(tagMap map[string]TagItem) []string {
  tags := []string{}
  for _, v := range tagMap {
    tags = append(tags, v.TagName)
  }
  return tags
}

func GetNewestTime(list []CompleteItem) (time.Time, error){
  ans := ""
  for _, v := range list {
    if v.Create > ans {
      ans = v.Create
    }
    if v.Update > ans {
      ans = v.Update
    }
  }
  return util.Str2Time(ans)
}
