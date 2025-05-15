package cmd

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/microcks/microcks-cli/pkg/util/rand"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/term"
)

func NewLoginCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var (
		ctxName          string
		username         string
		password         string
		sso              bool
		ssoLaunchBrowser bool
		ssoProt          int
	)
	loginCmd := &cobra.Command{

		Use:   "login SERVER",
		Short: "Login into Microcks instance",
		Long:  "Login into Microcks instance",
		Example: `microcks login http://locahost:8080

# Provide name to your logged in context (Defautl context name is server name)
microcks login http://localhost:8080 --name

# Provide username and password as flags
microcks login http://localhost:8080 --username --password

# Perform SSO login
microcks login http://localhost:8080 --sso

# Change port callback server for SSO login
microcks login http://localhost:8080 --sso --sso-port

# Get OAuth URI instead of getting redirect to browser for SSO login
microcks login http://localhost:8080 --sso --sso-launch-browser=false
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			var server string

			//Chekc if server name is provided or not
			if len(args) != 1 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			server = args[0]
			mc := connectors.NewMicrocksClient(server)
			keycloakUrl, err := mc.GetKeycloakURL()

			if err != nil {
				log.Fatal(err)
			}

			if ctxName == "" {
				ctxName = server
			}

			var authToken = ""
			var refreshToken = ""

			//Initialze Auth struct
			authCfg := config.Auth{
				Server:       server,
				ClientId:     "",
				ClientSecret: "",
			}

			configFile, err := config.DefaultLocalConfigPath()
			errors.CheckError(err)
			localConfig, err := config.ReadLocalConfig(configFile)
			errors.CheckError(err)

			if localConfig == nil {
				localConfig = &config.LocalConfig{}
			}

			if keycloakUrl == "null" {
				localConfig.UpsertServer(config.Server{
					Server:          server,
					InsecureTLS:     true,
					KeycloackEnable: false,
				})
				fmt.Print("No login required...\n")
			} else {
				if !sso {
					//Chek for the enviroment variables
					clientID := os.Getenv("MICROCKS_CLIENT_ID")
					clientSecret := os.Getenv("MICROCKS_CLIENT_SECRET")

					if clientID == "" || clientSecret == "" {
						fmt.Printf("Please Set 'MICROCKS_CLIENT_ID' & 'MICROCKS_CLIENT_SECRET' to perform password login\n")
						os.Exit(1)
					}
					//Perform login and retrive tokens
					authToken, refreshToken = passwordLogin(keycloakUrl, clientID, clientSecret, username, password)
					authCfg.ClientId = clientID
					authCfg.ClientSecret = clientSecret
				} else {
					httpClient := mc.HttpClient()
					ctx = oidc.ClientContext(ctx, httpClient)
					kc := connectors.NewKeycloakClient(keycloakUrl, "", "")
					oauth2conf, err := kc.GetOIDCConfig()
					errors.CheckError(err)
					authToken, refreshToken = oauth2login(ctx, ssoProt, oauth2conf, ssoLaunchBrowser)
					authCfg.ClientId = "microcks-app-js"
				}

				parser := jwt.NewParser(jwt.WithoutClaimsValidation())
				claims := jwt.MapClaims{}
				_, _, err = parser.ParseUnverified(authToken, &claims)

				if err != nil {
					fmt.Println(err)
				}

				em := StringField(claims, "preferred_username")
				fmt.Printf("'%s' logged in successfully\n", em)

				localConfig.UpsertServer(config.Server{
					Server:          server,
					InsecureTLS:     true,
					KeycloackEnable: true,
				})
			}

			localConfig.UpserAuth(authCfg)

			localConfig.UpsertUser(config.User{
				Name:         server,
				AuthToken:    authToken,
				RefreshToken: refreshToken,
			})

			localConfig.CurrentContext = ctxName
			localConfig.UpserContext(config.ContextRef{
				Name:   ctxName,
				Server: server,
				User:   server,
			})

			err = config.WriteLocalConfig(*localConfig, configFile)
			errors.CheckError(err)

			fmt.Printf("Context '%s' updated\n", ctxName)
		},
	}

	loginCmd.Flags().StringVar(&ctxName, "name", "", "Name to use for the context")
	loginCmd.Flags().StringVar(&username, "username", "", "The username of an account to authenticate")
	loginCmd.Flags().StringVar(&password, "password", "", "The password of an account to authenticate")
	loginCmd.Flags().BoolVar(&sso, "sso", false, "Perform SSO login")
	loginCmd.Flags().BoolVar(&ssoLaunchBrowser, "sso-launch-browser", true, "Automatically launch the system default browser when performing SSO login")
	loginCmd.Flags().IntVar(&ssoProt, "sso-port", 8085, "Port to run local OAuth2 login application")

	return loginCmd
}

func oauth2login(
	ctx context.Context,
	port int,
	oauth2conf *oauth2.Config,
	ssoLaunchBrowser bool,
) (string, string) {
	oauth2conf.ClientID = "microcks-app-js"
	oauth2conf.RedirectURL = fmt.Sprintf("http://localhost:%d/auth/callback", port)

	// handledRequests ensures we do not handle more requests than necessary
	handledRequests := 0
	// completionChan is to signal flow completed. Non-empty string indicates error
	completionChan := make(chan string)

	stateNonce, err := rand.String(24)
	errors.CheckError(err)
	var tokenString string
	var refreshToken string

	handleErr := func(w http.ResponseWriter, errMsg string) {
		http.Error(w, html.EscapeString(errMsg), http.StatusBadRequest)
		completionChan <- errMsg
	}

	// PKCE implementation of https://tools.ietf.org/html/rfc7636
	codeVerifier, err := rand.StringFromCharset(
		43,
		"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~",
	)
	errors.CheckError(err)
	codeChallengeHash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(codeChallengeHash[:])

	// Authorization redirect callback from OAuth2 auth flow.
	// Handles both implicit and authorization code flow
	callbackHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Callback: %s\n", r.URL)

		if formErr := r.FormValue("error"); formErr != "" {
			handleErr(w, fmt.Sprintf("%s: %s", formErr, r.FormValue("error_description")))
			return
		}

		handledRequests++
		if handledRequests > 2 {
			// Since implicit flow will redirect back to ourselves, this counter ensures we do not
			// fallinto a redirect loop (e.g. user visits the page by hand)
			handleErr(w, "Unable to complete login flow: too many redirects")
			return
		}

		if state := r.FormValue("state"); state != stateNonce {
			handleErr(w, "Unknown state nonce")
			return
		}

		tokenString = r.FormValue("id_token")
		if tokenString == "" {
			code := r.FormValue("code")
			if code == "" {
				handleErr(w, fmt.Sprintf("no code in request: %q", r.Form))
				return
			}
			opts := []oauth2.AuthCodeOption{oauth2.SetAuthURLParam("code_verifier", codeVerifier)}
			tok, err := oauth2conf.Exchange(ctx, code, opts...)
			if err != nil {
				handleErr(w, err.Error())
				return
			}
			tokenString = tok.AccessToken
			refreshToken = tok.RefreshToken

		}
		successPage := `
		<div style="height:100px; width:100%!; display:flex; flex-direction: column; justify-content: center; align-items:center; background-color:#2ecc71; color:white; font-size:22"><div>Authentication successful!</div></div>
		<p style="margin-top:20px; font-size:18; text-align:center">Authentication was successful, you can now return to CLI. This page will close automatically</p>
		<script>window.onload=function(){setTimeout(this.close, 1000)}</script>
		`
		fmt.Fprint(w, successPage)
		completionChan <- ""
	}

	srv := &http.Server{Addr: "localhost:" + strconv.Itoa(port)}
	http.HandleFunc("/auth/callback", callbackHandler)

	var url string
	opts := []oauth2.AuthCodeOption{}
	opts = append(opts, oauth2.SetAuthURLParam("client_id", "microcks-app-js"))
	opts = append(opts, oauth2.SetAuthURLParam("scope", "openid"))
	opts = append(opts, oauth2.SetAuthURLParam("code_challenge", codeChallenge))
	opts = append(opts, oauth2.SetAuthURLParam("code_challenge_method", "S256"))
	url = oauth2conf.AuthCodeURL(stateNonce, opts...)

	fmt.Printf("Performing %s flow login: %s\n", "authorization_code", url)
	time.Sleep(1 * time.Second)
	ssoAuthFlow(url, ssoLaunchBrowser)
	go func() {
		log.Printf("Listen: %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Temporary HTTP server failed: %s", err)
		}
	}()
	errMsg := <-completionChan
	if errMsg != "" {
		log.Fatal(errMsg)
	}
	fmt.Printf("Authentication successful\n")
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Printf("Token: %s\n", tokenString)
	log.Printf("Refresh Token: %s\n", refreshToken)
	return tokenString, refreshToken
}

func ssoAuthFlow(url string, ssoLaunchBrowser bool) {
	if ssoLaunchBrowser {
		fmt.Printf("Opening system default browser for authentication\n")
		err := open.Start(url)
		errors.CheckError(err)
	} else {
		fmt.Printf("To authenticate, copy-and-paste the following URL into your preferred browser: %s\n", url)
	}
}

func passwordLogin(keycloakURL, clientId, clientSecret, Username, Password string) (string, string) {
	kc := connectors.NewKeycloakClient(keycloakURL, clientId, clientSecret)
	username, password := promptCredentials(Username, Password)

	authToken, refreshToken, err := kc.ConnectAndGetTokenAndRefreshToken(username, password)

	if err != nil {
		panic(err)
	}

	return authToken, refreshToken
}

func promptCredentials(username, password string) (string, string) {
	return promptUserName(username), promptPassword(password)
}

func promptUserName(value string) string {
	for value == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Username" + ": ")
		valueRaw, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		value = strings.TrimSpace(valueRaw)
	}
	return value
}

func promptPassword(password string) string {
	for password == "" {
		fmt.Print("Password: ")
		passwordRaw, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		password = string(passwordRaw)
		fmt.Print("\n")
	}
	return password
}

func StringField(claims jwt.MapClaims, fieldName string) string {
	if fieldIf, ok := claims[fieldName]; ok {
		if field, ok := fieldIf.(string); ok {
			return field
		}
	}
	return ""
}
