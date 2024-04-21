package main

import (
	"log"
  "net/http"
  "flag"
  "os"
  "html/template"
  "time"
  //"context"
  "strings"
  "encoding/json"
  //"encoding/hex"
  "crypto/tls"
  
  "fmt"
  supa "github.com/nedpals/supabase-go"
  //"github.com/alexedwards/scs/postgresstore"
  "github.com/alexedwards/scs/v2"
  

  "github.com/go-playground/form/v4"
  //"context"
  "snippetbox.evanread.net/internal/models"
)

type application struct {
  debug bool
  errorLog *log.Logger
  infoLog *log.Logger
  snippets models.SnippetModelInterface //*models.SnippetModel
  users models.UserModelInterface //*models.UserModel
  templateCache map[string]*template.Template
  formDecoder *form.Decoder
  sessionManager *scs.SessionManager
}



type SupabaseSessionStore struct {
  client *supa.Client
}

func NewSupabaseSessionStore(client *supa.Client) *SupabaseSessionStore {
  return &SupabaseSessionStore{client: client}
}

func (s *SupabaseSessionStore) Delete(token string) error {
  fmt.Println("reached delete session")
  
  sessionData := SessionData{}
  
  err := s.client.DB.From("sessions").Delete().Eq("token", token).Execute(&sessionData)
  if err != nil && !strings.Contains(err.Error(), "session not found") {
      return err
  }

  return nil
}

type SessionData struct {
  Token string `json:"token"`
  Data []byte `json:"data"`
  Expiry time.Time `json:"expiry"`
}

type CommitResponse struct {
  Msg json.RawMessage `json:"msg"`
}

func (s *SupabaseSessionStore) Find(token string) ([]byte, bool, error) {
  fmt.Println("reached find session")
  
  sessionData := SessionData{}

  err := s.client.DB.From("sessions").Select("data").Eq("id", token).Execute(&sessionData)
  if err != nil {
    return nil, false, err
  }

  fmt.Println(sessionData)
  
  if sessionData.Token == "" || sessionData.Expiry.Before(time.Now()) {
      return nil, false, nil
  }

  return sessionData.Data, true, nil
}

type BufferData struct {
  Type string `json:"type"`
  Data []int `json:"data"`
}

func (s *SupabaseSessionStore) Commit(token string, b []byte, expiry time.Time) error {
  fmt.Println("reached commit session")

  //hexString := hex.EncodeToString(b)
  
  sessionData := SessionData{
    Token: token,
    Data: b,
    Expiry: expiry,
  }

  commitResp := CommitResponse{}

  //hexString := hex.EncodeToString(byteData)
  //fmt.Println(hexString)

  //sessionData.Data = hexString

  
  
  err := s.client.DB.From("sessions").Upsert(sessionData).Execute(&commitResp.Msg)
  if err != nil {
    fmt.Println(err)
    return err
  }

  //fmt.Println(commitResp)
  
  return nil
}

func (s *SupabaseSessionStore) All() (map[string][]byte, error) {
  activeSessions := make(map[string][]byte)

  sessions := []SessionData{}
  err := s.client.DB.From("sessions").Select("token, data").Execute(&sessions)
  if err != nil {
    return nil, err
  }

  for _, session := range sessions {
    if session.Expiry.After(time.Now()) {
      activeSessions[session.Token] = session.Data
    }
  }

  return activeSessions, nil
}


func main() {
  addr := flag.String("addr", ":4000", "HTTP network address")
  debug := flag.Bool("debug", false, "Enable debug mode")
  flag.Parse()

  infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
  errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

  // Using Supabase Postgresql as DB
  
  //fmt.Println(serviceRoleKey)
  //fmt.Println(apiURL)
  db, err := openDB()
  if err != nil {
    errorLog.Fatal(err)
  }
  //var results []map[string]interface{}
  //err = db.DB.From("snippets").Select("*").Execute(&results)
  //if err != nil {
    //panic(err)
  //}
  //fmt.Println(results)

  templateCache, err := newTemplateCache()
  if err != nil {
    errorLog.Fatal(err)
  }

  formDecoder := form.NewDecoder()

  sessionManager := scs.New()
  sessionManager.Store = NewSupabaseSessionStore(db)
  sessionManager.Lifetime = 12 * time.Hour
  sessionManager.Cookie.Secure = true

  tlsConfig := &tls.Config{
    CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
  }

  app := &application{
    debug: *debug,
    errorLog: errorLog,
    infoLog: infoLog,
    snippets: &models.SnippetModel{DB: db},
    users: &models.UserModel{DB: db},
    templateCache: templateCache,
    formDecoder: formDecoder,
    sessionManager: sessionManager,
  }
  
  srv := &http.Server{
    Addr: *addr,
    ErrorLog: errorLog,
    Handler: app.routes(),
    TLSConfig: tlsConfig,
    IdleTimeout: time.Minute,
    ReadTimeout: 5 * time.Second,
    WriteTimeout: 10 * time.Second,
  }

  infoLog.Printf("Starting server on %s", *addr)
  err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
  errorLog.Fatal(err)
}

func openDB() (*supa.Client, error) {
  serviceRoleKey := os.Getenv("service_role")
  apiURL := os.Getenv("api_url")
  supabase := supa.CreateClient(apiURL, serviceRoleKey)
  
  return supabase, nil
}