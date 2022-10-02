package cognito

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"token-exchange-cli/pkg/util"
)

type ViperConfig struct {
	Domain       string
	ClientId     string
	ClientSecret string

	LoginUrl string // inferred

	Session *Session
}

var config = &ViperConfig{}

type Session struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IdToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func toOptions(args []string) {
	config.Domain = args[0]
	config.ClientId = args[1]
	config.ClientSecret = args[2]

	config.LoginUrl = fmt.Sprintf(
		"%s/login?response_type=code&client_id=%s&redirect_uri=http://%s",
		config.Domain, config.ClientId, util.CallbackUrl(),
	)
}

func healthCheck() error {
	// check if the hosted UI is set up with the correct redirect uri
	return nil
}

func cognitoOAuthRequest(formData url.Values) (int, []byte, error) {
	var err error

	oauthUrl := fmt.Sprintf("%s/oauth2/token", config.Domain)

	token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", config.ClientId, config.ClientSecret)))
	klog.V(100).InfoS("encoded client_id:client_secret", "token", token)

	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodPost, oauthUrl, strings.NewReader(formData.Encode()))
	req.Header.Add("User-Agent", "tx/dev")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", token))

	resp, _ := client.Do(req)

	body, err := io.ReadAll(resp.Body)
	util.CheckErr(err)
	util.CheckErr(resp.Body.Close())

	klog.V(100).InfoS("cognito oauth2 response", "status", resp.Status, "body", string(body))

	return resp.StatusCode, body, err
}

func resumeSession() error {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {config.ClientId},
		"refresh_token": {config.Session.RefreshToken},
	}

	statusCode, body, err := cognitoOAuthRequest(data)

	if statusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("failed to refresh session: status=%d body=%s", statusCode, string(body)))
	}

	util.CheckErr(json.Unmarshal(body, config.Session))
	return err
}

func loginCallback(params map[string][]string) {
	data := url.Values{
		"grant_type":   {"authorization_code"},
		"client_id":    {config.ClientId},
		"code":         params["code"],
		"redirect_uri": {fmt.Sprintf("http://%s", util.CallbackUrl())},
	}

	statusCode, body, err := cognitoOAuthRequest(data)
	util.CheckErr(err)

	if statusCode != http.StatusOK {
		klog.Fatal("failed to fetch tokens:", "status", statusCode, "body", string(body))
	}

	util.CheckErr(json.Unmarshal(body, config.Session))

}

func complete() {
	viper.Set("cognito", config)
	fmt.Print(config.Session.AccessToken)
	os.Exit(0)
}

var Cmd = &cobra.Command{
	Use:     "cognito [COGNITO-DOMAIN] [CLIENT-ID] [CLIENT_SECRET]",
	Aliases: []string{"cognito"},
	Short:   "returns access token by default. see --help for more options",
	//Args:    cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		err := viper.UnmarshalKey("cognito", config)
		if err != nil {
			klog.V(50).InfoS("Failed to parse config", "err", err)
			return
		}
		klog.V(100).InfoS("current config", "domain", config.Domain, "session", config.Session)

		if config.Session.RefreshToken != "" {
			klog.V(50).InfoS("refresh token found. refreshing session...")

			err := resumeSession()
			if err != nil {
				klog.V(50).InfoS("failed to resume session", "err", err)
			} else {
				complete()
			}
		}

		// otherwise init new session; check for args
		toOptions(args)
		wg := new(sync.WaitGroup)
		wg.Add(2)
		go util.RunCallbackServer(wg, loginCallback)
		go util.OpenBrowser(wg, config.LoginUrl)
		wg.Wait()

		complete()
	},
}
