// Generated file, do not edit.

// Copyright 2024 ko Build Authors All Rights Reserved.
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

package caps

var ddTests = []ddTest{
	{permitted: "chown", inheritable: "", effective: false, res: "AAAAAgEAAAAAAAAAAAAAAAAAAAA="},
	{permitted: "chown", inheritable: "", effective: true, res: "AQAAAgEAAAAAAAAAAAAAAAAAAAA="},
	{permitted: "", inheritable: "chown", effective: false, res: "AAAAAgAAAAABAAAAAAAAAAAAAAA="},
	{permitted: "chown", inheritable: "chown", effective: true, res: "AQAAAgEAAAABAAAAAAAAAAAAAAA="},
	{permitted: "dac_override", inheritable: "dac_override", effective: true, res: "AQAAAgIAAAACAAAAAAAAAAAAAAA="},
	{permitted: "dac_read_search", inheritable: "dac_read_search", effective: true, res: "AQAAAgQAAAAEAAAAAAAAAAAAAAA="},
	{permitted: "fowner", inheritable: "fowner", effective: true, res: "AQAAAggAAAAIAAAAAAAAAAAAAAA="},
	{permitted: "fsetid", inheritable: "fsetid", effective: true, res: "AQAAAhAAAAAQAAAAAAAAAAAAAAA="},
	{permitted: "kill", inheritable: "kill", effective: true, res: "AQAAAiAAAAAgAAAAAAAAAAAAAAA="},
	{permitted: "setgid", inheritable: "setgid", effective: true, res: "AQAAAkAAAABAAAAAAAAAAAAAAAA="},
	{permitted: "setuid", inheritable: "setuid", effective: true, res: "AQAAAoAAAACAAAAAAAAAAAAAAAA="},
	{permitted: "setpcap", inheritable: "setpcap", effective: true, res: "AQAAAgABAAAAAQAAAAAAAAAAAAA="},
	{permitted: "linux_immutable", inheritable: "linux_immutable", effective: true, res: "AQAAAgACAAAAAgAAAAAAAAAAAAA="},
	{permitted: "net_bind_service", inheritable: "net_bind_service", effective: true, res: "AQAAAgAEAAAABAAAAAAAAAAAAAA="},
	{permitted: "net_broadcast", inheritable: "net_broadcast", effective: true, res: "AQAAAgAIAAAACAAAAAAAAAAAAAA="},
	{permitted: "net_admin", inheritable: "net_admin", effective: true, res: "AQAAAgAQAAAAEAAAAAAAAAAAAAA="},
	{permitted: "net_raw", inheritable: "net_raw", effective: true, res: "AQAAAgAgAAAAIAAAAAAAAAAAAAA="},
	{permitted: "ipc_lock", inheritable: "ipc_lock", effective: true, res: "AQAAAgBAAAAAQAAAAAAAAAAAAAA="},
	{permitted: "ipc_owner", inheritable: "ipc_owner", effective: true, res: "AQAAAgCAAAAAgAAAAAAAAAAAAAA="},
	{permitted: "sys_module", inheritable: "sys_module", effective: true, res: "AQAAAgAAAQAAAAEAAAAAAAAAAAA="},
	{permitted: "sys_rawio", inheritable: "sys_rawio", effective: true, res: "AQAAAgAAAgAAAAIAAAAAAAAAAAA="},
	{permitted: "sys_chroot", inheritable: "sys_chroot", effective: true, res: "AQAAAgAABAAAAAQAAAAAAAAAAAA="},
	{permitted: "sys_ptrace", inheritable: "sys_ptrace", effective: true, res: "AQAAAgAACAAAAAgAAAAAAAAAAAA="},
	{permitted: "sys_pacct", inheritable: "sys_pacct", effective: true, res: "AQAAAgAAEAAAABAAAAAAAAAAAAA="},
	{permitted: "sys_admin", inheritable: "sys_admin", effective: true, res: "AQAAAgAAIAAAACAAAAAAAAAAAAA="},
	{permitted: "sys_boot", inheritable: "sys_boot", effective: true, res: "AQAAAgAAQAAAAEAAAAAAAAAAAAA="},
	{permitted: "sys_nice", inheritable: "sys_nice", effective: true, res: "AQAAAgAAgAAAAIAAAAAAAAAAAAA="},
	{permitted: "sys_resource", inheritable: "sys_resource", effective: true, res: "AQAAAgAAAAEAAAABAAAAAAAAAAA="},
	{permitted: "sys_time", inheritable: "sys_time", effective: true, res: "AQAAAgAAAAIAAAACAAAAAAAAAAA="},
	{permitted: "sys_tty_config", inheritable: "sys_tty_config", effective: true, res: "AQAAAgAAAAQAAAAEAAAAAAAAAAA="},
	{permitted: "mknod", inheritable: "mknod", effective: true, res: "AQAAAgAAAAgAAAAIAAAAAAAAAAA="},
	{permitted: "lease", inheritable: "lease", effective: true, res: "AQAAAgAAABAAAAAQAAAAAAAAAAA="},
	{permitted: "audit_write", inheritable: "audit_write", effective: true, res: "AQAAAgAAACAAAAAgAAAAAAAAAAA="},
	{permitted: "audit_control", inheritable: "audit_control", effective: true, res: "AQAAAgAAAEAAAABAAAAAAAAAAAA="},
	{permitted: "setfcap", inheritable: "setfcap", effective: true, res: "AQAAAgAAAIAAAACAAAAAAAAAAAA="},
	{permitted: "mac_override", inheritable: "mac_override", effective: true, res: "AQAAAgAAAAAAAAAAAQAAAAEAAAA="},
	{permitted: "mac_admin", inheritable: "mac_admin", effective: true, res: "AQAAAgAAAAAAAAAAAgAAAAIAAAA="},
	{permitted: "syslog", inheritable: "syslog", effective: true, res: "AQAAAgAAAAAAAAAABAAAAAQAAAA="},
	{permitted: "wake_alarm", inheritable: "wake_alarm", effective: true, res: "AQAAAgAAAAAAAAAACAAAAAgAAAA="},
	{permitted: "block_suspend", inheritable: "block_suspend", effective: true, res: "AQAAAgAAAAAAAAAAEAAAABAAAAA="},
	{permitted: "audit_read", inheritable: "audit_read", effective: true, res: "AQAAAgAAAAAAAAAAIAAAACAAAAA="},
	{permitted: "perfmon", inheritable: "perfmon", effective: true, res: "AQAAAgAAAAAAAAAAQAAAAEAAAAA="},
	{permitted: "bpf", inheritable: "bpf", effective: true, res: "AQAAAgAAAAAAAAAAgAAAAIAAAAA="},
	{permitted: "checkpoint_restore", inheritable: "checkpoint_restore", effective: true, res: "AQAAAgAAAAAAAAAAAAEAAAABAAA="},
}
