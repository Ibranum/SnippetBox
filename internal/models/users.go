package models

import (
  //"datebase/sql"
  "time"
  supa "github.com/nedpals/supabase-go"
  "errors"
  "strings"
  "golang.org/x/crypto/bcrypt"
  "fmt"
  "github.com/google/uuid"
  //"strings"
  "log"
  "net/http"
  "os"
  "io/ioutil"
  "encoding/json"
  //"strconv"
)

type UserModelInterface interface {
  Insert(name, email, password string) error
  Authenticate(email, password string) (int, error)
  Exists(id uuid.UUID) (bool, error)
  Get(id uuid.UUID) (*User, error)
  PasswordUpdate(id uuid.UUID, currentPassword, newPassword string) error
}

type User struct {
  ID uuid.UUID `json:"id"`
  Name string `json:"name"`
  Email string `json:"email"`
  HashedPassword []byte `json:"hashed_password"`
  Created time.Time `json:"created"`
}

type UserModel struct {
  DB *supa.Client
}

func (m *UserModel) Insert(name, email, password string) error {
  hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
  if err != nil {
    return err
  }

  currentTime := time.Now()
  id := uuid.New()
  fmt.Println(string(hashedPassword))

  userAuth := User{
    ID: id,
    Name: name,
    Email: email,
    HashedPassword: hashedPassword,
    Created: currentTime,
  }

  var results []User
  err = m.DB.DB.From("users").Insert(userAuth).Execute(&results)
  if err != nil {
    if strings.Contains(err.Error(), "users_uc_email") {
      return ErrDuplicateEmail
    }
    
    panic(err)
    return err
  }
  
  return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
  var id int
  var hashedPassword []byte

  type authenticateUser struct {
    ID uuid.UUID `json:"id"`
    HashedPassword []byte `json:"hashed_password"`
    Email string `json:"email"`
  }

  //SELECT id, hashed_password FROM users WHERE email LIKE 'test1@gmail' //this sql query works when executed within supabase
  var userAuth []authenticateUser
  //err := m.DB.DB.From("users").
  //Select("id, hashed_password"). //"id, hashed_password"
  //Eq("email", email). //does not work due to quotes being added in REST request, issue with the nedpals supabase-go package
  //Execute(&userAuth)
  apiKey := os.Getenv("service_role")
  sqlURL := ("https://ijbzmfeptygfasohelez.supabase.co/rest/v1/users?email=eq." + email + "&select=id%2C+hashed_password%2C+email&apikey=" + apiKey)
  //sqlURL := ("https://ijbzmfeptygfasohelez.supabase.co/rest/v1/users?email=eq.jest@gmail.com&select=id%2C+hashed_password%2C+email&apikey=" + apiKey)

  resp, err := http.Get(sqlURL)
  if err != nil {
    log.Fatalf("Failed to make SQL request: %v", err)
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatalf("Failed to read the response body: %v", err)
  }

  err = json.Unmarshal(body, &userAuth)
  if err != nil {
    log.Fatalf("Failed to unmarshal the JSON data: %v", err)
  }

  fmt.Println("reached login Auth 1")
  //fmt.Println(userAuth)
  //fmt.Println(sqlURL)

  if len(userAuth) < 1 {
    return 0, ErrInvalidCredentials
    // also could be because RLS is turned on in Supabase. FIX FIX FIX
  }
  if email != userAuth[0].Email { // this func wont technically work till RLS issue is fixed. The above check temporarily fixes the RLS throwing a huge error. FIX FIX FIX
    fmt.Println("reached email comp, fix this, why returning incorrect/similar emails?") //needs to be fixed, potential problem with eq? Maybe need to use "like" or "ilike" instead.
    fmt.Println(userAuth)
    fmt.Println(email)
    return 0, ErrInvalidCredentials
  }

  hashedPassword = userAuth[0].HashedPassword
  
  //fmt.Println("past querying db")
  fmt.Println(userAuth)
  fmt.Println(password)
  //fmt.Println(email)

  err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
  if err != nil {
    if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
      return 0, ErrInvalidCredentials
    } else {
      return 0, err
    }
  }
  
  return id, nil
}

func (m *UserModel) Exists(id uuid.UUID) (bool, error) {
  var exists bool

  fmt.Println("reached exists")

  userUUID := uuid.UUID.String(id)

  var results []User
  err := m.DB.DB.From("users").Select(userUUID).Execute(&results)
  if err != nil {
    exists = false
    return exists, err
  }

  fmt.Println("reached exists 2")

  exists = true

  return exists, err
}

func (m *UserModel) Get(id uuid.UUID) (*User, error) {
  

  var userGet User
  err := m.DB.DB.From("users").Select("id, name, email, created").Eq("id", uuid.UUID.String(id)).Execute(&userGet)
  if err != nil {
    panic(err)
    return nil, ErrNoRecord
  }

  return &userGet, nil
}

func (m *UserModel) PasswordUpdate(id uuid.UUID, currentPassword, newPassword string) error {
  var currentHashedPassword []byte

  var userGet User
  err := m.DB.DB.From("users").Select("hashed_password").Eq("id", uuid.UUID.String(id)).Execute(&userGet)
  if err != nil {
    return nil
  }

  currentHashedPassword = userGet.HashedPassword

  err = bcrypt.CompareHashAndPassword(currentHashedPassword, []byte(currentPassword))
  if err != nil {
    if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
      return ErrInvalidCredentials
    } else {
      return err
    }
  }

  newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
  if err != nil {
    return err
  }

  row := userGet
  userGet.HashedPassword = newHashedPassword

  err = m.DB.DB.From("users").Update(row).Eq("id", uuid.UUID.String(id)).Execute(&userGet)
  return err
}

