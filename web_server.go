package main

import "github.com/digininja/vuLnDAP/config"
import "gopkg.in/russross/blackfriday.v2"
import "io/ioutil"
import (
	"bytes"
	"fmt"
	"gopkg.in/ldap.v2"
	"net/http"
	"reflect"
	"strings"
)

type WebServer struct {
	verbose    bool
	listenIP   string
	listenPort int
}

func NewWebServer(config config.Config) WebServer {
	webserver := WebServer{}
	webserver.listenIP = config.WebServer.ListenIP
	webserver.listenPort = config.WebServer.ListenPort
	return webserver
}

func (w *WebServer) getHTMLHeader(title string) string {
	header := `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd"><html>`
	header = header + "<head>"
	header = header + "<title>*TITLE*</title>"
	header = header + "<link rel=\"shortcut icon\" type=\"image/x-icon\" href=\"/favicon.ico\" />"
	header = header + "<link rel=\"apple-touch-icon\" href=\"/apple-touch-icon.png\" />"
	header = header + "<head><title>*TITLE*</title></head><body>"
	header = header + "</head><body>"
	header = strings.Replace(header, "*TITLE*", title, -1)
	return header
}

func (w *WebServer) getHTMLFooter() string {
	footer := `<hr /><p>Lab created by <a href='https://twitter.com/digininja'>Digininja</a>, for more information see <a href='https://digi.ninja/projects/vulndap.php'>the vuLnDAP project</a>.</p>

<!-- Global site tag (gtag.js) - Google Analytics -->
<script async src="https://www.googletagmanager.com/gtag/js?id=UA-7503551-6"></script>
<script src="/script"></script>

</body></html>`
	return footer
}

func (w *WebServer) handleRequestStock(writer http.ResponseWriter, r *http.Request) {
	form := w.getHTMLHeader("Stock Control")
	form = form + `
	<h1>Stock Control</h1>
	<p>
	<a href="/">Main Menu</a>
	</p>
	<p>Please select a category:</p>
	<ul>
		<li><a href="/fruit_or_veg?objectClass=fruits">Fruit</a></li>
		<li><a href="/fruit_or_veg?objectClass=vegetables">Veg</a></li>
	</ul>
	`
	form = form + w.getHTMLFooter()
	fmt.Fprintf(writer, form)
}

func (w *WebServer) handleRequestItem(writer http.ResponseWriter, r *http.Request) {
	var filter string
	var attributes []string
	var disp string
	var cn string

	r.ParseForm() // parse arguments, you have to call this by yourself

	if val, ok := r.Form["cn"]; ok {
		cn = val[0]
		clientLogger.Debug(fmt.Sprintf("cn: %s", cn))
	} else {
		clientLogger.Debug("No choice")
		http.Redirect(writer, r, "/stock", http.StatusFound)
		return
	}

	if val, ok := r.Form["disp"]; ok {
		disp = val[0]
		clientLogger.Debug(fmt.Sprintf("attributes: %s", disp))
	} else {
		disp = "stock,cn,description"
	}

	filter = fmt.Sprintf("(cn=%s)", cn)
	disp = disp + ",objectClass"
	attributes = strings.Split(disp, ",")

	fmt.Fprintf(writer, w.getHTMLHeader(cn))

	if filter != "" && len(attributes) > 0 {
		clientLogger.Debug("Params passed")

		fmt.Fprintf(writer, "<!-- Search filter: %s -->", filter)
		sr, err := VulndapConnection.search(writer, filter, attributes, 0)
		if err != nil {
			fmt.Fprintf(writer, "<p>Search failed:<br />")
			fmt.Fprintf(writer, "%s</p>", err)
			clientLogger.Info("Search failed, possibly invalid search filter")
			if w.verbose {
				clientLogger.Debug(err)
			}
			return
		}
		clientLogger.Info(fmt.Sprintf("Returned %d result(s)", len(sr.Entries)))

		if len(sr.Entries) == 0 {
			fmt.Fprintf(writer, "<p>No results found</p>")
			return
		}

		entry := sr.Entries[0]

		// fmt.Fprintf(writer, "entry type: %s<br />", reflect.TypeOf(entry))
		clientLogger.Debug(fmt.Sprintf("DN: %s", entry.DN))

		cn := ""
		var buffer bytes.Buffer

		// Default to fruits so back goes back to something
		objectClass := "fruits"

		for _, attr := range entry.Attributes {
			// clientLogger.Info("attr type: %s<br />", reflect.TypeOf(attr))
			switch attr.Name {
			case "cn":
				cn = attr.Values[0]
				clientLogger.Debug(fmt.Sprintf("cn: %s", cn))
			case "objectClass":
				objectClass = attr.Values[0]
				clientLogger.Debug(fmt.Sprintf("objectClass: %s", objectClass))
			default:
				clientLogger.Debug(fmt.Sprintf("%s: %s", attr.Name, strings.Join(attr.Values, ", ")))
				buffer.WriteString(fmt.Sprintf("<dt>%s</dt><dd>%s</dd>", attr.Name, strings.Join(attr.Values, ", ")))
			}

		}

		fmt.Fprintf(writer, "<h1>%s</h1>", cn)
		fmt.Fprintf(writer, "<p><a href='/'>Main Menu</a></p>")

		fmt.Fprintf(writer, "<dl>%s</dl>", buffer.String())
		fmt.Fprintf(writer, "<p><a href='/fruit_or_veg?objectClass=%s'>&laquo; Back</a></p>", objectClass)

		fmt.Fprintf(writer, "<br />")
	}
	fmt.Fprintf(writer, w.getHTMLFooter())
}

