// Copyright 2023 ko Build Authors All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package build

import "strings"

type Cap int

// POSIX-draft defined capabilities.
const (
	// In a system with the [_POSIX_CHOWN_RESTRICTED] option defined, this
	// overrides the restriction of changing file ownership and group
	// ownership.
	CAP_CHOWN = Cap(0)

	// Override all DAC access, including ACL execute access if
	// [_POSIX_ACL] is defined. Excluding DAC access covered by
	// CAP_LINUX_IMMUTABLE.
	CAP_DAC_OVERRIDE = Cap(1)

	// Overrides all DAC restrictions regarding read and search on files
	// and directories, including ACL restrictions if [_POSIX_ACL] is
	// defined. Excluding DAC access covered by CAP_LINUX_IMMUTABLE.
	CAP_DAC_READ_SEARCH = Cap(2)

	// Overrides all restrictions about allowed operations on files, where
	// file owner ID must be equal to the user ID, except where CAP_FSETID
	// is applicable. It doesn't override MAC and DAC restrictions.
	CAP_FOWNER = Cap(3)

	// Overrides the following restrictions that the effective user ID
	// shall match the file owner ID when setting the S_ISUID and S_ISGID
	// bits on that file; that the effective group ID (or one of the
	// supplementary group IDs) shall match the file owner ID when setting
	// the S_ISGID bit on that file; that the S_ISUID and S_ISGID bits are
	// cleared on successful return from chown(2) (not implemented).
	CAP_FSETID = Cap(4)

	// Overrides the restriction that the real or effective user ID of a
	// process sending a signal must match the real or effective user ID
	// of the process receiving the signal.
	CAP_KILL = Cap(5)

	// Allows setgid(2) manipulation
	// Allows setgroups(2)
	// Allows forged gids on socket credentials passing.
	CAP_SETGID = Cap(6)

	// Allows set*uid(2) manipulation (including fsuid).
	// Allows forged pids on socket credentials passing.
	CAP_SETUID = Cap(7)

	// Linux-specific capabilities

	// Without VFS support for capabilities:
	//   Transfer any capability in your permitted set to any pid,
	//   remove any capability in your permitted set from any pid
	// With VFS support for capabilities (neither of above, but)
	//   Add any capability from current's capability bounding set
	//     to the current process' inheritable set
	//   Allow taking bits out of capability bounding set
	//   Allow modification of the securebits for a process
	CAP_SETPCAP = Cap(8)

	// Allow modification of S_IMMUTABLE and S_APPEND file attributes
	CAP_LINUX_IMMUTABLE = Cap(9)

	// Allows binding to TCP/UDP sockets below 1024
	// Allows binding to ATM VCIs below 32
	CAP_NET_BIND_SERVICE = Cap(10)

	// Allow broadcasting, listen to multicast
	CAP_NET_BROADCAST = Cap(11)

	// Allow interface configuration
	// Allow administration of IP firewall, masquerading and accounting
	// Allow setting debug option on sockets
	// Allow modification of routing tables
	// Allow setting arbitrary process / process group ownership on
	// sockets
	// Allow binding to any address for transparent proxying (also via NET_RAW)
	// Allow setting TOS (type of service)
	// Allow setting promiscuous mode
	// Allow clearing driver statistics
	// Allow multicasting
	// Allow read/write of device-specific registers
	// Allow activation of ATM control sockets
	CAP_NET_ADMIN = Cap(12)

	// Allow use of RAW sockets
	// Allow use of PACKET sockets
	// Allow binding to any address for transparent proxying (also via NET_ADMIN)
	CAP_NET_RAW = Cap(13)

	// Allow locking of shared memory segments
	// Allow mlock and mlockall (which doesn't really have anything to do
	// with IPC)
	CAP_IPC_LOCK = Cap(14)

	// Override IPC ownership checks
	CAP_IPC_OWNER = Cap(15)

	// Insert and remove kernel modules - modify kernel without limit
	CAP_SYS_MODULE = Cap(16)

	// Allow ioperm/iopl access
	// Allow sending USB messages to any device via /proc/bus/usb
	CAP_SYS_RAWIO = Cap(17)

	// Allow use of chroot()
	CAP_SYS_CHROOT = Cap(18)

	// Allow ptrace() of any process
	CAP_SYS_PTRACE = Cap(19)

	// Allow configuration of process accounting
	CAP_SYS_PACCT = Cap(20)

	// Allow configuration of the secure attention key
	// Allow administration of the random device
	// Allow examination and configuration of disk quotas
	// Allow setting the domainname
	// Allow setting the hostname
	// Allow calling bdflush()
	// Allow mount() and umount(), setting up new smb connection
	// Allow some autofs root ioctls
	// Allow nfsservctl
	// Allow VM86_REQUEST_IRQ
	// Allow to read/write pci config on alpha
	// Allow irix_prctl on mips (setstacksize)
	// Allow flushing all cache on m68k (sys_cacheflush)
	// Allow removing semaphores
	// Used instead of CAP_CHOWN to "chown" IPC message queues, semaphores
	// and shared memory
	// Allow locking/unlocking of shared memory segment
	// Allow turning swap on/off
	// Allow forged pids on socket credentials passing
	// Allow setting readahead and flushing buffers on block devices
	// Allow setting geometry in floppy driver
	// Allow turning DMA on/off in xd driver
	// Allow administration of md devices (mostly the above, but some
	// extra ioctls)
	// Allow tuning the ide driver
	// Allow access to the nvram device
	// Allow administration of apm_bios, serial and bttv (TV) device
	// Allow manufacturer commands in isdn CAPI support driver
	// Allow reading non-standardized portions of pci configuration space
	// Allow DDI debug ioctl on sbpcd driver
	// Allow setting up serial ports
	// Allow sending raw qic-117 commands
	// Allow enabling/disabling tagged queuing on SCSI controllers and sending
	// arbitrary SCSI commands
	// Allow setting encryption key on loopback filesystem
	// Allow setting zone reclaim policy
	CAP_SYS_ADMIN = Cap(21)

	// Allow use of reboot()
	CAP_SYS_BOOT = Cap(22)

	// Allow raising priority and setting priority on other (different
	// UID) processes
	// Allow use of FIFO and round-robin (realtime) scheduling on own
	// processes and setting the scheduling algorithm used by another
	// process.
	// Allow setting cpu affinity on other processes
	CAP_SYS_NICE = Cap(23)

	// Override resource limits. Set resource limits.
	// Override quota limits.
	// Override reserved space on ext2 filesystem
	// Modify data journaling mode on ext3 filesystem (uses journaling
	// resources)
	// NOTE: ext2 honors fsuid when checking for resource overrides, so
	// you can override using fsuid too
	// Override size restrictions on IPC message queues
	// Allow more than 64hz interrupts from the real-time clock
	// Override max number of consoles on console allocation
	// Override max number of keymaps
	CAP_SYS_RESOURCE = Cap(24)

	// Allow manipulation of system clock
	// Allow irix_stime on mips
	// Allow setting the real-time clock
	CAP_SYS_TIME = Cap(25)

	// Allow configuration of tty devices
	// Allow vhangup() of tty
	CAP_SYS_TTY_CONFIG = Cap(26)

	// Allow the privileged aspects of mknod()
	CAP_MKNOD = Cap(27)

	// Allow taking of leases on files
	CAP_LEASE = Cap(28)

	CAP_AUDIT_WRITE   = Cap(29)
	CAP_AUDIT_CONTROL = Cap(30)
	CAP_SETFCAP       = Cap(31)

	// Override MAC access.
	// The base kernel enforces no MAC policy.
	// An LSM may enforce a MAC policy, and if it does and it chooses
	// to implement capability based overrides of that policy, this is
	// the capability it should use to do so.
	CAP_MAC_OVERRIDE = Cap(32)

	// Allow MAC configuration or state changes.
	// The base kernel requires no MAC configuration.
	// An LSM may enforce a MAC policy, and if it does and it chooses
	// to implement capability based checks on modifications to that
	// policy or the data required to maintain it, this is the
	// capability it should use to do so.
	CAP_MAC_ADMIN = Cap(33)

	// Allow configuring the kernel's syslog (printk behaviour)
	CAP_SYSLOG = Cap(34)

	// Allow triggering something that will wake the system
	CAP_WAKE_ALARM = Cap(35)

	// Allow preventing system suspends
	CAP_BLOCK_SUSPEND = Cap(36)

	// Allow reading audit messages from the kernel
	CAP_AUDIT_READ = Cap(37)
)

