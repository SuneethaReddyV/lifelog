package ActivityLoggerMain

import (
	"appengine"
	//	"appengine/datastore"
	"appengine/user"
	//"errors"
	"fmt"
	"helpers"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//[TODO: need to document this function]
func handleRoot(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var dst []helpers.ActivityLog // dst to store the query results
	var f []helpers.Filter        // filter slice to store the number filters
	var OrderBy string

	// retrieve the Started activities
	f = []helpers.Filter{{"Status=", helpers.ActivityStatusStarted}}
	OrderBy = "-TimeStamp"
	dst = helpers.GetActivity(c, f, OrderBy)
	//............................................

	// retrieve the Paused activities
	// since the datastore query doesn't support filter on multiple values (like an IN operator or OR operator in SQL), we are doing it in 2 passes and storing the results in the same variable (i.e merging the result sets).
	f = []helpers.Filter{{"Status=", helpers.ActivityStatusPaused}}
	OrderBy = "-TimeStamp"
	dst2 := helpers.GetActivity(c, f, OrderBy)

	// merge the started and paused activities
	dst = append(dst, dst2...)

	t := template.Must(template.ParseFiles(
		"html/home.html",
		"html/_ActivityList.html",
		"html/_SvgButtons.html",
		"html/_mdl.html",
		"html/_footer.html",
		"html/_header.html",
	))

	// prepare the final data structure to pass to templates: add the user name to the activities list.
	a := helpers.HomePgData{user.Current(c).String(), dst}

	// execute the template whilo passing the required data to be rendered.
	if err := t.Execute(w, a); err != nil {
		panic(err)
	}
}

//[TODO: need to document this function]
func handleHistory(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var dst []helpers.ActivityLog

	dst = helpers.GetActivity(c, nil, "-TimeStamp")

	t := template.Must(template.ParseFiles(
		"html/history.html",
		"html/_ActivityList.html",
		"html/_SvgButtons.html",
		"html/_mdl.html",
		"html/_footer.html",
		"html/_header.html",
	))

	a := helpers.HomePgData{user.Current(c).String(), dst}

	if err := t.Execute(w, a); err != nil {
		panic(err)
	}

}

//[TODO: need to document this function]
func handleActivityAdd(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)

	if r.Method == "GET" {
		t := template.Must(template.ParseFiles(
			"html/AddActivity.html",
			"html/_SvgButtons.html",
			"html/_header.html",
			"html/_mdl.html",
		))
		if err := t.Execute(w, nil); err != nil {
			panic(err)
		}
	} else if r.Method == "POST" {

		c := appengine.NewContext(r)
		//r.ParseForm()
		//a.ActivityName = r.Form["activity"][0] // Note: if Form is used instead of FormValue, then it prequisite is to execute ParseForm() before. also Form returns an slice, where as FormValue returns just a string.

		a := helpers.ActivityLog{
			ActivityName: r.FormValue("activity"),
			TimeStamp:    time.Now(),
			StartTime:    time.Now(),
			Status:       helpers.ActivityStatusStarted,
			UserId:       user.Current(c).String(),
		}
		helpers.InsertActivity(c, a)

		http.Redirect(w, r, "/", http.StatusFound) // note: if there are any print commands (like fmt.Fprintln) before calling redirect operation, the redirect operation doesn't seams to be working
	}
}

// handleActivityUpdate() handles the logic for "/activity/update" route.
func handleActivityUpdate(w http.ResponseWriter, r *http.Request) {
	//[TODO: query param approach is not the effecient way to handle, as the parameters values in the url are visible to anyone,a nd it could pose security issues. so, need to explore the way in which we can pass the values as part of the HTTP header/body instead of URL]
	i := strings.Index(r.RequestURI, "?")       // since url.ParseQuery is not able to retrieve the first key (of the query string) correctly, find the position of ? in the url and
	qs := r.RequestURI[i+1 : len(r.RequestURI)] //  substring it and then

	m, _ := url.ParseQuery(qs) // parse it
	c := appengine.NewContext(r)

	err := helpers.UpdateActivity(c, m["ActivityName"][0], m["StartTime"][0], m["Status"][0], m["NewStatus"][0])

	if err != nil {
		http.Error(w, "Error while changing the status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}