func (w *WebServer) handleRequestFruitOrVeg(writer http.ResponseWriter, r *http.Request) {
	r.ParseForm() // parse arguments, you have to call this by yourself
	/*
		fmt.Println(r.Form) // print form information in server side
		fmt.Println("path", r.URL.Path)
		fmt.Println("scheme", r.URL.Scheme)
		fmt.Println(r.Form["url_long"])
	*/

	var filter string
	var attributes []string
	var fruit_or_veg string

	fruit_or_veg = ""
	if val, ok := r.Form["objectClass"]; ok {
		fruit_or_veg = val[0]
		clientLogger.Info(fmt.Sprintf("They chose: %s", fruit_or_veg))
	} else {
		clientLogger.Info("No choice, redirecting")
		http.Redirect(writer, r, "/stock", http.StatusFound)
	}

	form := w.getHTMLHeader("XXX")
	form = form + `
	<h1>XXX</h1>
	<p>
	<a href="/">Main Menu</a>
	</p>
	`

	fruit_or_veg = strings.Replace(fruit_or_veg, "(", "", -1)
	fruit_or_veg = strings.Replace(fruit_or_veg, ")", "", -1)

	if fruit_or_veg != "" {
		form = strings.Replace(form, "XXX", fruit_or_veg, -1)
		fmt.Fprintf(writer, form)
		filter = fmt.Sprintf("(objectClass=%s)", ldap.EscapeFilter(fruit_or_veg))
		attributes = []string{"cn", "description"}
		if filter != "" && len(attributes) > 0 {
			clientLogger.Debug("Params passed")

			fmt.Fprintf(writer, "<!-- Search filter: %s -->", filter)
			sr, err := VulndapConnection.search(writer, filter, attributes, 0)
			if err != nil {
				fmt.Fprintf(writer, "<p>Search failed:<br />")
				fmt.Fprintf(writer, "%s</p>", err)
				clientLogger.Info("Search failed, possibly invalid search filter")
				if w.verbose {
					clientLogger.Error(err)
				}
				return
			}
			message := (fmt.Sprintf("Returned %d result(s)", len(sr.Entries)))
			clientLogger.Info(message)
			fmt.Fprintf(writer, message+"<br />")

			// sr type: ldap.SearchResult
			// entry type: *ldap.Entry - https://godoc.org/gopkg.in/ldap.v2#Entry
			// attr type: *ldap.EntryAttribute - https://godoc.org/gopkg.in/ldap.v2#EntryAttribute

			// clientLogger.Info("sr type: %s<br />", reflect.TypeOf(sr))
			for _, entry := range sr.Entries {
				// fmt.Fprintf(writer, "entry type: %s<br />", reflect.TypeOf(entry))
				clientLogger.Debug(fmt.Sprintf("DN: %s", entry.DN))
				message = fmt.Sprintf("\n<!-- DN: %s-->\n", entry.DN)
				fmt.Fprintf(writer, message)

				cn := ""
				description := ""
				for _, attr := range entry.Attributes {
					clientLogger.Debug(fmt.Sprintf("attr type: %s<br />", reflect.TypeOf(attr)))
					clientLogger.Debug(fmt.Sprintf("%s: %s", attr.Name, strings.Join(attr.Values, ", ")))
					if attr.Name == "cn" {
						cn = attr.Values[0]
					}
					if attr.Name == "description" {
						description = attr.Values[0]
					}
				}
				fmt.Fprintf(writer, "<h2>%s</h2>", cn)
				fmt.Fprintf(writer, "<p>%s</p>", description)
				fmt.Fprintf(writer, "<p><a href='/item?cn=%s&disp=stock,description,cn'>More Info</a></p>", cn)

			}
		}
	} else {
		form = strings.Replace(form, "XXX", "No Results", -1)
		fmt.Fprintf(writer, form)
	}
	fmt.Fprintf(writer, "<p><a href='/stock'>&laquo; Back</a></p>")
	fmt.Fprintf(writer, w.getHTMLFooter())
}

