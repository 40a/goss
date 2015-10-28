package system

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/aelsabbahy/GOnetstat"
	"github.com/codegangsta/cli"
	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/util"
	"github.com/mitchellh/go-ps"
)

type Resource interface {
	Exists() (interface{}, error)
}

type System struct {
	NewPackage  func(string, *System) Package
	NewFile     func(string, *System) File
	NewAddr     func(string, *System) Addr
	NewPort     func(string, *System) Port
	NewService  func(string, *System) Service
	NewUser     func(string, *System) User
	NewGroup    func(string, *System) Group
	NewCommand  func(string, *System) Command
	NewDNS      func(string, *System) DNS
	NewProcess  func(string, *System) Process
	NewURL      func(string, *System) URL
	NewGossfile func(string, *System) Gossfile
	Dbus        *dbus.Conn
	ports       map[string]GOnetstat.Process
	portsOnce   sync.Once
	procOnce    sync.Once
	procMap     map[string][]ps.Process
}

func (s *System) Ports() map[string]GOnetstat.Process {
	s.portsOnce.Do(func() {
		s.ports = GetPorts(false)
	})
	return s.ports
}

func (s *System) ProcMap() map[string][]ps.Process {
	s.procOnce.Do(func() {
		s.procMap = GetProcs()
	})
	return s.procMap
}

func New(c *cli.Context) *System {
	system := &System{
		NewFile:     NewDefFile,
		NewAddr:     NewDefAddr,
		NewPort:     NewDefPort,
		NewUser:     NewDefUser,
		NewGroup:    NewDefGroup,
		NewCommand:  NewDefCommand,
		NewDNS:      NewDefDNS,
		NewProcess:  NewDefProcess,
		NewURL:      NewDefURL,
		NewGossfile: NewDefGossfile,
	}

	if util.IsRunningSystemd() {
		system.NewService = NewServiceDbus
		dbus, err := dbus.New()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		system.Dbus = dbus
	} else {
		system.NewService = NewServiceInit
	}

	switch {
	case c.GlobalString("package") == "rpm":
		system.NewPackage = NewRpmPackage
	case c.GlobalString("package") == "deb":
		system.NewPackage = NewDebPackage
	default:
		system.NewPackage = detectPackage()
	}

	return system
}

func detectPackage() func(string, *System) Package {
	switch {
	case isRpm():
		return NewRpmPackage
	case isDeb():
		return NewDebPackage
	default:
		return NewNullPackage
	}
}

func isDeb() bool {
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		return true
	}

	// See if it has only one of the package managers
	if hasCommand("dpkg") && !hasCommand("rpm") {
		return true
	}

	return false
}

func isRpm() bool {
	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		return true
	}

	if _, err := os.Stat("/etc/system-release"); err == nil {
		return true
	}

	// See if it has only one of the package managers
	if hasCommand("rpm") && !hasCommand("dpkg") {
		return true
	}
	return false
}

func hasCommand(cmd string) bool {
	if _, err := exec.LookPath(cmd); err == nil {
		return true
	}
	return false
}
