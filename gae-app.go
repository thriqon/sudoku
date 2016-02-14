//+build appengine

package sudoku

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"html/template"
)

func init() {
	http.HandleFunc("/solve", jsonHandler)
	http.HandleFunc("/", pageHandler)
}

func getSource(r *http.Request) (string, error) {
	if r.Header.Get("content-type") == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			return "", fmt.Errorf("Unable to parse form values")
		}
		return r.PostFormValue("sudoku"), nil
	}

	var bs []byte
	var err error
	if bs, err = ioutil.ReadAll(r.Body); err != nil {
		return "", fmt.Errorf("Unable to get data")
	}

	return string(bs), nil
}

func genericSolve(input string) (*Sudoku, error) {
	s, err := Parse(input)
	if err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}

	solved, err := s.Solve()
	if err != nil {
		return nil, fmt.Errorf("No solution found")
	}

	return &solved, nil
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		http.Error(w, "Invalid method, only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	source, err := getSource(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	solved, err := genericSolve(source)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if solved != nil {
		w.Header().Add("Content-type", "application/json")

		if err := json.NewEncoder(w).Encode(solved.AsInts()); err != nil {
			http.Error(w, "Error during encode", http.StatusInternalServerError)
		}
	}
}

type pageTemplateData struct {
	Err       error
	Cells     [9][9]uint8
	ShowCells bool
	Source    string
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		source, err := getSource(r)
		if err != nil {
			pageTemplate.Execute(w, pageTemplateData{Err: err, Source: source})
			return
		}

		solved, err := genericSolve(source)
		if err != nil {
			pageTemplate.Execute(w, pageTemplateData{Err: err, Source: source, ShowCells: true})
			return
		}
		pageTemplate.Execute(w, pageTemplateData{ShowCells: true, Cells: solved.AsInts(), Source: source})
		return
	}

	pageTemplate.Execute(w, pageTemplateData{Source: ""})
}

var pageTemplate = template.Must(template.New("page").Parse(`
<!doctype HTML>
<html>
  <head>
    <title>Sudoku Solver</title>
    <link href="https://maxcdn.bootstrapcdn.com/bootswatch/3.3.6/lumen/bootstrap.min.css" rel="stylesheet" integrity="sha256-QSktus/KATft+5BD6tKvAwzSxP75hHX0SrIjYto471M= sha512-787L1W8XyGQkqtvQigyUGnPxsRudYU2fEunzUP5c59Z3m4pKl1YaBGTcdhfxOfBvqTmJFmb6GDgm0iQRVWOvLQ==" crossorigin="anonymous">
		<style>
			td { text-align: center }
			td:nth-child(3), td:nth-child(6) { border-right: 1px solid #555 !important; }
			tr:nth-child(3) td, tr:nth-child(6) td { border-bottom: 1px solid #555 !important; }
		</style>
  </head>

  <body>
		<div class="container" style="margin-top: 5em">
			<div class="row">
				<div class="col-md-6">
					{{if .Err}}
						<div class="alert alert-danger">{{.Err}}</div>
					{{end}}

					<form action="/" method="POST">
						<div class="panel panel-default">
							<div class="panel-heading">
								<h2 class="panel-title">Input</h2>
							</div>
							<div class="panel-body">
								<textarea name="sudoku" class="form-control" rows="15">{{.Source}}</textarea>
							</div>
							<div class="panel-footer">
								<input type="submit" value="Compute" class="btn btn-primary">
							</div>
						</div>
					</form>
				</div>
				<div class="col-md-6">
					{{if .ShowCells}}
						<div class="panel panel-success">
							<div class="panel-heading">
								<h2 class="panel-title">Solution</h2>
							</div>
							<div class="panel-body">
								<table class="table">
								{{range .Cells}}
									<tr>
										{{range .}}
											<td>{{.}}</td>
										{{end}}
									</tr>
								{{end}}
								</table>
							</div>
						</div>
				{{end}}
			</div>
		</div>
  </body>
</html>
`))
