package main 

import (
	"errors"
	"kqnc/models"
	"kqnc/controllers"
	"thegoods.biz/httpbuf"
	"code.google.com/p/gorilla/pat"
	"code.google.com/p/gorilla/sessions"
	"encoding/gob"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"os/signal"
	"fmt"
	"flag"

	"kqnc/lib"
	"kqnc/lib/auth"
	"kqnc/lib/register"
	"kqnc/lib/store"
	"kqnc/net"
	"github.com/jeffail/util"
	"github.com/jeffail/util/log"
	"github.com/jeffail/util/metrics"
	"github.com/jeffail/util/path"
)

var sstore sessions.Store
var session *mgo.Session
var database string
var curator *lib.Curator
var router *pat.Router

type LeapsConfig struct {
	NumProcesses         int                      `json:"num_processes" yaml:"num_processes"`
	LoggerConfig         log.LoggerConfig         `json:"logger" yaml:"logger"`
	MetricsConfig        metrics.Config           `json:"metrics" yaml:"metrics"`
	StoreConfig          store.Config             `json:"storage" yaml:"storage"`
	AuthenticatorConfig  auth.Config              `json:"authenticator" yaml:"authenticator"`
	CuratorConfig        lib.CuratorConfig        `json:"curator" yaml:"curator"`
	HTTPServerConfig     net.HTTPServerConfig     `json:"http_server" yaml:"http_server"`
	InternalServerConfig net.InternalServerConfig `json:"admin_server" yaml:"admin_server"`
}

var (
	sharePathOverride *string
)

func init() {
	sharePathOverride = flag.String("share", "", "Override the path for file system sharing configs")
}

/*--------------------------------------------------------------------------------------------------
 */

var errEndpointNotConfigured = errors.New("HTTP Endpoint API required but not configured")

type endpointsRegister struct {
	publicRegister  register.EndpointRegister
	privateRegister register.EndpointRegister
}

func newEndpointsRegister(public, private register.EndpointRegister) register.PubPrivEndpointRegister {
	return &endpointsRegister{
		publicRegister:  public,
		privateRegister: private,
	}
}

func (e *endpointsRegister) RegisterPublic(endpoint, description string, handler http.HandlerFunc) error {
	if e.publicRegister == nil {
		return errEndpointNotConfigured
	}
	e.publicRegister.Register(endpoint, description, handler)
	return nil
}

func (e *endpointsRegister) RegisterPrivate(endpoint, description string, handler http.HandlerFunc) error {
	if e.publicRegister == nil {
		return errEndpointNotConfigured
	}
	e.privateRegister.Register(endpoint, description, handler)
	return nil
}

type handler func(http.ResponseWriter, *http.Request, *models.Context) error

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  //create the context
  ctx, err := models.NewContext(r, sstore, session, database, curator)
  if err != nil {
  	  //http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  defer ctx.Close()

  //run the handler and grab the error, and report it
  buf := new(httpbuf.Buffer)
  err = h(buf, r, ctx)
  if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
  }

  //save the session
  if err = ctx.Session.Save(r, buf); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
  }

  //apply the buffered response to the writer
  buf.Apply(w)
}

func init() {
  gob.Register(bson.ObjectId(""))
}

