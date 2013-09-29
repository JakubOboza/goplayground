package main

import(
  "fmt"
  "io/ioutil"
  "net/http"
  "html/template"
  "path"
  "regexp"
  "errors"
)

type Page struct {
  Title string
  Body string
}

// var templates = CacheTemplates("templates")

// func CacheTemplates(dirname string) *template.Template {
//   files, err := ioutil.ReadDir("templates")
//   if(err != nil){
//     fmt.Println(fmt.Sprintf("Could not find TEMPLATES DIRECTORY '%s'", dirname))
//     panic(1)
//   }
//   fileNames := make([]string, len(files))
//   for i := 0; i < len(files); i++ {
//     fileNames[i] = path.Join(dirname,files[i].Name())
//   }
//   return template.Must(template.ParseFiles(fileNames...))
// }

//var templates = template.Must(template.ParseFiles("templates/edit.html", "templates/show.html"))

var titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")

const lenPath = len("/wiki/")

func (p *Page) save() error {
  filename := p.Title + ".text"
  return ioutil.WriteFile(filename, []byte(p.Body), 0600)
}

func (p *Page) Render(w http.ResponseWriter, template_name string) {
  t, err := template.ParseFiles(path.Join("templates/",template_name + ".html"))
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  t.Execute(w, p)

  //templates.ExecuteTemplate(w, path.Join("templates",template_name + ".html"), p)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    title, e := getTitle(w, r)
    if (e != nil){
      return
    }
    fn(w,r, title)
  }
}

func loadPage(title string) (*Page, error) {
  filename := title + ".text"
  body, error := ioutil.ReadFile(filename)
  if error != nil {
    return nil, error
  }
  return &Page{Title: title, Body: string(body)}, nil
}

func getTitle(w http.ResponseWriter, r *http.Request) (title string, err error) {
  if(len(r.URL.Path) < lenPath ){
    http.NotFound(w, r)
    err = errors.New("Invalid Page Title")
    return
  }
  title = r.URL.Path[lenPath:]
  if !titleValidator.MatchString(title){
    http.NotFound(w, r)
    err = errors.New("Invalid Page Title")
    return
  }
  return
}

func pageHandler(w http.ResponseWriter, r *http.Request, title string){
  p, err := loadPage(title)
  if (err != nil){
    http.Redirect(w, r, "/edit/" + title, http.StatusFound)
  }
  p.Render(w, "show")
}

func editHandler(w http.ResponseWriter, r *http.Request, title string){
  p, err := loadPage(title)
  if (err != nil){
    p = &Page{Title: title, Body: ""}
  }
  p.Render(w, "edit")
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string){
  body := r.FormValue("body")
  p := &Page{Title: title, Body: body}
  p.save()
  http.Redirect(w, r, "/wiki/" + title, http.StatusFound)
}


func main(){

  port := ":8080"
  fmt.Println(fmt.Sprintf("Serving on %s", port))
  http.HandleFunc("/wiki/", makeHandler(pageHandler))
  http.HandleFunc("/edit/", makeHandler(editHandler))
  http.HandleFunc("/save/", makeHandler(saveHandler))
  http.ListenAndServe(port, nil)
}