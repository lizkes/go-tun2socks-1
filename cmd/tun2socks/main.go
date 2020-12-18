package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/eycorsican/go-tun2socks/common/dns/blocker"
	"github.com/eycorsican/go-tun2socks/common/log"
	_ "github.com/eycorsican/go-tun2socks/common/log/simple" // Register a simple logger.
	"github.com/eycorsican/go-tun2socks/core"
	"github.com/eycorsican/go-tun2socks/routes"
	"github.com/eycorsican/go-tun2socks/tun"
)

var version = "undefined"

var handlerCreater = make(map[string]func(), 0)

func registerHandlerCreater(name string, creater func()) {
	handlerCreater[name] = creater
}

var postFlagsInitFn = make([]func(), 0)

func addPostFlagsInitFn(fn func()) {
	postFlagsInitFn = append(postFlagsInitFn, fn)
}

type CmdArgs struct {
	Version         *bool
	TunName         *string
	TunAddr         *string
	TunGw           *string
	TunMask         *string
	TunDns          *string
	TunMTU          *int
	TunPersist      *bool
	BlockOutsideDns *bool
	ProxyType       *string
	ProxyServer     *string
	ProxyHost       *string
	ProxyPort       *uint16
	UdpTimeout      *time.Duration
	LogLevel        *string
	DnsFallback     *bool
	Routes          *string
	Exclude         *string
}

type cmdFlag uint

const (
	fProxyServer cmdFlag = iota
	fUdpTimeout
)

var flagCreaters = map[cmdFlag]func(){
	fProxyServer: func() {
		if args.ProxyServer == nil {
			args.ProxyServer = flag.String("proxyServer", "1.2.3.4:1087", "Proxy server address")
		}
	},
	fUdpTimeout: func() {
		if args.UdpTimeout == nil {
			args.UdpTimeout = flag.Duration("udpTimeout", 1*time.Minute, "UDP session timeout")
		}
	},
}

func (a *CmdArgs) addFlag(f cmdFlag) {
	if fn, found := flagCreaters[f]; found && fn != nil {
		fn()
	} else {
		log.Fatalf("unsupported flag")
	}
}

var args = new(CmdArgs)

var lwipWriter io.Writer

const (
	maxMTU = 65535
)

func main() {
	// linux and darwin pick up the tun index automatically
	// windows requires the exact tun name
	defaultTunName := ""
	switch runtime.GOOS {
	case "darwin":
		defaultTunName = "utun"
	case "windows":
		defaultTunName = "socks2tun"
	}
	args.TunName = flag.String("tunName", defaultTunName, "TUN interface name")

	args.Version = flag.Bool("version", false, "Print version")
	args.TunAddr = flag.String("tunAddr", "10.255.0.2", "TUN interface address")
	args.TunGw = flag.String("tunGw", "10.255.0.1", "TUN interface gateway")
	args.TunMask = flag.String("tunMask", "255.255.255.255", "TUN interface netmask, it should be a prefixlen (a number) for IPv6 address")
	args.TunDns = flag.String("tunDns", "8.8.8.8,8.8.4.4", "DNS resolvers for TUN interface (only need on Windows)")
	args.TunMTU = flag.Int("tunMTU", 1300, "TUN interface MTU")
	args.TunPersist = flag.Bool("tunPersist", false, "Persist TUN interface after the program exits or the last open file descriptor is closed (Linux only)")
	args.BlockOutsideDns = flag.Bool("blockOutsideDns", false, "Prevent DNS leaks by blocking plaintext DNS queries going out through non-TUN interface (may require admin privileges) (Windows only) ")
	args.ProxyType = flag.String("proxyType", "socks", "Proxy handler type")
	args.LogLevel = flag.String("loglevel", "info", "Logging level. (debug, info, warn, error, none)")
	args.Routes = flag.String("routes", "", "Subnets to forward via TUN interface")
	args.Exclude = flag.String("exclude", "", "Subnets or hostnames to exclude from forwarding via TUN interface")

	flag.Parse()

	if *args.TunMTU > maxMTU {
		fmt.Printf("MTU exceeds %d\n", maxMTU)
		os.Exit(1)
	}

	if *args.Version {
		fmt.Println(version)
		os.Exit(0)
	}

	// Initialization ops after parsing flags.
	for _, fn := range postFlagsInitFn {
		if fn != nil {
			fn()
		}
	}

	// Set log level.
	switch strings.ToLower(*args.LogLevel) {
	case "debug":
		log.SetLevel(log.DEBUG)
	case "info":
		log.SetLevel(log.INFO)
	case "warn":
		log.SetLevel(log.WARN)
	case "error":
		log.SetLevel(log.ERROR)
	case "none":
		log.SetLevel(log.NONE)
	default:
		panic("unsupport logging level")
	}

	tunGw, tunRoutes, err := routes.Get(*args.Routes, *args.Exclude, *args.TunAddr, *args.TunGw, *args.TunMask)
	if err != nil {
		log.Fatalf("cannot parse config values: %v", err)
	}

	// Open the tun device.
	dnsServers := strings.Split(*args.TunDns, ",")
	tunDev, err := tun.OpenTunDevice(*args.TunName, *args.TunAddr, *args.TunGw, *args.TunMask, dnsServers, *args.TunPersist, *args.TunMTU)
	if err != nil {
		log.Fatalf("failed to open tun device: %v", err)
	}

	// unset routes on exit, when provided
	defer routes.Unset(*args.TunName, tunGw, tunRoutes)

	// set routes, when provided
	routes.Set(*args.TunName, tunGw, tunRoutes)

	if runtime.GOOS == "windows" && *args.BlockOutsideDns {
		if err := blocker.BlockOutsideDns(*args.TunName); err != nil {
			log.Fatalf("failed to block outside DNS: %v", err)
		}
	}

	// Setup TCP/IP stack.
	lwipWriter := core.NewLWIPStack().(io.Writer)

	// Register TCP and UDP handlers to handle accepted connections.
	if creater, found := handlerCreater[*args.ProxyType]; found {
		creater()
	} else {
		log.Fatalf("unsupported proxy type: %s", *args.ProxyType)
	}

	if args.DnsFallback != nil && *args.DnsFallback {
		// Override the UDP handler with a DNS-over-TCP (fallback) UDP handler.
		if creater, found := handlerCreater["dnsfallback"]; found {
			creater()
		} else {
			log.Fatalf("DNS fallback connection handler not found, build with `dnsfallback` tag")
		}
	}

	// Register an output callback to write packets output from lwip stack to tun
	// device, output function should be set before input any packets.
	core.RegisterOutputFn(func(data []byte) (int, error) {
		return tunDev.Write(data)
	})

	// Copy packets from tun device to lwip stack, it's the main loop.
	go func() {
		_, err := io.CopyBuffer(lwipWriter, tunDev, make([]byte, maxMTU))
		if err != nil {
			log.Fatalf("copying data failed: %v", err)
		}
	}()

	log.Infof("Running tun2socks")

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
	<-osSignals
}