func (w *WebServer) handleRequestRaw(writer http.ResponseWriter, r *http.Request) {
	form := w.getHTMLHeader("Raw Queries")
	form = form + `
	<h1>Raw Queries</h1>
	<p>
	<a href="/">Main Menu</a>
	</p>
	<p>Enter a raw query.</p>
	<form method="get">
		Filter: <input type="text" name="filter" id="filter" value="**FILTER**" /><br />
		Attributes: <input type="text" name="attributes" id="filter" value="**ATTRIBUTES**" /><br />
		<input type="submit" value="Search" name="search" />
	</form>
	`
	r.ParseForm() // parse arguments, you have to call this by yourself
	/*
		fmt.Println(r.Form) // print form information in server side
		fmt.Println("path", r.URL.Path)
		fmt.Println("scheme", r.URL.Scheme)
		fmt.Println(r.Form["url_long"])
	*/

	var filter string
	var attributes []string
	var attr_string string

	if val, ok := r.Form["attributes"]; ok {
		attr_string = val[0]
		clientLogger.Info(fmt.Sprintf("Attributes: %s", attr_string))
	} else {
		clientLogger.Info("Missing attributes")
	}

	if val, ok := r.Form["filter"]; ok {
		filter = val[0]
		clientLogger.Info(fmt.Sprintf("Filter: %s", filter))
	} else {
		clientLogger.Info("Missing Filter")
	}

	form = strings.Replace(form, "**FILTER**", filter, -1)
	form = strings.Replace(form, "**ATTRIBUTES**", attr_string, -1)
	fmt.Fprintf(writer, form)

	attributes = strings.Split(attr_string, ",")
	if filter != "" && len(attributes) > 0 {
		clientLogger.Debug("Params passed")

		sr, err := VulndapConnection.search(writer, filter, attributes, 0)

		if err != nil {
			fmt.Fprintf(writer, "<p>Search failed:<br />")
			fmt.Fprintf(writer, "%s</p>", err)
			clientLogger.Info("Search failed, possibly invalid search filter")
			if w.verbose {
				clientLogger.Error(err)
			}
			return
		}
		message := (fmt.Sprintf("Returned %d result(s)", len(sr.Entries)))
		clientLogger.Info(message)
		fmt.Fprintf(writer, "<p>%s</p>", message)

		// sr type: ldap.SearchResult
		// entry type: *ldap.Entry - https://godoc.org/gopkg.in/ldap.v2#Entry
		// attr type: *ldap.EntryAttribute - https://godoc.org/gopkg.in/ldap.v2#EntryAttribute

		// clientLogger.Info("sr type: %s<br />", reflect.TypeOf(sr))
		for _, entry := range sr.Entries {
			// fmt.Fprintf(writer, "entry type: %s<br />", reflect.TypeOf(entry))
			clientLogger.Debug(fmt.Sprintf("DN: %s", entry.DN))
			message = fmt.Sprintf("<h2>DN: %s</h2>\n", entry.DN)
			fmt.Fprintf(writer, message)

			for _, attr := range entry.Attributes {
				clientLogger.Debug(fmt.Sprintf("attr type: %s", reflect.TypeOf(attr)))
				clientLogger.Debug(fmt.Sprintf("%s: %s", attr.Name, strings.Join(attr.Values, ", ")))
				fmt.Fprintf(writer, "<p>%s:%s</p>", attr.Name, strings.Join(attr.Values, ", "))
			}
		}
	}
	fmt.Fprintf(writer, w.getHTMLFooter())
}

