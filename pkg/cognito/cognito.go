package cognito

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"token-exchange-cli/pkg/util"
)

type Session struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	IdToken      string        `json:"id_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    time.Duration `json:"expires_in"`
}

var session *Session

type Options struct {
	Domain       string
	ClientId     string
	ClientSecret string

	LoginUrl string // inferred
}

var options Options

func toOptions(args []string) {
	options = Options{
		Domain:       args[0],
		ClientId:     args[1],
		ClientSecret: args[2],
	}
	options.LoginUrl = fmt.Sprintf(
		"%s/login?response_type=code&client_id=%s&redirect_uri=http://%s",
		options.Domain, options.ClientId, util.CallbackUrl(),
	)
}

func resumeSession() error {
	// automatically attempt to get new access token
	return nil
}

func healthCheck() error {
	// check if the hosted UI is set up with the correct redirect uri
	return nil
}

func loginCallback(params map[string][]string) {
	// fetch tokens
	code := params["code"]

	token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", options.ClientId, options.ClientSecret)))
	klog.V(50).InfoS("encoded", "token", token)
	oauthUrl := fmt.Sprintf("%s/oauth2/token", options.Domain)

	data := url.Values{
		"grant_type":   {"authorization_code"},
		"client_id":    {options.ClientId},
		"code":         code,
		"redirect_uri": {fmt.Sprintf("http://%s", util.CallbackUrl())},
	}

	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodPost, oauthUrl, strings.NewReader(data.Encode()))
	req.Header.Add("User-Agent", "tx/dev")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", token))

	resp, _ := client.Do(req)

	if resp.StatusCode != http.StatusOK {
		util.CheckErr(errors.New("failed to fetch tokens"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.V(50).InfoS("problem", "err", err)
	}
	util.CheckErr(err)
	util.CheckErr(resp.Body.Close())

	klog.V(50).InfoS("Response from aws", "status", resp.Status, "body", string(body))

	util.CheckErr(json.Unmarshal(body, &session))

}

var Cmd = &cobra.Command{
	Use:     "cognito [COGNITO-DOMAIN] [CLIENT-ID] [CLIENT_SECRET]",
	Aliases: []string{"cognito"},
	Short:   "returns access token by default. see --help for more options",
	Args:    cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {

		toOptions(args)

		wg := new(sync.WaitGroup)
		wg.Add(2)
		go util.RunCallbackServer(wg, loginCallback)
		go util.OpenBrowser(wg, options.LoginUrl)
		wg.Wait()

		// print output
		fmt.Print(session.AccessToken)

	},
}
