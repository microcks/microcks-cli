package config

import (
	"fmt"
	"os"
	"path"
	"slices"

	configUtil "github.com/microcks/microcks-cli/pkg/util"
)

type LocalConfig struct {
	CurrentContext string       `yaml:"current-context"`
	Contexts       []ContextRef `yaml:"contexts"`
	Servers        []Server     `yaml:"servers"`
	Users          []User       `yaml:"users"`
	Instances      []Instance   `yaml:"instances"`
	Auths          []Auth       `yaml:"auths"`
}

type ContextRef struct {
	Name     string `yaml:"name"`
	Server   string `yaml:"server"`
	User     string `yaml:"user"`
	Instance string `yaml:"instance"`
}

type Context struct {
	Name     string
	User     User
	Server   Server
	Instance Instance
}

type User struct {
	Name         string `yaml:"name"`
	AuthToken    string `yaml:"auth-token"`
	RefreshToken string `yaml:"refresh-token"`
}

type Server struct {
	Name            string `yaml:"name"`
	Server          string `yaml:"server"`
	InsecureTLS     bool   `yaml:"insecureTLS"`
	KeycloackEnable bool   `yaml:"keycloakEnable"`
}

type Instance struct {
	Name        string `yaml:"name"`
	Image       string `yaml:"image"`
	Status      string `yaml:"status"`
	Port        string `yaml:"port"`
	ContainerID string `yaml:"containerID"`
	AutoRemove  bool   `yaml:"autoRemove"`
	Driver      string `yaml:"driver"`
}

type Auth struct {
	Server       string
	ClientId     string
	ClientSecret string
}

type WatchConfig struct {
	Entries []WatchEntry `yaml:"entries"`
}

type WatchEntry struct {
	FilePath     string   `yaml:"filePath"`
	Context      []string `yaml:"context"`
	MainArtifact bool     `yaml:"mainartifact"`
}

// ReadLocalConfig loads up the local configuration file. Returns nil if config does not exist
func ReadLocalConfig(path string) (*LocalConfig, error) {
	var err error
	var config LocalConfig

	// check file permission only when microcks config exists
	if fi, err := os.Stat(path); err == nil {
		err = getFilePermission(fi)
		if err != nil {
			return nil, err
		}
	}

	err = configUtil.UnmarshalLocalFile(path, &config)
	if os.IsNotExist(err) {
		return nil, nil
	}

	err = ValidateLocalConfig(config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// DefaultConfigDir returns the local configuration path for settings such as cached authentication tokens.
func DefaultConfigDir() (string, error) {

	configDir := os.Getenv("MICROCKS_CONFIG_DIR")

	if configDir != "" {
		return configDir, nil
	}

	homeDir, err := getHomeDir()
	if err != nil {
		return "", nil
	}

	configDir = path.Join(homeDir, ".config", "microcks")

	return configDir, nil
}

func getHomeDir() (string, error) {
	homedir, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	return homedir, nil
}

// DefaultLocalConfigPath returns the local configuration path for settings such as cached authentication tokens.
func DefaultLocalConfigPath() (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}
	return path.Join(dir, "config"), nil
}

func DefaultLocalWatchPath() (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}
	return path.Join(dir, "watch"), nil
}

func ValidateLocalConfig(config LocalConfig) error {
	if config.CurrentContext == "" {
		return nil
	}
	if _, err := config.ResolveContext(config.CurrentContext); err != nil {
		return fmt.Errorf("local config invalid: %w", err)
	}
	return nil
}

// WriteLocalConfig writes a new local configuration file.
func WriteLocalConfig(config LocalConfig, configPath string) error {
	err := os.MkdirAll(path.Dir(configPath), os.ModePerm)
	if err != nil {
		return err
	}
	return configUtil.MarshalLocalYAMLFile(configPath, &config)
}

func (l *LocalConfig) DeleteLocalConfig(configPath string) error {
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return err
	}
	return os.Remove(configPath)
}

// ResolveContext resolves the specified context. If unspecified, resolves the current context
func (l *LocalConfig) ResolveContext(name string) (*Context, error) {
	if name == "" {
		if l.CurrentContext == "" {
			return nil, fmt.Errorf("local config: current-context unset")
		}
		name = l.CurrentContext
	}
	for _, ctx := range l.Contexts {
		if ctx.Name == name {
			server, err := l.GetServer(ctx.Server)
			if err != nil {
				return nil, err
			}
			user, err := l.GetUser(ctx.User)
			if err != nil {
				return nil, err
			}
			instance, err := l.GetInstance(ctx.Instance)
			if err != nil {
				instance = &Instance{}
			}
			return &Context{
				Name:     ctx.Name,
				Server:   *server,
				User:     *user,
				Instance: *instance,
			}, nil
		}
	}
	return nil, fmt.Errorf("Context '%s' undefined", name)
}

func (l *LocalConfig) UpserContext(context ContextRef) {
	for i, c := range l.Contexts {
		if c.Name == context.Name {
			l.Contexts[i] = context
			return
		}
	}
	l.Contexts = append(l.Contexts, context)
}