func main() {
	var (
		err       error
		closeChan = make(chan bool)
	)

	leapsConfig := LeapsConfig{
		NumProcesses:         runtime.NumCPU(),
		LoggerConfig:         log.DefaultLoggerConfig(),
		MetricsConfig:        metrics.NewConfig(),
		StoreConfig:          store.NewConfig(),
		AuthenticatorConfig:  auth.NewConfig(),
		CuratorConfig:        lib.DefaultCuratorConfig(),
		HTTPServerConfig:     net.DefaultHTTPServerConfig(),
		InternalServerConfig: net.NewInternalServerConfig(),
	}

	// A list of default config paths to check for if not explicitly defined
	defaultPaths := []string{"config/leaps_example.yaml"}

	/* If we manage to get the path of our executable then we want to try and find config files
	 * relative to that path, we always check from the parent folder since we assume leaps is
	 * stored within the bin folder.
	 */
	if executablePath, err := path.BinaryPath(); err == nil {
		defaultPaths = append(defaultPaths, filepath.Join(executablePath, "..", "config.yaml"))
		defaultPaths = append(defaultPaths, filepath.Join(executablePath, "..", "config", "leaps.yaml"))
		defaultPaths = append(defaultPaths, filepath.Join(executablePath, "..", "config.json"))
		defaultPaths = append(defaultPaths, filepath.Join(executablePath, "..", "config", "leaps.json"))
	}

	defaultPaths = append(defaultPaths, []string{
		filepath.Join(".", "leaps.yaml"),
		filepath.Join(".", "leaps.json"),
		"/etc/leaps.yaml",
		"/etc/leaps.json",
		"/etc/leaps/config.yaml",
		"/etc/leaps/config.json",
	}...)

	// Load configuration etc
	if !util.Bootstrap(&leapsConfig, defaultPaths...) {
		return
	}

	// Logging and stats aggregation
	logger := log.NewLogger(os.Stdout, leapsConfig.LoggerConfig)
	var stats metrics.Aggregator
	if s, err := metrics.New(leapsConfig.MetricsConfig); err == nil {
		stats = s
	} else {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Metrics init error: %v\n", err))
		return
	}
	defer stats.Close()

	// Document storage engine
	documentStore, err := store.Factory(leapsConfig.StoreConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Document store error: %v\n", err))
		return
	}

	// Authenticator
	authenticator, err := auth.Factory(leapsConfig.AuthenticatorConfig, logger, stats)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Authenticator error: %v\n", err))
		return
	}

	// Curator of documents
	curator, err = lib.NewCurator(leapsConfig.CuratorConfig, logger, stats, authenticator, documentStore)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Curator error: %v\n", err))
		return
	}
	defer curator.Close()

	// HTTP API
	leapHTTP, err := net.CreateHTTPServer(curator, leapsConfig.HTTPServerConfig, logger, stats)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("HTTP error: %v\n", err))
		return
	}
	defer leapHTTP.Stop()

	go func() {
		if httperr := leapHTTP.Listen(); httperr != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("Http listen error: %v\n", httperr))
		}
		closeChan <- true
	}()







	session, err = mgo.Dial(os.Getenv("127.0.0.1"))
	if err != nil {
		panic(err)
	}

	database = session.DB("").Name
	if err := session.DB("").C("users").EnsureIndex(mgo.Index{
        Key:    []string{"email"},
        Unique: true,
    }); err != nil {
        //log.Println("Ensuring unqiue index on users:", err)
    }

	sstore = sessions.NewCookieStore([]byte(os.Getenv("kqnc")))

	router = pat.New()
	controllers.Init(router)

	router.Add("GET", "/login", handler(controllers.LoginForm)).Name("login")
	router.Add("POST", "/login", handler(controllers.Login))
	router.Add("GET", "/logout", handler(controllers.Logout)).Name("logout")
	router.Add("GET", "/register", handler(controllers.RegisterForm)).Name("register")
	router.Add("POST", "/register", handler(controllers.Register))

	router.Add("GET", "/documents/new", handler(controllers.NewDocumentForm))
	router.Add("POST", "/documents/new", handler(controllers.NewDocument))
	router.Add("GET", "/documents/{id}", handler(controllers.DocumentForm))
	router.Add("GET", "/documents", handler(controllers.DocumentIndexForm))

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	router.Add("GET", "/", handler(controllers.Index)).Name("index")

	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}

