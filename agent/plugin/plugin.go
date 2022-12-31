package plugin

import (
	"errors"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bytedance/Elkeid/agent/plugin/procotol"
	"github.com/bytedance/Elkeid/agent/spec"
	"github.com/prometheus/procfs"
	"github.com/tinylib/msgp/msgp"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// The time to wait before forcing the plug-in to kill,
// this is to leave the necessary time for the plugin to the clean environment normally
const exitTimeout = 1 * time.Second

// Plugin contains the process, socket, metadata and other information of a plugin
type Plugin struct {
	name       string
	version    string
	checksum   string
	cmd        *exec.Cmd
	conn       net.Conn
	runtimePID int
	pgid       int
	IO         uint64
	CPU        float64
	reader     *msgp.Reader
	Counter    atomic.Uint64
}

type Dockercfg struct {
	Docker_reglist          []string
	Docker_volume_path_list []string
	Target_fileserver       string
}

var dockercfg Dockercfg
var hostid string
var remoteurl string

// Name func returns the name of the plugin
func (p *Plugin) Name() string {
	return p.name
}

// Version func returns the version of the plugin
func (p *Plugin) Version() string {
	return p.version
}

// Checksum func returns the checksum of the plugin
func (p *Plugin) Checksum() string {
	return p.checksum
}

// PID func returns the real run pid of the plugin
func (p *Plugin) PID() int {
	return p.runtimePID
}

// Close func is used to close this plugin,
// when closing it will kill all processes under the same process group
func (p *Plugin) Close(timeout bool) {
	if p.conn != nil {
		p.conn.Close()
	}
	if timeout {
		time.Sleep(exitTimeout)
	}
	if p.pgid != 0 {
		syscall.Kill(-p.pgid, syscall.SIGKILL)
	}
	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Kill()
	}
}

func RegMatch(reglist []string, str string) bool {
	for i := range reglist {
		match, _ := regexp.MatchString(reglist[i], str)
		if match {
			return true
		}
	}
	return false
}

func SetDockercfg(docker_reglist, docker_volume_path_list []string, target_fileserver string) {
	dockercfg.Docker_reglist = docker_reglist
	dockercfg.Docker_volume_path_list = docker_volume_path_list
	dockercfg.Target_fileserver = target_fileserver
}

func SetHostid(remoteUrl string) {
	remoteurl = remoteUrl
	hostid = Gethostnamereq(remoteurl + "/honeypot/myhostid")
	//hostid = "612d6dd5-2874-465f-a875-e29518c725d8"
}

// Receive func is used to read data from the socket connection of plugin
func (p *Plugin) Receive() (*spec.Data, error) {
	data := &spec.Data{}
	res := &spec.Data{}
	//ppidargv_reglist := []string{"prlshprint"}
	//argv_reglist := []string{"sed -n s/device"}
	docker_reglist := []string{"([a-z]|[0-9]){12}"}
	//docker_reglist := []string{"ubuntu"}
	//docker_volume_path_list := []string{"/var/lib/docker/volumes/"}
	//target_fileserver := "http://10.201.111.42:8080/upload"
	err := data.DecodeMsg(p.reader)
	for i := range *data {
		if RegMatch(docker_reglist, (*data)[i]["nodename"]) {
			if (*data)[i]["data_type"] == "602" && RegMatch(dockercfg.Docker_volume_path_list, (*data)[i]["file_path"]) {
				if strings.Contains((*data)[i]["file_path"], "temp") {
					continue
				}
				go Http_postFile((*data)[i]["file_path"], dockercfg.Target_fileserver+"/honeypot/suspeciousfiles", hostid)
			} else if (*data)[i]["data_type"] == "602" && !RegMatch(dockercfg.Docker_volume_path_list, (*data)[i]["file_path"]) {
				continue
			}
			(*data)[i]["hostid"] = hostid
			(*res) = append((*res), (*data)[i])
		}
	}
	//p.Counter.Add(uint64(len(*data)))
	p.Counter.Add(uint64(len(*res)))
	return res, err
}

// Send func is used to send tasks to this plugin
func (p *Plugin) Send(t spec.Task) error {
	w := msgp.NewWriter(p.conn)
	err := t.EncodeMsg(w)
	if err != nil {
		return err
	}
	err = w.Flush()
	return err
}

func (p *Plugin) Run() error {
	if p.cmd == nil {
		return errors.New("Plugin cmd is nil")
	}
	err := p.cmd.Start()
	if err != nil {
		return err
	}
	go p.cmd.Wait()
	if p.cmd.Process == nil {
		return errors.New("Plugin cmd process is nil")
	}
	pgid, err := syscall.Getpgid(p.cmd.Process.Pid)
	if err != nil {
		return err
	}
	p.pgid = pgid
	return nil
}

func (p *Plugin) Connected() bool {
	return p.conn != nil
}

// Connect func is used to verify the connection request,
// if the pgid is inconsistent, an error will be returned
// Note that it is necessary to call Server's Delete func to clean up after this func returns error
func (p *Plugin) Connect(req procotol.RegistRequest, conn net.Conn) error {
	if p.conn != nil {
		return errors.New("The same plugin has been connected, it may be a malicious attack")
	}
	reqPgid, err := syscall.Getpgid(int(req.Pid))
	if err != nil {
		return errors.New("Cann't get req process which pid is " + strconv.FormatUint(uint64(req.Pid), 10))
	}
	cmdPgid, err := syscall.Getpgid(p.cmd.Process.Pid)
	if err != nil {
		return errors.New("Cann't get cmd process which pid is " + strconv.FormatUint(uint64(p.cmd.Process.Pid), 10))
	}
	if reqPgid != cmdPgid {
		return errors.New("Pgid does not match")
	}
	p.runtimePID = int(req.Pid)
	proc, err := procfs.NewProc(p.runtimePID)
	if err == nil {
		procIO, err := proc.IO()
		if err == nil {
			p.IO = procIO.ReadBytes + procIO.WriteBytes
		}
		procStat, err := proc.Stat()
		if err == nil {
			p.CPU = procStat.CPUTime()
		}
	}
	p.conn = conn
	p.version = req.Version
	p.name = req.Name
	p.reader = msgp.NewReaderSize(conn, 8*1024)

	return nil
}

// NewPlugin func creates a new plugin instance
func NewPlugin(name, version, checksum, runPath string) (*Plugin, error) {
	var err error
	runPath, err = filepath.Abs(runPath)
	if err != nil {
		return nil, err
	}
	dir, file := path.Split(runPath)
	zap.S().Infof("Plugin work directory: %s", dir)
	c := exec.Command(runPath)
	c.Dir = dir
	c.Stderr, err = os.OpenFile(dir+file+".stderr", os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		return nil, err
	}
	c.Stdin = nil
	c.Stdout, err = os.OpenFile(dir+file+".stdout", os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		return nil, err
	}
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}
	p := Plugin{cmd: c, name: name, version: version, checksum: checksum}
	return &p, nil
}