/*
func (w *WebServer) handleRequestLogin(writer http.ResponseWriter, r *http.Request) {
	form := `
	<h1>Welcome to vuLnDAP!</h1>
	<p>
	<a href="/">Main Menu</a>
	</p>
	<p>The UID tesla will get you a search result</p>
	<form method="get">
		UID: <input type="text" name="uid" id="uid" value="**UID**" /><br />
		Given: <input type="text" name="given" id="given" value="**GIVEN**" /><br />

		<input type="submit" value="Search" name="search" />
	</form>
	`
	r.ParseForm() // parse arguments, you have to call this by yourself
	/*
		fmt.Println(r.Form) // print form information in server side
		fmt.Println("path", r.URL.Path)
		fmt.Println("scheme", r.URL.Scheme)
		fmt.Println(r.Form["url_long"])
	* /

	var filter string
	var uid string
	var given string

	attributes := []string{"*"}

	if val, ok := r.Form["given"]; ok {
		given = val[0]
		clientLogger.Info(fmt.Sprintf("Given: %s", given))
	} else {
		clientLogger.Println("Missing Given")
	}

	form = strings.Replace(form, "**GIVEN**", given, -1)

	if val, ok := r.Form["uid"]; ok {
		uid = val[0]
		clientLogger.Info(fmt.Sprintf("UID: %s", uid))
	} else {
		clientLogger.Println("Missing UID")
	}

	form = strings.Replace(form, "**UID**", uid, -1)
	fmt.Fprintf(writer, form)

	if uid != "" && len(attributes) > 0 {
		clientLogger.Println("Params passed")

		filter = fmt.Sprintf("(&(uid=%s)()", uid)
		filter = fmt.Sprintf("(&(uid=%s)(givenName=%s))", uid, given)

		VulndapConnection.search(writer, filter, attributes, 1)
	}
}

func (w *WebServer) handleRequestHarder(writer http.ResponseWriter, r *http.Request) {
	form := `
	<h1>Welcome to vuLnDAP!</h1>
	<p>
	<a href="/">&laquo; Menu</a>
	</p>
	<p>The UID tesla will get you a search result</p>
	<form method="get">
		UID: <input type="text" name="uid" id="uid" value="**UID**" /><br />
		sn: <input checked="checked" type="checkbox" value="sn" name="attributes[]" /><br />
		givenName: <input checked="checked" type="checkbox" value="givenName" name="attributes[]" /><br />
		cn: <input checked="checked" type="checkbox" value="cn" name="attributes[]" /><br />
		displayName: <input checked="checked" type="checkbox" value="displayName" name="attributes[]" /><br />
		uidNumber: <input checked="checked" type="checkbox" value="uidNumber" name="attributes[]" /><br />
		gidNumber: <input checked="checked" type="checkbox" value="gidNumber" name="attributes[]" /><br />
		gecos: <input checked="checked" type="checkbox" value="gecos" name="attributes[]" /><br />
		loginShell: <input checked="checked" type="checkbox" value="loginShell" name="attributes[]" /><br />
		homeDirectory: <input checked="checked" type="checkbox" value="homeDirectory" name="attributes[]" /><br />

		<input type="submit" value="Search" name="search" />
	</form>
	`
	r.ParseForm() // parse arguments, you have to call this by yourself
	/*
		fmt.Println(r.Form) // print form information in server side
		fmt.Println("path", r.URL.Path)
		fmt.Println("scheme", r.URL.Scheme)
		fmt.Println(r.Form["url_long"])
	* /

	var filter string
	var attributes []string
	var uid string

	if val, ok := r.Form["attributes[]"]; ok {
		attributes = val
		clientLogger.Info(fmt.Sprintf("Attributes: %s", strings.Join(attributes, ",")))
	} else {
		clientLogger.Println("Missing attributes")
	}

	if val, ok := r.Form["uid"]; ok {
		uid = val[0]
		clientLogger.Info(fmt.Sprintf("UID: %s", uid))
	} else {
		clientLogger.Println("Missing UID")
	}

	form = strings.Replace(form, "**UID**", uid, -1)
	fmt.Fprintf(writer, form)

	if uid != "" && len(attributes) > 0 {
		clientLogger.Println("Params passed")

		filter = fmt.Sprintf("(&(uid=%s)(objectClass=inetOrgPerson))", uid)

		VulndapConnection.search(writer, filter, attributes, 0)
	}
}

func (w *WebServer) handleRequestBasic(writer http.ResponseWriter, r *http.Request) {
	form := `
	<h1>Welcome to vuLnDAP!</h1>
	<script>
		function do_search() {
			var uid_val = document.getElementById("uid").value;
			var filter_ele = document.getElementById("filter")

			filter_ele.value = "(&(objectClass=inetOrgPerson)(uid=" + uid_val + "))";
			return true;
		}
	</script>
	<p>
	<a href="/">&laquo; Menu</a>
	</p>
	<p>The UID tesla will get you a search result</p>
	<form method="get">
		<input type="hidden" value="" id="filter" name="filter" />
		UID: <input type="text" name="uid" id="uid" value="**UID**" /><br />
		sn: <input checked="checked" type="checkbox" value="sn" name="attributes[]" /><br />
		givenName: <input checked="checked" type="checkbox" value="givenName" name="attributes[]" /><br />
		cn: <input checked="checked" type="checkbox" value="cn" name="attributes[]" /><br />
		displayName: <input checked="checked" type="checkbox" value="displayName" name="attributes[]" /><br />
		uidNumber: <input checked="checked" type="checkbox" value="uidNumber" name="attributes[]" /><br />
		gidNumber: <input checked="checked" type="checkbox" value="gidNumber" name="attributes[]" /><br />
		gecos: <input checked="checked" type="checkbox" value="gecos" name="attributes[]" /><br />
		loginShell: <input checked="checked" type="checkbox" value="loginShell" name="attributes[]" /><br />
		homeDirectory: <input checked="checked" type="checkbox" value="homeDirectory" name="attributes[]" /><br />

		<input onclick="return do_search()" type="submit" value="Search" name="search" />
	</form>
	`
	r.ParseForm() // parse arguments, you have to call this by yourself
	/*
		fmt.Println(r.Form) // print form information in server side
		fmt.Println("path", r.URL.Path)
		fmt.Println("scheme", r.URL.Scheme)
		fmt.Println(r.Form["url_long"])
	* /

	var filter string
	var uid string
	var attributes []string

	if val, ok := r.Form["attributes[]"]; ok {
		attributes = val
		clientLogger.Info(fmt.Sprintf("Attributes: %s", strings.Join(attributes, ",")))
	} else {
		clientLogger.Println("Missing attributes")
	}

	if val, ok := r.Form["filter"]; ok {
		filter = val[0]
		clientLogger.Info(fmt.Sprintf("Filter: %s", filter))
	} else {
		clientLogger.Println("Missing filter")
	}

	// This is only needed to reset the input field when the form is redrawn
	if val, ok := r.Form["uid"]; ok {
		uid = val[0]
		clientLogger.Info(fmt.Sprintf("UID: %s", uid))
	} else {
		clientLogger.Println("Missing uid")
	}

	form = strings.Replace(form, "**UID**", uid, -1)
	fmt.Fprintf(writer, form)

	if filter != "" && len(attributes) > 0 {
		VulndapConnection.search(writer, filter, attributes, 0)
	}

	/*
		for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}
	* /
}

*/

