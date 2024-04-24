package models

import (
  "time"
  supa "github.com/nedpals/supabase-go"
 "fmt"
  "strconv"
  "log"
  "math/rand"
)

type SnippetModelInterface interface {
  Insert(title string, content string, expires int) (int, error)
  Get(id int) (*Snippet, error)
  Latest() ([]*Snippet, error)
}


type Snippet struct {
  ID int `json:"id"`
  Title string `json:"title"`
  Content string `json:"content"`
  Created time.Time `json:"created"`
  Expires time.Time `json:"expires"`
}

type SnippetModel struct {
  DB *supa.Client
}

func (m *SnippetModel) Insert(title, content string, expires int) (int, error) {
  currentTime := time.Now()
  expirationTime := currentTime.AddDate(0, 0, 7)

  rand.Seed(time.Now().UnixNano())
  randomNumber := 10000000 + rand.Intn(90000000)

  snippet := Snippet{
    ID:      randomNumber,
    Title:    title,
    Content: content,
    Created: currentTime,
    Expires: expirationTime,
  }

  var results []Snippet
  err := m.DB.DB.From("snippets").Insert(snippet).Execute(&results)
  if err != nil {
    panic(err)
  }

  //fmt.Println(results)
  //fmt.Println()
  id := 0
  for _, snippetrange := range results {
      id = snippetrange.ID
  }



  return id, nil
}

func (m *SnippetModel) Get(id int) (*Snippet, error) {
  //currentTime := time.Now()


  var results []Snippet
  err := m.DB.DB.From("snippets").Select("id, title, content, created, expires").Eq("id", strconv.Itoa(id)).Execute(&results)
  if err != nil {
    panic(err)
  }

  //if results == nil {
    //return nil, ErrNoRecord
  //}

  if len(results) == 0 {
      return nil, ErrNoRecord
   }

  //fmt.Println(results)

  return &results[0], nil
}

func (m *SnippetModel) Latest() ([]*Snippet, error) {
  currentTime := time.Now().Format(time.RFC3339Nano)


  var results []*Snippet

  fmt.Println("reached latest 1")

  err := m.DB.DB.From("snippets").
  Select("id, title, content, created, expires").
  //Order("id", false). // cant get order to work, something weird with package
  Limit(10).
  Gt("expires", currentTime). // Assuming this is how you compare dates in your query
  Execute(&results)

  fmt.Println("reached latest 2")

  fmt.Println(results)

  if err != nil {
    log.Fatalf("Error executing query: %v", err)
  }

  //fmt.Println(results)


  return results, nil
}