func (l *LocalConfig) RemoveContext(serverName string) (string, bool) {
	for i, c := range l.Contexts {
		if c.Name == serverName {
			l.Contexts = append(l.Contexts[:i], l.Contexts[i+1:]...)
			return c.Server, true
		}
	}
	return "", false
}

// Returns true if user was removed successfully
func (l *LocalConfig) RemoveToken(serverName string) bool {
	for i, u := range l.Users {
		if u.Name == serverName {
			l.Users[i].RefreshToken = ""
			l.Users[i].AuthToken = ""
			return true
		}
	}
	return false
}

func (l *LocalConfig) GetUser(name string) (*User, error) {
	for _, u := range l.Users {
		if u.Name == name {
			return &u, nil
		}
	}
	return nil, fmt.Errorf("User '%s' undefined", name)
}

func (l *LocalConfig) UpsertUser(user User) {
	for i, u := range l.Users {
		if u.Name == user.Name {
			l.Users[i] = user
			return
		}
	}

	l.Users = append(l.Users, user)
}

// Returns true if user was removed successfully
func (l *LocalConfig) RemoveUser(serverName string) bool {
	for i, u := range l.Users {
		if u.Name == serverName {
			l.Users = append(l.Users[:i], l.Users[i+1:]...)
			return true
		}
	}
	return false
}

func (l *LocalConfig) GetServer(name string) (*Server, error) {
	for _, s := range l.Servers {
		if s.Server == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("Server '%s' undefined", name)
}

func (l *LocalConfig) UpsertServer(server Server) {
	for i, s := range l.Servers {
		if s.Server == server.Server {
			l.Servers[i] = server
			return
		}
	}
	l.Servers = append(l.Servers, server)
}

// Returns true if server was removed successfully
func (l *LocalConfig) RemoveServer(serverName string) bool {
	for i, s := range l.Servers {
		if s.Server == serverName {
			l.Servers = append(l.Servers[:i], l.Servers[i+1:]...)
			return true
		}
	}
	return false
}

func (l *LocalConfig) GetInstance(name string) (*Instance, error) {
	for _, i := range l.Instances {
		if i.Name == name {
			return &i, nil
		}
	}
	return nil, fmt.Errorf("Instance '%s' undefined", name)
}

func (l *LocalConfig) UpsertInstance(instance Instance) {
	for a, i := range l.Instances {
		if i.ContainerID == instance.ContainerID {
			l.Instances[a] = instance
			return
		}
	}
	l.Instances = append(l.Instances, instance)
}

// Returns true if server was removed successfully
func (l *LocalConfig) RemoveInstance(instanceName string) bool {
	if instanceName == "" {
		return true
	}
	for a, i := range l.Instances {
		if i.Name == instanceName {
			l.Instances = append(l.Instances[:a], l.Instances[a+1:]...)
			return true
		}
	}
	return false
}

func (l *LocalConfig) IsEmpty() bool {
	return len(l.Servers) == 0
}

func (l *LocalConfig) GetAuth(server string) (*Auth, error) {
	for _, a := range l.Auths {
		if a.Server == server {
			return &a, nil
		}
	}

	return nil, fmt.Errorf("Auth for '%s' is undifined", server)
}

func (l *LocalConfig) UpserAuth(auth Auth) {
	for i, a := range l.Auths {
		if a.Server == auth.Server {
			l.Auths[i] = auth
			return
		}
	}

	l.Auths = append(l.Auths, auth)
}

func (l *LocalConfig) RemoveAuth(server string) bool {
	for i, a := range l.Auths {
		if a.Server == server {
			l.Auths = append(l.Auths[:i], l.Auths[i+1:]...)
			return true
		}
	}
	return false
}

func (w *WatchConfig) UpsertEntry(entry WatchEntry) {
	for i, e := range w.Entries {
		if e.FilePath == entry.FilePath {
			contexts := w.Entries[i].Context
			if !slices.Contains(contexts, entry.Context[0]) {
				entry.Context = append(entry.Context, contexts...)
			}
			w.Entries[i] = entry
			return
		}
	}
	w.Entries = append(w.Entries, entry)
}

func ReadLocalWatchConfig(path string) (*WatchConfig, error) {
	var err error
	var config WatchConfig

	// check file permission only when microcks config exists
	if fi, err := os.Stat(path); err == nil {
		err = getFilePermission(fi)
		if err != nil {
			return nil, err
		}
	}

	err = configUtil.UnmarshalLocalFile(path, &config)
	if os.IsNotExist(err) {
		return nil, nil
	}

	return &config, nil
}

func WriteLocalWatchConfig(config WatchConfig, cfgPath string) error {
	err := os.MkdirAll(path.Dir(cfgPath), os.ModePerm)
	if err != nil {
		return err
	}
	return configUtil.MarshalLocalYAMLFile(cfgPath, &config)
}
