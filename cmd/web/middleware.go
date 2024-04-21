package main

import (
  "net/http"
  "fmt"
  "github.com/justinas/nosurf"
  "context"
  "github.com/google/uuid"
)

func secureHeaders(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Security-Policy",
                  "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")

    w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.Header().Set("X-Frame-Options", "deny")
    w.Header().Set("X-XSS-Protection", "0")

    next.ServeHTTP(w, r)
  })
}

func (app *application) logRequest(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

    next.ServeHTTP(w, r)
  })
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    defer func() {
      if err := recover(); err != nil {
        w.Header().Set("Connection", "close")
        app.serverError(w, fmt.Errorf("%s", err))
      }
    }()

    next.ServeHTTP(w, r)
  })
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if !app.isAuthenticated(r) {
      app.sessionManager.Put(r.Context(), "redirectPathAfterLogin", r.URL.Path)
      
      http.Redirect(w, r, "/user/login", http.StatusSeeOther)
      return
    }

    w.Header().Add("Cache-Control", "no-store")

    next.ServeHTTP(w, r)
  })
}

func noSurf(next http.Handler) http.Handler {
  csrfHandler := nosurf.New(next)
  csrfHandler.SetBaseCookie(http.Cookie{
    HttpOnly: true,
    Path: "/",
    Secure: true,
  })

  return csrfHandler
}

func (app *application) authenticate(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    //fmt.Println("reached auth 1")
    
    id := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
    if id == "" { //this might be the incorrect way to check this
      next.ServeHTTP(w, r)
      return
    }

    UUID, err := uuid.Parse(id)

    //fmt.Println("reached auth 2")

    exists, err := app.users.Exists(UUID)
    if err != nil {
      app.serverError(w, err)
      return
    }

    //fmt.Println("reached auth 3")

    if exists {
      ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
      r = r.WithContext(ctx)
    }

    //fmt.Println("reached auth 4")

    next.ServeHTTP(w, r)
  })
}