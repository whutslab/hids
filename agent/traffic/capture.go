package traffic

import (
	"github.com/bytedance/Elkeid/agent/plugin"
	"github.com/fsnotify/fsnotify"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

var (
	device       string = "enp0s5"
	snapshot_len int32  = 1024
	promiscuous  bool   = false
	err          error
	timeout      time.Duration = -1 * time.Second
	handle       *pcap.Handle
	packetCount  int = 0
)

type netTool struct {
	App  string
	Args []string
}

var capturecmd netTool
var Hostid string
var remoteurl string
var m sync.Map
var lck sync.Mutex

func SetHostid(remoteUrl string) {
	remoteurl = remoteUrl
	Hostid = plugin.Gethostnamereq(remoteurl + "/honeypot/myhostid")
	Hostid = "612d6dd5-2874-465f-a875-e29518c725d8"
	//fmt.Printf("fname %s tag %s ",filename,target_url)
}

func Start() {
	url := remoteurl + "/honeypot/pcaps"
	go func() {
		packetCount = 0
		for {
			nowTime := time.Now().Format("20060102150405")
			name := "./pcapfiles/honey" + nowTime + ".pcap"
			filter := "not host 192.168.137.1 and not host 192.168.137.65"
			f, _ := os.Create(name)
			w := pcapgo.NewWriter(f)
			w.WriteFileHeader((uint32)(snapshot_len), layers.LinkTypeEthernet)
			defer f.Close()
			handle, err = pcap.OpenLive(device, snapshot_len, promiscuous, timeout)
			if err != nil {
				log.Fatalf("Pcapcapture Handle error:%s", err)
			}
			err = handle.SetBPFFilter(filter)
			if err != nil {
				log.Printf("Pcapcapture SetBPFFilter error:%s", err)
			}
			defer handle.Close()
			packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
			startTime := time.Now()
			for packet := range packetSource.Packets() {
				w.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
				packetCount++
				if packetCount > 1000 || (time.Now().Sub(startTime).Seconds() >= 40 && packetCount >= 50) {
					log.Println("cpu:", GetCpuPercent(), "mem:", GetMemPercent())
					packetCount = 0
					break
				}
			}
			go plugin.Http_postPcap(name, url, Hostid, GetCpuPercent(), GetMemPercent())
		}
	}()
}

func GetCpuPercent() float64 {
	percent, _ := cpu.Percent(time.Second, false)
	return percent[0]
}

func GetMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent
}

func MonitorFile() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.Add("pcapfiles")
	if err == nil {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write && path.Ext(event.Name) == ".pcap" {
					fi, err := os.Stat(event.Name)
					if err != nil {
						log.Fatal(err)
					}
					if fi.Size() >= 1048576 {
						m.Store(event.Name, 1)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Fatal(err)
					return
				}
			}
		}
	} else {
		log.Fatal(err)
	}
}

func SendFile() {
	for {
		var del_list []string
		m.Range(func(key, value interface{}) bool {
			fname := key.(string)
			//status := value.(int)
			del_list = append(del_list, fname)
			url := remoteurl + "/honeypot/pcaps"
			//url := "http://192.168.137.1:9000/honeypot/pcaps"
			//fmt.Println(url)
			go plugin.Http_postFile(fname, url, Hostid)
			return true
		})
		for _, ind := range del_list {
			//fmt.Println("del ",ind)
			m.Delete(ind)
			err := os.Remove(ind)
			if err != nil {
				log.Fatalf("Remove pcap file failed with %s\n", err)
			}
		}
		time.Sleep(time.Second * 30)
	}
}

//go MonitorFile()
//go SendFile()
////fmt.Println(capturecmd.App,capturecmd.Args)
//cmd := exec.Command(capturecmd.App, capturecmd.Args...)
//var stdout, stderr bytes.Buffer
//cmd.Stdout = &stdout
//cmd.Stderr = &stderr
//err := cmd.Run()
//if err != nil {
//	log.Fatalf("cmd.Run() failed with %s\n", err)
//}
