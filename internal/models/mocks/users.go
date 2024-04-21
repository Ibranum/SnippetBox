package mocks

import (
  "snippetbox.evanread.net/internal/models"
  "time"
  "github.com/google/uuid"
  "fmt"
)

type UserModel struct{}

func (m *UserModel) Insert(name, email, password string) error {
  switch email {
    case "dupe@example.com":
      return models.ErrDuplicateEmail
    default:
      return nil
  }
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
  if email == "alice@example.com" && password == "pa$$word" {
    return 1, nil
  }

  return 0, models.ErrInvalidCredentials
}

func (m *UserModel) Exists(id uuid.UUID) (bool, error) {
  if len(id) < 15 {
    return false, nil
  }
  return true, nil
  
  //switch idT {
    //case 1:
      //return true, nil
    //default:
      //return false, nil
  //}
}

func (m *UserModel) Get(id uuid.UUID) (*models.User, error) {
  userUUID, err := uuid.Parse("e7eb2a56-876e-4f19-a6aa-a9155a68a810")
  if err != nil{
    fmt.Println("error here in users.go mocks 98765")
  }
  
  if id == userUUID {
    
    
    u := &models.User{
      ID: userUUID,
      Name: "Alice",
      Email: "alice@example.com",
      Created: time.Now(),
    }

    return u, nil
  }

  return nil, models.ErrNoRecord
}

func (m *UserModel) PasswordUpdate(id uuid.UUID, currentPassword, newPassword string) error {
    userUUID, err := uuid.Parse("e7eb2a56-876e-4f19-a6aa-a9155a68a810")
    if err != nil{
      fmt.Println("error here in users.go mocks 98765")
    }

  if id == userUUID {
    if currentPassword != "pa$$word" {
      return models.ErrInvalidCredentials
    }

    return nil
  }

  return models.ErrNoRecord
}