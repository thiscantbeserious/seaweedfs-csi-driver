package driver

import (
	"fmt"
	"time"

	"github.com/chrislusf/seaweedfs/weed/glog"
	"os/exec"
	"k8s.io/utils/mount"
)

// Config holds values to configure the driver
type Config struct {
	// Region          string
	Filer string
}

type Mounter interface {
	Mount(target string) error
}

func newMounter(bucketName string, filer string) (Mounter, error) {
	return newSeaweedFsMounter(bucketName, filer)
}

func fuseMount(path string, command string, args []string) error {
	cmd := exec.Command(command, args...)
	glog.V(3).Infof("Mounting fuse with command: %s and args: %s", command, args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error fuseMount command: %s\nargs: %s\noutput: %s", command, args, out)
	}

	return waitForMount(path, 10*time.Second)
}

func fuseUnmount(path string) error {
	if err := mount.New("").Unmount(path); err != nil {
		return err
	}
	// as fuse quits immediately, we will try to wait until the process is done
	process, err := findFuseMountProcess(path)
	if err != nil {
		glog.Errorf("Error getting PID of fuse mount: %s", err)
		return nil
	}
	if process == nil {
		glog.Warningf("Unable to find PID of fuse mount %s, it must have finished already", path)
		return nil
	}
	glog.Infof("Found fuse pid %v of mount %s, checking if it still runs", process.Pid, path)
	return waitForProcess(process, 1)
}

func newConfigFromSecrets(secrets map[string]string) *Config {
	t := &Config{
		Filer: secrets["filer"],
	}
	return t
}