func mainLeap() {
	var (
		err       error
		closeChan = make(chan bool)
	)

	leapsConfig := LeapsConfig{
		NumProcesses:         runtime.NumCPU(),
		LoggerConfig:         log.DefaultLoggerConfig(),
		MetricsConfig:        metrics.NewConfig(),
		StoreConfig:          store.NewConfig(),
		AuthenticatorConfig:  auth.NewConfig(),
		CuratorConfig:        lib.DefaultCuratorConfig(),
		HTTPServerConfig:     net.DefaultHTTPServerConfig(),
		InternalServerConfig: net.NewInternalServerConfig(),
	}

	// A list of default config paths to check for if not explicitly defined
	defaultPaths := []string{"config/leaps_example.yaml"}

	/* If we manage to get the path of our executable then we want to try and find config files
	 * relative to that path, we always check from the parent folder since we assume leaps is
	 * stored within the bin folder.
	 */
	if executablePath, err := path.BinaryPath(); err == nil {
		defaultPaths = append(defaultPaths, filepath.Join(executablePath, "..", "config.yaml"))
		defaultPaths = append(defaultPaths, filepath.Join(executablePath, "..", "config", "leaps.yaml"))
		defaultPaths = append(defaultPaths, filepath.Join(executablePath, "..", "config.json"))
		defaultPaths = append(defaultPaths, filepath.Join(executablePath, "..", "config", "leaps.json"))
	}

	defaultPaths = append(defaultPaths, []string{
		filepath.Join(".", "leaps.yaml"),
		filepath.Join(".", "leaps.json"),
		"/etc/leaps.yaml",
		"/etc/leaps.json",
		"/etc/leaps/config.yaml",
		"/etc/leaps/config.json",
	}...)

	// Load configuration etc
	if !util.Bootstrap(&leapsConfig, defaultPaths...) {
		return
	}

	if len(*sharePathOverride) > 0 {
		leapsConfig.AuthenticatorConfig.FileConfig.SharePath = *sharePathOverride
		leapsConfig.StoreConfig.StoreDirectory = *sharePathOverride
	}

	// Logging and stats aggregation
	logger := log.NewLogger(os.Stdout, leapsConfig.LoggerConfig)
	var stats metrics.Aggregator
	if s, err := metrics.New(leapsConfig.MetricsConfig); err == nil {
		stats = s
	} else {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Metrics init error: %v\n", err))
		return
	}
	defer stats.Close()

	fmt.Printf("Launching a leaps instance, use CTRL+C to close.\n\n")

	// Document storage engine
	documentStore, err := store.Factory(leapsConfig.StoreConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Document store error: %v\n", err))
		return
	}

	// Authenticator
	authenticator, err := auth.Factory(leapsConfig.AuthenticatorConfig, logger, stats)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Authenticator error: %v\n", err))
		return
	}

	// Curator of documents
	curator, err := lib.NewCurator(leapsConfig.CuratorConfig, logger, stats, authenticator, documentStore)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Curator error: %v\n", err))
		return
	}
	defer curator.Close()

	// HTTP API
	leapHTTP, err := net.CreateHTTPServer(curator, leapsConfig.HTTPServerConfig, logger, stats)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("HTTP error: %v\n", err))
		return
	}
	defer leapHTTP.Stop()

	go func() {
		if httperr := leapHTTP.Listen(); httperr != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("Http listen error: %v\n", httperr))
		}
		closeChan <- true
	}()

	var adminRegister register.EndpointRegister

	// Internal admin HTTP API
	if 0 < len(leapsConfig.InternalServerConfig.Address) {
		adminHTTP, err := net.NewInternalServer(curator, leapsConfig.InternalServerConfig, logger, stats)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("Admin HTTP error: %v\n", err))
			return
		}
		adminRegister = adminHTTP

		go func() {
			if httperr := adminHTTP.Listen(); httperr != nil {
				fmt.Fprintln(os.Stderr, fmt.Sprintf("Admin HTTP listen error: %v\n", httperr))
			}
			closeChan <- true
		}()
	}

	// Register for allowing other components to set API endpoints.
	register := newEndpointsRegister(leapHTTP, adminRegister)
	if err = authenticator.RegisterHandlers(register); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Register authentication endpoints failed: %v\n", err))
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for termination signal
	select {
	case <-sigChan:
	case <-closeChan:
	}
}