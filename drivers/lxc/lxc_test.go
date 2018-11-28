//+build linux,lxc

package lxc

import (
	"testing"

	"github.com/hashicorp/nomad/helper/testlog"
	"github.com/hashicorp/nomad/helper/uuid"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/stretchr/testify/require"
)

func TestLXCDriver_Mounts(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	task := &drivers.TaskConfig{
		ID:   uuid.Generate(),
		Name: "test",
		Resources: &drivers.Resources{
			NomadResources: &structs.Resources{
				CPU:      1,
				MemoryMB: 2,
			},
			LinuxResources: &drivers.LinuxResources{
				CPUShares:        1024,
				MemoryLimitBytes: 2 * 1024,
			},
		},
		Mounts: []*drivers.MountConfig{
			{HostPath: "/dev", TaskPath: "/task-mounts/dev-path"},
			{HostPath: "/bin/sh", TaskPath: "/task-mounts/task-path-ro", Readonly: true},
		},
		Devices: []*drivers.DeviceConfig{
			{HostPath: "/dev", TaskPath: "/task-devices/dev-path", Permissions: "rw"},
			{HostPath: "/bin/sh", TaskPath: "/task-devices/task-path-ro", Permissions: "ro"},
		},
	}
	taskConfig := TaskConfig{
		Template: "busybox",
		Volumes: []string{
			"relative/path:/usr-config/container/path",
			"relative/path2:usr-config/container/relative",
		},
	}

	d := NewLXCDriver(testlog.HCLogger(t)).(*Driver)
	d.config.Enabled = true

	entries, err := d.mountEntries(task, taskConfig)
	require.NoError(err)

	expectedEntries := []string{
		"test/relative/path usr-config/container/path none rw,bind,create=dir",
		"test/relative/path2 usr-config/container/relative none rw,bind,create=dir",
		"/dev task-mounts/dev-path none rw,bind,create=dir",
		"/bin/sh task-mounts/task-path-ro none ro,bind,create=file",
		"/dev task-devices/dev-path none rw,bind,create=dir",
		"/bin/sh task-devices/task-path-ro none ro,bind,create=file",
	}

	for _, e := range expectedEntries {
		require.Contains(entries, e)
	}
}