func CapFromString(value string) Cap {
	switch strings.ToUpper(value) {
	case "CAP_CHOWN":
		return CAP_CHOWN // 0
	case "CAP_DAC_OVERRIDE":
		return CAP_DAC_OVERRIDE // 1
	case "CAP_DAC_READ_SEARCH":
		return CAP_DAC_READ_SEARCH // 2
	case "CAP_FOWNER":
		return CAP_FOWNER // 3
	case "CAP_FSETID":
		return CAP_FSETID // 4
	case "CAP_KILL":
		return CAP_KILL // 5
	case "CAP_SETGID":
		return CAP_SETGID // 6
	case "CAP_SETUID":
		return CAP_SETUID // 7
	case "CAP_SETPCAP":
		return CAP_SETPCAP // 8
	case "CAP_LINUX_IMMUTABLE":
		return CAP_LINUX_IMMUTABLE // 9
	case "CAP_NET_BIND_SERVICE":
		return CAP_NET_BIND_SERVICE // 10
	case "CAP_NET_BROADCAST":
		return CAP_NET_BROADCAST // 11
	case "CAP_NET_ADMIN":
		return CAP_NET_ADMIN // 12
	case "CAP_NET_RAW":
		return CAP_NET_RAW // 13
	case "CAP_IPC_LOCK":
		return CAP_IPC_LOCK // 14
	case "CAP_IPC_OWNER":
		return CAP_IPC_OWNER // 15
	case "CAP_SYS_MODULE":
		return CAP_SYS_MODULE // 16
	case "CAP_SYS_RAWIO":
		return CAP_SYS_RAWIO // 17
	case "CAP_SYS_CHROOT":
		return CAP_SYS_CHROOT // 18
	case "CAP_SYS_PTRACE":
		return CAP_SYS_PTRACE // 19
	case "CAP_SYS_PACCT":
		return CAP_SYS_PACCT // 20
	case "CAP_SYS_ADMIN":
		return CAP_SYS_ADMIN // 21
	case "CAP_SYS_BOOT":
		return CAP_SYS_BOOT // 22
	case "CAP_SYS_NICE":
		return CAP_SYS_NICE // 23
	case "CAP_SYS_RESOURCE":
		return CAP_SYS_RESOURCE // 24
	case "CAP_SYS_TIME":
		return CAP_SYS_TIME // 25
	case "CAP_SYS_TTY_CONFIG":
		return CAP_SYS_TTY_CONFIG // 26
	case "CAP_MKNOD":
		return CAP_MKNOD // 27
	case "CAP_LEASE":
		return CAP_LEASE // 28
	case "CAP_AUDIT_WRITE":
		return CAP_AUDIT_WRITE // 29
	case "CAP_AUDIT_CONTROL":
		return CAP_AUDIT_CONTROL // 30
	case "CAP_SETFCAP":
		return CAP_SETFCAP // 31
	case "CAP_MAC_OVERRIDE":
		return CAP_MAC_OVERRIDE // 32
	case "CAP_MAC_ADMIN":
		return CAP_MAC_ADMIN // 33
	case "CAP_SYSLOG":
		return CAP_SYSLOG // 34
	case "CAP_WAKE_ALARM":
		return CAP_WAKE_ALARM // 35
	case "CAP_BLOCK_SUSPEND":
		return CAP_BLOCK_SUSPEND // 36
	case "CAP_AUDIT_READ":
		return CAP_AUDIT_READ // 37
	default:
		return -1
	}
}
