package actiontec

// Screen scraping functions for the Actiontec UI live here.

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

// Everything revolves around this context thing, which is basically just a
// logged in http.Client object and a host name or IP address for the router.
type Context struct {
	client  *http.Client
	address string
}

// Create a context, but don't log in.
func NewContext(address string) (*Context, error) {
	c := new(Context)

	// I don't actually think the UI uses cookies (it appears to be IP based:
	// once you log in, you're always logged in from that IP), but maybe that'll
	// eventually get fixed.
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	c.client = &http.Client{
		Jar: jar,
	}

	c.address = address

	return c, nil
}

// Gather line and overall stats and return them. If you haven't logged in via
// Login(), you'll have to before this function will work.
func (c *Context) GetStatus() (*Status, []LineStats, error) {
	var ls []LineStats

	// Get the first line, since we need it to figure out how many more lines
	// there are.
	status, err := c.statusForLine(0)
	if err != nil {
		return nil, nil, err
	}
	ls = append(ls, status.LineStats)

	// If there are more lines, then iterate over them and grab the line stats
	// for each one.
	for i := 1; i < len(status.LineRates); i++ {
		// Sadly, we have to regather the entire status, as there's global state in
		// the Actiontec UI. Although the modem status JS seems to indicate that
		// you can send a bonded line identifier to just get stats for that, it
		// lies: you _must_ reload the entire page before the refresh API (which is
		// as close as the router gets to a REST API) gives you data for the right
		// line.
		lineStatus, err := c.statusForLine(1)
		if err != nil {
			return nil, nil, err
		}

		ls = append(ls, lineStatus.LineStats)
	}

	return status, ls, nil
}

// Log into the UI.
func (c *Context) Login(username string, password string) error {
	resp, err := c.client.PostForm(c.url("/login.cgi"), url.Values{
		"inputUserName": []string{username},
		"inputPassword": []string{password},
		"nothankyou":    []string{"1"},
	})
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// After login, the UI redirects via JavaScript to the appropriate page. We
	// have to sniff the string that would show the error message.
	if strings.Contains(string(data), "msg=err") {
		return errors.New("User name or password incorrect")
	}

	return err
}

// Calls the refresh status page, which is a plain text API. See status.go for
// more details on how that's parsed.
func (c *Context) refreshStatus() (string, error) {
	resp, err := c.client.Get(c.url("/modemstatus_wanstatus_refresh.html"))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Calls the top level WAN status page, which is required to reset which line
// we care about.
func (c *Context) requestStatus(line int) error {
	_, err := c.client.PostForm(c.url("/modemstatus_wanstatus.cgi"), url.Values{
		"bondingLineNum": []string{strconv.Itoa(line)},
	})

	// Don't care about the content.
	return err
}

func (c *Context) statusForLine(line int) (*Status, error) {
	// See the comment in GetStatus() for why we always have to perform this two
	// step dance instead of just calling refreshStatus().
	if err := c.requestStatus(line); err != nil {
		return nil, err
	}

	data, err := c.refreshStatus()
	if err != nil {
		return nil, err
	}

	status, err := ParseStatus(data)
	if err != nil {
		return nil, err
	}

	return status, nil
}

// Convenience function to build an absolute URL from a relative one, given a
// context.
func (c *Context) url(rel string) string {
	return fmt.Sprintf("http://%s%s", c.address, rel)
}
