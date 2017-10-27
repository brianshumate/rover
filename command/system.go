// Package command for system
// System executes operating system commands and stores output
package command

import (
	"fmt"
	"github.com/brianshumate/rover/internal"
	"github.com/mitchellh/cli"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Darwin OS
const Darwin string = "darwin"

// FreeBSD OS
const FreeBSD string = "freebsd"

// Linux OS
const Linux string = "linux"

// SystemCommand describes system related fields
type SystemCommand struct {
	HostName     string
	OS           string
	ReleaseFiles []string
	UI           cli.Ui
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

	// Internal logging
	internal.LogSetup()

	c.HostName = internal.GetHostName()
	c.OS = runtime.GOOS
	c.ReleaseFiles = []string{"/etc/redhat-release",
		"/etc/fedora-release",
		"/etc/slackware-release",
		"/etc/debian_release",
		"/etc/os-release"}

	log.Printf("[i] Hello from the rover system module on %s!", c.HostName)

	log.Printf("[i] The OS detected is %s", c.OS)

	// Handle creating the command output directory
	// TODO: maybe not such a hardcoded path?
	outPath := filepath.Join(".", fmt.Sprintf("%s/system", c.HostName))
	if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
		out := fmt.Sprintf("[e] Cannot create directory %s!", outPath)
		log.Println(out)
		c.UI.Error(out)
		os.Exit(-1)
	}

	// Grab OS release info
	for i, file := range c.ReleaseFiles {
		if internal.FileExist(file) {
			internal.Dump("system", "os_release", "cat", file)
			log.Printf("[i] Dumping release file %s %v", file, i)
		}
	}

	// Common commands and file contents from which to gather information
	// across the currently supported range of operating system commands
	internal.Dump("system", "date", "date")
	internal.Dump("system", "df", "df")
	internal.Dump("system", "df_i", "df", "-i")
	internal.Dump("system", "df_h", "df", "-h")
	internal.Dump("system", "dmesg", "dmesg")
	internal.Dump("system", "hostname", "hostname")
	internal.Dump("system", "last", "last")
	internal.Dump("system", "mount", "mount")
	internal.Dump("system", "netstat_anW", "netstat", "-anW")
	internal.Dump("system", "netstat_indW", "netstat", "-indW")
	internal.Dump("system", "netstat_mmmW", "netstat", "-mmmW")
	internal.Dump("system", "netstat_nralW", "netstat", "-nralW")
	internal.Dump("system", "netstat_rn", "netstat", "-rn")
	internal.Dump("system", "netstat_sW", "netstat", "-sW")
	internal.Dump("system", "pfctl_rules", "pfctl", "-s rules")
	internal.Dump("system", "pfctl_nat", "pfctl", "-s nat")
	internal.Dump("system", "sysctl", "sysctl", "-a")
	internal.Dump("system", "uname", "uname", "-a")
	internal.Dump("system", "w", "w")

	// File contents
	internal.Dump("system", "file_etc_fstab", "cat", "/etc/fstab")
	internal.Dump("system", "file_etc_hosts", "cat", "/etc/hosts")
	internal.Dump("system", "file_etc_resolv_conf", "cat", "/etc/resolv.conf")

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
		internal.Dump("system", "ifconfig", "ifconfig", "-a")
		internal.Dump("system", "netstat_rs", "netstat", "-rs")
		internal.Dump("system", "ps", "ps", "aux")
		internal.Dump("system", "top", "top", "-l 1")
		internal.Dump("system", "vm_stat", "vm_stat")

	case FreeBSD:

		// FreeBSD specific ommands
		internal.Dump("system", "arp_a", "arp", "-a")
		internal.Dump("system", "ifconfig", "ifconfig", "-a")
		internal.Dump("system", "iostat_bsd", "iostat", "-c 10")
		internal.Dump("system", "pkg_info", "pkg", "info")
		internal.Dump("system", "ps", "ps", "aux")
		internal.Dump("system", "swapinfo", "swapinfo")
		internal.Dump("system", "sysctl", "sysctl", "-a")
		internal.Dump("system", "top", "top", "-n", "-b")
		internal.Dump("system", "vmstat", "vmstat", "1", "10")

		// File contents
		internal.Dump("system", "file_var_run_dmesg_boot", "cat", "/var/run/dmesg.boot")
		internal.Dump("system", "file_var_log_messages", "cat", "/var/log/messages")
		internal.Dump("system", "file_etc_rc_conf", "cat", "/etc/rc.conf")
		internal.Dump("system", "file_etc_sysctl_conf", "cat", "/etc/sysctl.conf")

	case Linux:

		// Linux specific commands
		internal.Dump("system", "bonding", "find", "/proc/net/bonding/", "-type", "f", "-print", "-exec", "cat", "{}", ";")
		internal.Dump("system", "disk_by_id", "ls", "-l", "/dev/disk/by-id")
		internal.Dump("system", "dmesg", "dmesg")
		internal.Dump("system", "dpkg", "dpkg", "-l")
		internal.Dump("system", "free", "free", "-m")
		internal.Dump("system", "ifconfig", "ifconfig", "-a")
		internal.Dump("system", "iostat_linux", "iostat", "-mx", "1", "10")
		internal.Dump("system", "ip_addr", "ip", "addr")
		internal.Dump("system", "lsb_release", "lsb_release")
		internal.Dump("system", "ps", "ps", "-aux")
		internal.Dump("system", "rpm", "rpm", "-qa")
		internal.Dump("system", "rx_crc_errors", "find", "/sys/class/net/", "-type", "l", "-print", "-exec", "cat", "{}/statistics/rx_crc_errors", ";")
		internal.Dump("system", "schedulers", "find", "/sys/block/", "-type", "l", "-print", "-exec", "cat", "{}/queue/scheduler", ";")
		internal.Dump("system", "sestatus", "sestatus", "-v")
		internal.Dump("system", "swapctl", "swapctl", "-s")
		internal.Dump("system", "swapon", "swapon", "-s")
		internal.Dump("system", "top", "top", "-n 1", "-b")
		internal.Dump("system", "vmstat", "vmstat", "1", "10")

		// ¡¿ systemd stuff ¡¿
		if internal.FileExist("/run/systemd/system") {
			log.Println("[i] There is evidence of systemd present here")
			log.Println("[i] Attempting to gather systemd related information ...")
			internal.Dump("system", "journalctl_dmesg", "journalctl", "--dmesg", "--no-pager")
			internal.Dump("system", "journalctl_system", "journalctl", "--system", "--no-pager")
			internal.Dump("system", "systemctl_all", "systemctl", "--all", "--no-pager")
			internal.Dump("system", "systemctl_unit_files", "systemctl", "list-unit-files", "--no-pager")
		} else {
			log.Println("[i] There is no evidence of systemd present here")
		}

		// File contents
		internal.Dump("system", "file_var_log_daemon", "cat", "/var/log/daemon")
		internal.Dump("system", "file_var_log_debug", "cat", "/var/log/debug")
		internal.Dump("system", "file_etc_security_limits", "cat", "/etc/security/limits.conf")
		internal.Dump("system", "file_var_log_kern", "cat", "/var/log/kern.log")
		internal.Dump("system", "file_var_log_messages", "cat", "/var/log/messages")
		internal.Dump("system", "file_var_log_syslog", "cat", "/var/log/syslog")
		internal.Dump("system", "file_var_log_system_log", "cat", "/var/log/system.log")

		// proc entries
		internal.Dump("system", "proc_cgroups", "cat", "/proc/cgroups")
		internal.Dump("system", "proc_cpuinfo", "cat", "/proc/cpuinfo")
		internal.Dump("system", "proc_diskstats", "cat", "/proc/diskstats")
		internal.Dump("system", "proc_interrupts", "cat", "/proc/interrupts")
		internal.Dump("system", "proc_meminfo", "cat", "/proc/meminfo")
		internal.Dump("system", "proc_mounts", "cat", "/proc/mounts")
		internal.Dump("system", "proc_partitions", "cat", "/proc/partitions")
		internal.Dump("system", "proc_stat", "cat", "/proc/stat")
		internal.Dump("system", "proc_swaps", "cat", "/proc/swaps")
		internal.Dump("system", "proc_uptime", "cat", "/proc/uptime")
		internal.Dump("system", "proc_version", "cat", "/proc/version")
		internal.Dump("system", "proc_vmstat", "cat", "/proc/vmstat")
		internal.Dump("system", "proc_sys_vm_swappiness", "cat", "/proc/sys/vm/swappiness")
	}

	out := "Executed system commands and stored output"
	c.UI.Output(out)

	return 0
}

// Synopsis for command
func (c *SystemCommand) Synopsis() string {
	return "Execute system commands and store output"
}
