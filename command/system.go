// Package command for system
// System executes operating system commands and stores output
package command

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/cli"
)

const (
	// Darwin OS
	Darwin string = "darwin"
	// FreeBSD OS
	FreeBSD string = "freebsd"
	// Linux OS
	Linux string = "linux"
	// NetBSD OS
	NetBSD string = "netbsd"
	// OpenBSD OS
	OpenBSD string = "openbsd"
	// Solaris OS
	Solaris string = "solaris"
	// Windows OS
	Windows string = "windows"
)

// SystemCommand describes system related fields
type SystemCommand struct {
	Arch         string
	HostName     string
	OS           string
	ReleaseFiles []string
	UI           cli.Ui
	LogFile      string
}

// Help output
func (c *SystemCommand) Help() string {
	helpText := `
Usage: rover system
	Executes operating system commands and saves output to text files
`

	return strings.TrimSpace(helpText)
}

// Run the command
func (c *SystemCommand) Run(_ []string) int {
	c.Arch = runtime.GOARCH
	h, err := GetHostName()
	if err != nil {
		out := fmt.Sprintf("Cannot get system hostname with error %v", err)
		c.UI.Output(out)

		return 1
	}
	c.HostName = h
	c.OS = runtime.GOOS
	c.ReleaseFiles = []string{"/etc/redhat-release",
		"/etc/fedora-release",
		"/etc/slackware-release",
		"/etc/debian_release",
		"/etc/os-release"}

	// Shout out to Ye Olde School BSD spinner!
	roverSpinnerSet := []string{"/", "|", "\\", "-", "|", "\\", "-"}
	s := spinner.New(roverSpinnerSet, 174*time.Millisecond)
	s.Writer = os.Stderr
	err = s.Color("fgHiCyan")
	if err != nil {
		log.Printf("install", "weird-error", err.Error())
	}
	s.Suffix = " Gathering system data, please wait ..."
	s.FinalMSG = "Gathered system data\n"
	s.Start()

	// Internal logging
	l := "rover.log"
	p := filepath.Join(fmt.Sprintf("%s", c.HostName), "log")
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		fmt.Println(fmt.Sprintf("Cannot create log directory %s.", p))
		os.Exit(1)
	}
	f, err := os.OpenFile(filepath.Join(p, l), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to open log file %s with error: %v", f, err))
		os.Exit(1)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "rover", Level: hclog.LevelFromString("INFO"), Output: w})

	logger.Info("system", "hello from the System module at", c.HostName)
	logger.Info("system", "our detected OS", c.OS)

	// Handle creating the command output directory
	// TODO: maybe not such a hardcoded path?
	outPath := filepath.Join(".", fmt.Sprintf("%s/system", c.HostName))
	if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
		out := fmt.Sprintf("Cannot create directory %s!", outPath)
		logger.Error("system", "cannot create directory with error", err.Error())
		c.UI.Error(out)
		os.Exit(1)
	}

	// Grab OS release info
	for _, file := range c.ReleaseFiles {
		if FileExist(file) {
			Dump("system", "os_release", "cat", file)
			logger.Info("system", "dumping release file", file)
		}
	}

	// Common commands and file contents from which to gather information
	// across the currently supported range of operating system commands
	Dump("system", "date", "date")
	Dump("system", "df", "df")
	Dump("system", "df_i", "df", "-i")
	Dump("system", "df_h", "df", "-h")
	Dump("system", "dmesg", "dmesg")
	Dump("system", "hostname", "hostname")
	Dump("system", "last", "last")
	Dump("system", "mount", "mount")
	Dump("system", "netstat_anW", "netstat", "-anW")
	Dump("system", "netstat_indW", "netstat", "-indW")
	Dump("system", "netstat_mmmW", "netstat", "-mmmW")
	Dump("system", "netstat_nralW", "netstat", "-nralW")
	Dump("system", "netstat_rn", "netstat", "-rn")
	Dump("system", "netstat_sW", "netstat", "-sW")
	Dump("system", "pfctl_rules", "pfctl", "-s rules")
	Dump("system", "pfctl_nat", "pfctl", "-s nat")
	Dump("system", "sysctl", "sysctl", "-a")
	Dump("system", "uname", "uname", "-a")
	Dump("system", "w", "w")

	// File contents
	Dump("system", "file_etc_fstab", "cat", "/etc/fstab")
	Dump("system", "file_etc_hosts", "cat", "/etc/hosts")
	Dump("system", "file_etc_resolv_conf", "cat", "/etc/resolv.conf")

	// Different command subsets chosen by OS
	// We use runtime.GOOS for now as it is accurate enough for
	// the platforms we are targeting...
	// Could look at build constraints later as well
	//
	// Future commands which are optional on some systems
	//
	// dig -t any -c any brianshumate.com
	// lsof
	//
	//
	switch c.OS {

	case Darwin:

		// Darwin specific commands
		Dump("system", "ifconfig", "ifconfig", "-a")
		Dump("system", "netstat_rs", "netstat", "-rs")
		Dump("system", "ps", "ps", "aux")
		Dump("system", "top", "top", "-l 1")
		Dump("system", "vm_stat", "vm_stat")

	case FreeBSD:

		// FreeBSD specific ommands
		Dump("system", "arp_a", "arp", "-a")
		Dump("system", "ifconfig", "ifconfig", "-a")
		Dump("system", "iostat_bsd", "iostat", "-c 10")
		Dump("system", "pkg_info", "pkg", "info")
		Dump("system", "ps", "ps", "aux")
		Dump("system", "swapinfo", "swapinfo")
		Dump("system", "sysctl", "sysctl", "-a")
		Dump("system", "top", "top", "-n", "-b")
		Dump("system", "vmstat", "vmstat", "1", "10")

		// File contents
		Dump("system", "file_var_run_dmesg_boot", "cat", "/var/run/dmesg.boot")
		Dump("system", "file_var_log_messages", "cat", "/var/log/messages")
		Dump("system", "file_etc_rc_conf", "cat", "/etc/rc.conf")
		Dump("system", "file_etc_sysctl_conf", "cat", "/etc/sysctl.conf")

	case Linux:

		// Linux specific commands
		Dump("system", "bonding", "find", "/proc/net/bonding/", "-type", "f", "-print", "-exec", "cat", "{}", ";")
		Dump("system", "disk_by_id", "ls", "-l", "/dev/disk/by-id")
		Dump("system", "dmesg", "dmesg")
		Dump("system", "dpkg", "dpkg", "-l")
		Dump("system", "free", "free", "-m")
		Dump("system", "ifconfig", "ifconfig", "-a")
		Dump("system", "iostat_linux", "iostat", "-mx", "1", "10")
		Dump("system", "ip_addr", "ip", "addr")
		Dump("system", "lsb_release", "lsb_release")
		Dump("system", "ps", "ps", "-aux")
		Dump("system", "rpm", "rpm", "-qa")
		Dump("system", "rx_crc_errors", "find", "/sys/class/net/", "-type", "l", "-print", "-exec", "cat", "{}/statistics/rx_crc_errors", ";")
		Dump("system", "schedulers", "find", "/sys/block/", "-type", "l", "-print", "-exec", "cat", "{}/queue/scheduler", ";")
		Dump("system", "sestatus", "sestatus", "-v")
		Dump("system", "swapctl", "swapctl", "-s")
		Dump("system", "swapon", "swapon", "-s")
		Dump("system", "top", "top", "-n 1", "-b")
		Dump("system", "vmstat", "vmstat", "1", "10")
		Dump("system", "sys-class-net", "ls", "/sys/class/net")
		Dump("system", "proc-net-fib_trie", "cat", "/proc/net/fib_trie")

		// ¡¿ systemd stuff ¡¿
		if FileExist("/run/systemd/system") {
			logger.Info("system", "evidence of systemd present here")
			logger.Info("system", "attempting to gather systemd related information")
			Dump("system", "journalctl_dmesg", "journalctl", "--dmesg", "--no-pager")
			Dump("system", "journalctl_system", "journalctl", "--system", "--no-pager")
			Dump("system", "systemctl_all", "systemctl", "--all", "--no-pager")
			Dump("system", "systemctl_unit_files", "systemctl", "list-unit-files", "--no-pager")
		} else {
			logger.Info("system", "there is no evidence of systemd present here")
		}

		// File contents
		Dump("system", "file_var_log_daemon", "cat", "/var/log/daemon")
		Dump("system", "file_var_log_debug", "cat", "/var/log/debug")
		Dump("system", "file_etc_security_limits", "cat", "/etc/security/limits.conf")
		Dump("system", "file_var_log_kern", "cat", "/var/log/kern.log")
		Dump("system", "file_var_log_messages", "cat", "/var/log/messages")
		Dump("system", "file_var_log_syslog", "cat", "/var/log/syslog")
		Dump("system", "file_var_log_system_log", "cat", "/var/log/system.log")

		// proc entries
		Dump("system", "proc_cgroups", "cat", "/proc/cgroups")
		Dump("system", "proc_cpuinfo", "cat", "/proc/cpuinfo")
		Dump("system", "proc_diskstats", "cat", "/proc/diskstats")
		Dump("system", "proc_interrupts", "cat", "/proc/interrupts")
		Dump("system", "proc_meminfo", "cat", "/proc/meminfo")
		Dump("system", "proc_mounts", "cat", "/proc/mounts")
		Dump("system", "proc_partitions", "cat", "/proc/partitions")
		Dump("system", "proc_stat", "cat", "/proc/stat")
		Dump("system", "proc_swaps", "cat", "/proc/swaps")
		Dump("system", "proc_uptime", "cat", "/proc/uptime")
		Dump("system", "proc_version", "cat", "/proc/version")
		Dump("system", "proc_vmstat", "cat", "/proc/vmstat")
		Dump("system", "proc_sys_vm_swappiness", "cat", "/proc/sys/vm/swappiness")
	}

	// XXX: old style
	// out := "Executed system commands and stored output"
	// c.UI.Output(out)
	s.Stop()
	return 0
}

// Synopsis for command
func (c *SystemCommand) Synopsis() string {
	return "Execute system commands and store output"
}
