package pocket

import "time"

type Payload struct {
  ConsumerKey string `json:"consumer_key"`
  AccessToken string `json:"access_token"`
  DetailType string `json:"detailType"`
  ContentType string `json:"contentType"`
  Count int `json:"count,omitempty"`
  Offset int `json:"offset,omitempty"`
  State string `json:"state,omitempty"`
  Favorite int `json:"favorite,omitempty"`
  Tag string `json:"tag,omitempty"`
  Sort string `json:"sort,omitempty"`
  Since int64 `json:"since,omitempty"`
}

func (c Client) NewArchiveSimplePayload(count int, offset int) Payload {
  return Payload{
    ConsumerKey: c.ConsumerKey,
    AccessToken: c.accessToken,
    DetailType: "simple",
    ContentType: "article",
    State: "archive",
    Sort: "newest",
    Count: count,
    Offset: offset,
  }
}

func (c Client) NewArchiveCompletePayload(count int, offset int) Payload {
  return Payload{
    ConsumerKey: c.ConsumerKey,
    AccessToken: c.accessToken,
    DetailType: "complete",
    ContentType: "article",
    State: "archive",
    Sort: "newest",
    Count: count,
    Offset: offset,
  }
}

func (c Client) NewUnreadSimplePayload(count int, offset int) Payload {
  return Payload{
    ConsumerKey:c.ConsumerKey,
    AccessToken:c.accessToken,
    DetailType:"simple",
    ContentType:"article",
    State: "unread",
    Sort: "newest",
    Count: count,
    Offset: offset,
  }
}

func (c Client) NewUnreadCompletePayload(count int, offset int) Payload {
  payload := c.NewUnreadSimplePayload(count, offset)
  payload.DetailType = "complete"
  return payload
}

func (c Client) NewAllPayload(count int, offset int) Payload {
  payload := c.NewUnreadCompletePayload(count, offset)
  payload.State = "all"
  payload.Sort = "oldest"
  return payload
}

func (c Client) NewLatestPayload(since time.Time) Payload {
  return Payload{
    ConsumerKey: c.ConsumerKey,
    AccessToken: c.accessToken,
    DetailType: "complete",
    ContentType: "article",
    Since: int64(since.Unix()),
  }
}

func (c Client) NewAfterPayload(since time.Time, count int, offset int) Payload {
  payload := c.NewUnreadCompletePayload(count, offset)
  payload.State = "all"
  payload.Sort = "oldest"
  payload.Since = int64(since.Unix())
  return payload
}