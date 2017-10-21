package util

import (
  "time"
  "strconv"
)

func Str2Time(str string) (time.Time, error) {
  ret, err := strconv.ParseInt(str, 10, 64)
  if err != nil {
    return time.Now(), err
  }
  return time.Unix(ret, 0), nil
}

func BuildWildCardString(input string) string {
  return "*" + input + "*"
}
