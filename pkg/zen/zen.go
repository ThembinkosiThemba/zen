package zen

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Engine is the framework instance
type Engine struct {
	*RouterGroup
	router *Router
	groups []*RouterGroup
}

type Route struct {
	Method string
	Path   string
}

// Colors for terminal output
const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	cyan   = "\033[36m"
	gray   = "\033[37m"
	white  = "\033[97m"
)

var (
	// HotReloadEnabled controls whether hot reload is enabled
	HotReloadEnabled bool

	// HotReloadConfig allows customization of hot reload behaviour
	HotReloadConfig = struct {
		// Directories to watch for changes
		WatchDirs []string

		// File extensions to watch
		Extensions []string

		// Paths to ignore
		IgnorePaths []string
	}{
		WatchDirs:   []string{"."},
		Extensions:  []string{".go"},
		IgnorePaths: []string{"tmp", "vendor", ".git"},
	}
)

// New creates a new Engine instance
func New() *Engine {
	engine := &Engine{
		router: newRouter(),
	}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Run starts the HTTP server
func (engine *Engine) Run(addr string) error {
	if HotReloadEnabled {
		return engine.runWithHotReload(addr)
	}
	// For normal starts or if this is the initial hot reload start
	if os.Getenv("ZEN_HOT_RELOAD_CHILD") != "1" || os.Getenv("ZEN_INITIAL_START") == "true" {
		engine.printRoutes()
		fmt.Print(engine.zenAsciiArt(addr))
	}
	return http.ListenAndServe(addr, engine)
}

// ServeHTTP implements the http.Handler interface
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	engine.router.handle(c)
}

// running the server with hot reload enabled
func (engine *Engine) runWithHotReload(addr string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	// Add directories to watch
	for _, dir := range HotReloadConfig.WatchDirs {
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip if matches ignore paths
			for _, ignorePath := range HotReloadConfig.IgnorePaths {
				if strings.Contains(path, ignorePath) {
					return nil
				}
			}

			if info.IsDir() {
				return watcher.Add(path)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk directory: %v", err)
		}
	}

	// Initial application start with routes display
	cmd := runInitialApp(engine, addr)
	if cmd == nil {
		return fmt.Errorf("failed to start application")
	}

	// Channel to track reload status
	reloading := make(chan bool, 1)
	debouncer := time.NewTimer(0)
	<-debouncer.C

	// Cleanup function
	cleanup := func() {
		if cmd != nil && cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}

	// Handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cleanup()
		os.Exit(0)
	}()

	lastReload := time.Now()
	cooldownPeriod := 30 * time.Second

	for {
		select {
		case event := <-watcher.Events:
			if time.Since(lastReload) < cooldownPeriod {
				continue
			}

			ext := filepath.Ext(event.Name)
			if !containsString(HotReloadConfig.Extensions, ext) {
				continue
			}

			skip := false
			for _, ignorePath := range HotReloadConfig.IgnorePaths {
				if strings.Contains(event.Name, ignorePath) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}

			select {
			case reloading <- true:
				if !debouncer.Stop() {
					select {
					case <-debouncer.C:
					default:
					}
				}
				debouncer.Reset(500 * time.Millisecond)

				go func() {
					<-debouncer.C
					lastReload = time.Now()

					fmt.Printf("%s[HOT RELOAD] Restarting server...%s\n", yellow, reset)
					cleanup()

					// Use quiet restart for subsequent reloads
					cmd = runQuietApp(addr)
					fmt.Printf("%s[HOT RELOAD] Server restarted%s\n", green, reset)

					<-reloading
				}()
			default:
				// Already reloading, skip
			}

		case err := <-watcher.Errors:
			log.Printf("watcher error: %v", err)
		}
	}
}

// runInitialApp starts the application with full output (routes and banner)
func runInitialApp(engine *Engine, addr string) *exec.Cmd {
	// Print routes and banner for initial start
	engine.printRoutes()
	fmt.Print(engine.zenAsciiArt(addr))

	return startApp(addr, true)
}

// runQuietApp starts the application without printing routes and banner
func runQuietApp(addr string) *exec.Cmd {
	return startApp(addr, false)
}

// startApp handles the common app starting logic
func startApp(addr string, isInitial bool) *exec.Cmd {
	args := os.Args[1:]
	hasAddr := false
	for i, arg := range args {
		if strings.HasPrefix(arg, ":") {
			args[i] = addr
			hasAddr = true
			break
		}
	}

	if !hasAddr {
		args = append(args, addr)
	}

	runArgs := append([]string{"run", "."}, args...)

	cmd := exec.Command("go", runArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"ZEN_HOT_RELOAD_CHILD=1",
		fmt.Sprintf("ZEN_INITIAL_START=%v", isInitial),
	)

	if err := cmd.Start(); err != nil {
		log.Printf("failed to start application: %v", err)
		return nil
	}

	return cmd
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (engine *Engine) Use(middlewares ...HandlerFunc) {
	engine.RouterGroup.Use(middlewares...)
}

func (engine *Engine) Routes() []Route {
	var routes []Route

	for method, paths := range engine.router.handlers {
		for path := range paths {
			routes = append(routes, Route{
				Method: method,
				Path:   path,
			})
		}
	}

	return routes
}

func (engine *Engine) printRoutes() {
	routes := engine.Routes()

	// Print header
	fmt.Printf("\n%s%s%s\n", green, "Registered Routes", reset)

	// Print horizontal line with box characters
	fmt.Printf("╔═════════╦%s╗\n", strings.Repeat("═", 30))

	// Print routes
	for _, r := range routes {
		methodColor := getMethodColor(r.Method)

		// Print each route with proper alignment
		fmt.Printf("║ %s%-7s%s ║ %-28s║\n",
			methodColor,
			r.Method,
			reset,
			r.Path)
	}

	// Print bottom border
	fmt.Printf("╚═════════╩%s╝\n", strings.Repeat("═", 30))
}

func (engine *Engine) zenAsciiArt(port string) string {
	return fmt.Sprintf(`
%s    ███████╗███████╗███╗   ██╗
    ╚══███╔╝██╔════╝████╗  ██║
      ███╔╝ █████╗  ██╔██╗ ██║
     ███╔╝  ██╔══╝  ██║╚██╗██║
    ███████╗███████╗██║ ╚████║
    ╚══════╝╚══════╝╚═╝  ╚═══╝%s
    
    %s🎋 Lightweight, Secure & Fast HTTP Framework for Modern Apps%s
    %s⚡ Running on port %s%s
    %s✨ %s%s
    `,
		cyan, reset,
		green, reset,
		yellow, port, reset,
		purple, time.Now().Format("2006-01-02 15:04:05"), reset)
}

// getMethodColor returns the color for HTTP methods
func getMethodColor(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return green
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return cyan
	case "HEAD":
		return purple
	case "OPTIONS":
		return white
	default:
		return reset
	}
}
