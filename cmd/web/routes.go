package main

import (
  "net/http"
  "github.com/justinas/alice"
  "github.com/julienschmidt/httprouter"
  "snippetbox.evanread.net/ui"
  )

func (app *application) routes() http.Handler {

  //mux := http.NewServeMux()
  router := httprouter.New()

  router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    app.notFound(w)
  })

  //fileServer := http.FileServer(http.Dir("./ui/static"))
  //fileServer := http.FileServer(http.Dir("./ui/static"))
  fileServer := http.FileServer(http.FS(ui.Files))
  
  //mux.Handle("/static/", http.StripPrefix("/static", fileServer))
  //router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))
  router.Handler(http.MethodGet, "/static/*filepath", fileServer)

  router.HandlerFunc(http.MethodGet, "/ping", ping)

  //uncomment the below to enable CSRF protection. It is disabled because im doing all of my current testing with cURL.
  dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)
  //dynamic := alice.New(app.sessionManager.LoadAndSave, app.authenticate) //this one does not use CSRF

  //mux.HandleFunc("/", app.home)
  //mux.HandleFunc("/snippet/view", app.snippetView)
  //mux.HandleFunc("/snippet/create", app.snippetCreate)
  router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
  router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))
  router.Handler(http.MethodGet, "/about", dynamic.ThenFunc(app.about))
  

  router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))
  router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))
  router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
  router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))
  

  protected := dynamic.Append(app.requireAuthentication)
  
  router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.snippetCreate))
  router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.snippetCreatePost))
  router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))
  router.Handler(http.MethodGet, "/account/view", protected.ThenFunc(app.accountView))
  router.Handler(http.MethodGet, "/account/password/update", protected.ThenFunc(app.accountPasswordUpdate))
  router.Handler(http.MethodPost, "/account/password/update", protected.ThenFunc(app.accountPasswordUpdatePost))
  

  standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

  //return app.recoverPanic(app.logRequest(secureHeaders(mux)))
  return standard.Then(router)
}