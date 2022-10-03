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
	"tx/pkg/util"
)

type cognitoConfig struct {
	Domain       string
	ClientId     string
	ClientSecret string

	LoginUrl string // inferred

	Session *Session
}

var config = &cognitoConfig{}

type Session struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IdToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

var newSession bool

var Cmd = &cobra.Command{
	Use:     "cognito [COGNITO-DOMAIN] [CLIENT-ID] ([CLIENT_SECRET] if required)",
	Aliases: []string{"cognito"},
	Short:   "returns access token by default. see --help for more options",
	Run: func(cmd *cobra.Command, args []string) {
		err := viper.UnmarshalKey("cognito", config)
		if err != nil {
			klog.V(50).InfoS("Failed to parse config", "err", err)
			return
		}

		if newSession == false && config.Session != nil && config.Session.RefreshToken != "" {
			klog.V(50).InfoS("refresh token found. refreshing session...")

			err := resumeSession()
			if err != nil {
				klog.V(50).InfoS("failed to resume session", "err", err)
			} else {
				complete()
			}
		}

		// otherwise, init new session
		err = toOptions(args)
		if err != nil {
			cmd.Help()
			os.Exit(0)
		}

		util.CheckErr(healthCheck())

		wg := new(sync.WaitGroup)
		wg.Add(3)
		go util.RunCallbackServer(wg, loginCallback)
		go util.OpenBrowser(wg, config.LoginUrl)
		wg.Wait()

		complete()
	},
}

func complete() {
	viper.Set("cognito", config)
	util.CheckErr(viper.WriteConfig())

	fmt.Print(config.Session.AccessToken)
	os.Exit(0)
}

func toOptions(args []string) error {
	var err error

	klog.V(50).InfoS("clearing previous session")
	config.Session = &Session{}

	if config.Domain == "" || config.ClientId == "" || len(args) > 0 {
		if len(args) < 2 {
			return errors.New("see help for usage")

		}
		config.Domain = args[0]
		config.ClientId = args[1]
		if len(args) == 3 {
			config.ClientSecret = args[2]
		} else {
			config.ClientSecret = ""
		}

	}

	if !strings.HasPrefix(config.Domain, "http") {
		return errors.New("cognito domain should start with 'https://'")
	}

	config.LoginUrl = fmt.Sprintf(
		"%s/login?response_type=code&client_id=%s&redirect_uri=http://%s",
		config.Domain, config.ClientId, util.CallbackUrl(),
	)

	return err
}

func healthCheck() error {
	req, err := http.NewRequest("GET", config.LoginUrl, nil)
	if err != nil {
		return err
	}

	client := new(http.Client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if strings.Contains(req.URL.String(), "redirect_mismatch") {
			return fmt.Errorf("invalid redirect path: 'http://%s' is not added to allowed redirect urls in your user pool", util.CallbackUrl())
		} else if strings.Contains(req.URL.String(), "Client+does+not+exist") {
			return fmt.Errorf("invalid client id '%s'", config.ClientId)
		}
		return nil
	}

	resp, err := client.Do(req)
	if err == nil {
		if resp.StatusCode != http.StatusOK {
			return nil
		}
	} else {
		return err
	}

	util.CheckErr(resp.Body.Close())
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
	if config.ClientSecret != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", token))
	}

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
		return fmt.Errorf("failed to refresh session: status=%d body=%s", statusCode, string(body))
	}

	util.CheckErr(json.Unmarshal(body, &config.Session))
	return err
}

func loginCallback(params map[string][]string) error {
	klog.V(50).InfoS("params from callback", "params", params)
	data := url.Values{
		"grant_type":   {"authorization_code"},
		"client_id":    {config.ClientId},
		"code":         params["code"],
		"redirect_uri": {fmt.Sprintf("http://%s", util.CallbackUrl())},
	}

	statusCode, body, err := cognitoOAuthRequest(data)
	util.CheckErr(err)

	if statusCode != http.StatusOK {
		var obj struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(body, &obj)
		if obj.Error == "invalid_client" {
			fmt.Printf("does your client require a secret? if so, provide it as third arg. your secret may be invalid too\n")
		}

		return fmt.Errorf("failed to fetch tokens: statusCode=%d body=%s", statusCode, string(body))
	}

	util.CheckErr(json.Unmarshal(body, &config.Session))
	return nil
}

func init() {
	Cmd.Flags().BoolVar(&newSession, "new-session", false, "skips resuming previous session")
}