func (w *WebServer) handleRequestScript(writer http.ResponseWriter, r *http.Request) {
	writer.Header().Set("Content-Type", "application/javascript")

	form := `window.dataLayer = window.dataLayer || [];
function gtag(){dataLayer.push(arguments);}
gtag('js', new Date());

gtag('config', 'UA-7503551-6');`
	fmt.Fprintf(writer, form)
}

func (w *WebServer) handleRequestIndex(writer http.ResponseWriter, r *http.Request) {
	form := w.getHTMLHeader("vuLnDAP")
	form = form + `
	<h1>vuLnDAP</h1>
	<p>
	The challenge... Useless Inc. is running its stock control system on the same LDAP server the network uses for authentication, abuse this relationship to find details of the system users. As a bonus, it is known that the admins store SSH keys in the database, see if you can find anything interesting in there.</p>
	<ul>
		<li><a href="/stock">Stock Control</a></li>
	</ul>
	<p>Here are some extra features which can help if you get stuck.</p>
	<ul>
		<li><a href="/raw">Enter raw commands</a></li>
		<li><a href="/resources">Some useful resources</a></li>
	</ul>
	<p>If you get really stuck I've published a <a href="https://digi.ninja/blog/vulndap_walkthrough.php">full walkthrough</a>.</p>
	`
	fmt.Fprintf(writer, form)
	fmt.Fprintf(writer, w.getHTMLFooter())
}

func (w *WebServer) handleRequestResources(writer http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(writer, w.getHTMLHeader("Resources"))
	markdownResources, err := ioutil.ReadFile("resources.md")
	if err != nil {
		clientLogger.Println("Unable to open resources.md")
		fmt.Fprintf(writer, "Unable to open resources.md")
		return
	}

	htmlResources := blackfriday.Run(markdownResources)

	menu := `
	<p>
	<a href="/">Main Menu</a>
	</p>
	`
	fmt.Fprintf(writer, menu)
	fmt.Fprintf(writer, string(htmlResources))
	fmt.Fprintf(writer, w.getHTMLFooter())
}

func appleHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "resources/apple-touch-icon.png")
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "resources/favicon.ico")
}

func (w *WebServer) startWebApp() {
	http.HandleFunc("/", w.handleRequestIndex) // set route
	http.HandleFunc("/raw", w.handleRequestRaw)
	http.HandleFunc("/stock", w.handleRequestStock)
	http.HandleFunc("/fruit_or_veg", w.handleRequestFruitOrVeg)
	http.HandleFunc("/item", w.handleRequestItem)
	http.HandleFunc("/resources", w.handleRequestResources)
	http.HandleFunc("/script", w.handleRequestScript)
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/apple-touch-icon.png", appleHandler)

	/*
		not currently used
		http.HandleFunc("/basic", w.handleRequestBasic)
		http.HandleFunc("/harder", w.handleRequestHarder)
		http.HandleFunc("/login", w.handleRequestLogin)
	*/
	clientLogger.Info(fmt.Sprintf("Starting web server on %s:%d", w.listenIP, w.listenPort))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", w.listenIP, w.listenPort), nil) // set listen IP and port
	if err != nil {
		clientLogger.Fatal(fmt.Sprintf("ListenAndServe: %s", err))
	}
}
