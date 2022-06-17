package main

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// checksum all file in folder

// A result is the product of reading and summing a file using MD5.
type result struct {
	path string
	sum  [md5.Size]byte
	err  error
}

// sumFiles starts goroutines to walk the directory tree at root and digest each
// regular file.  These goroutines send the results of the digests on the result
// channel and send the result of the walk on the error channel.  If done is
// closed, sumFiles abandons its work.
func sumFiles(done <-chan struct{}, root string) (<-chan result, <-chan error) {
	// For each regular file, start a goroutine that sums the file and sends
	// the result on c.  Send the result of the walk on errc.
	c := make(chan result)
	errc := make(chan error, 1)
	go func() { // HL
		var wg sync.WaitGroup
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			wg.Add(1)
			go func() { // HL
				data, err := ioutil.ReadFile(path)
				select {
				case c <- result{path, md5.Sum(data), err}: // HL
				case <-done: // HL
				}
				wg.Done()
			}()
			// Abort the walk if done is closed.
			select {
			case <-done: // HL
				return errors.New("walk canceled")
			default:
				return nil
			}
		})
		// Walk has returned, so all calls to wg.Add are done.  Start a
		// goroutine to close c once all the sends are done.
		go func() { // HL
			wg.Wait()
			close(c) // HL
		}()
		// No select needed here, since errc is buffered.
		errc <- err // HL
	}()
	return c, errc
}

// MD5All reads all the files in the file tree rooted at root and returns a map
// from file path to the MD5 sum of the file's contents.  If the directory walk
// fails or any read operation fails, MD5All returns an error.  In that case,
// MD5All does not wait for inflight read operations to complete.
func MD5All(root string) (map[string][md5.Size]byte, map[string]string, error) {
	// MD5All closes the done channel when it returns; it may do so before
	// receiving all the values from c and errc.
	done := make(chan struct{}) // HLdone
	defer close(done)           // HLdone

	c, errc := sumFiles(done, root) // HLdone

	m := make(map[string][md5.Size]byte)
	m2 := make(map[string]string)
	for r := range c { // HLrange
		if r.err != nil {
			return nil, nil, r.err
		}
		m[r.path] = r.sum
		arr := r.sum
		m2[string(arr[:])] = r.path
	}
	if err := <-errc; err != nil {
		return nil, nil, err
	}
	return m, m2, nil
}

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "librum-pubsub"

// start subsriber to topic
func subscribe(subscriber *pubsub.Subscription, ctx context.Context, hostID peer.ID) []byte {
	for {
		msg, err := subscriber.Next(ctx)
		if err != nil {
			panic(err)
		}

		// only consider messages delivered by other peers
		if msg.ReceivedFrom == hostID {
			continue
		}

		fmt.Printf("got message, from: %s\n", msg.ReceivedFrom.Pretty())
		//fmt.Println(msg.Data)
		return msg.Data
	}
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID.Pretty())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	}
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}

// Main
func main() {
	//Calculate the MD5 sum of all files under the specified directory,
	m, m2, err := MD5All("/home/ubuntu/golang/Code/Code_Golang/repost-week-1/checksum_sync/test_folder2") //os.Args[1]
	if err != nil {
		fmt.Println(err)
		return
	}

	// tranfer MD5 to slice
	var paths []string
	for path := range m {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var SlcMD5 []string
	for _, path := range paths {
		arr := m[path]
		strMD5 := string(arr[:])
		SlcMD5 = append(SlcMD5, strMD5)
	}

	ctx := context.Background()
	// create a new libp2p Host that listens on a random TCP port
	// we can specify port like /ip4/0.0.0.0/tcp/3326

	host, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		panic(err)
	}

	// view host details and addresses
	fmt.Printf("host ID %s\n", host.ID().Pretty())
	fmt.Printf("following are the assigned addresses\n")
	for _, addr := range host.Addrs() {
		fmt.Printf("%s\n", addr.String())
	}
	fmt.Printf("\n")

	// create a new PubSub service using the GossipSub router
	gossipSub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		panic(err)
	}

	// setup local mDNS discovery
	if err := setupDiscovery(host); err != nil {
		panic(err)
	}

	// join the pubsub topic
	room := "filename"
	room2 := "MD5"
	room3 := "SendingFile"
	topic, err := gossipSub.Join(room)
	if err != nil {
		panic(err)
	}
	topic2, err := gossipSub.Join(room2)
	if err != nil {
		panic(err)
	}
	topic3, err := gossipSub.Join(room3)
	if err != nil {
		panic(err)
	}
	FileName := scan_filename()
	//publish name and checksum of all file to topic
	go publish_All(ctx, topic, topic2)
	//subcribe from topic
	go subcribe_All_N_Publish(ctx, topic, topic2, topic3, host.ID(), FileName, SlcMD5, m2)
}

// ----------------------publish function group---------------------------------
func publish_All(ctx context.Context, topic *pubsub.Topic, topic2 *pubsub.Topic) {
	for {
		scan_N_pub(ctx, topic)
		check_N_Pub(ctx, topic2)
		time.Sleep(30 * time.Second)
	}
}
func check_N_Pub(ctx context.Context, topic *pubsub.Topic) {
	m, _, err := MD5All("/home/ubuntu/golang/Code/Code_Golang/repost-week-1/checksum_sync/test_folder2") //os.Args[1]
	if err != nil {
		fmt.Println(err)
		return
	}
	var paths []string
	for path := range m {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	//var slc []string
	for _, path := range paths {
		var MD5arr [16]byte = m[path]
		var MD5slc []byte = MD5arr[0:16]
		topic.Publish(ctx, MD5slc)
		time.Sleep(5 * time.Second)
	}
	var arr0 [16]byte
	topic.Publish(ctx, arr0[0:16])
	fmt.Println("sending last message ... ")
}
func scan_filename() []string {
	dirname := "/home/ubuntu/golang/Code/Code_Golang/repost-week-1/checksum_sync/test_folder2"
	var fileName []string
	f, err := os.Open(dirname)
	if err != nil {
		log.Fatal(err)
	}
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fileName = append(fileName, file.Name())
	}
	sort.Strings(fileName)
	return fileName
}
func scan_N_pub(ctx context.Context, topic *pubsub.Topic) {
	fileName := scan_filename()
	var dataSend []byte
	for nameStr := range fileName {
		nameByte := byte(nameStr)
		dataSend = append(dataSend, nameByte)
	}
	topic.Publish(ctx, dataSend)
	topic.Publish(ctx, nil)
}

// -------------------------subcribe function group-------------------------------------------
func subcribe_All_N_Publish(ctx context.Context, topic *pubsub.Topic, topic2 *pubsub.Topic, topic3 *pubsub.Topic, hostID peer.ID, FileName []string, SlcMD5 []string, m2 map[string]string) {
	for {
		GetFileName := get_fileName(ctx, topic, hostID)
		GetMD5 := get_MD5(ctx, topic2, hostID)
		compare_Map(FileName, GetFileName, SlcMD5, GetMD5, m2)
	}
}
func get_fileName(ctx context.Context, topic *pubsub.Topic, hostID peer.ID) []string {
	subscriber, err := topic.Subscribe()
	if err != nil {
		panic(err)
	}

	var fileNames []string

	for {
		DataRecieve := subscribe(subscriber, ctx, hostID)
		if DataRecieve != nil {
			for i := range DataRecieve {
				FileName := string(DataRecieve[i])
				fileNames = append(fileNames, FileName)
			}
		} else {
			break
		}
	}
	return fileNames
}
func get_MD5(ctx context.Context, topic2 *pubsub.Topic, hostID peer.ID) []string {
	subscriber, err := topic2.Subscribe()
	if err != nil {
		panic(err)
	}
	var SlcStrMD5 []string
	for {
		SlcMD5 := subscribe(subscriber, ctx, hostID)
		var arrMD5 [16]byte
		copy(arrMD5[:], SlcMD5[:16])
		var arr0 [16]byte
		if arrMD5 != arr0 {
			strMD5 := string(arrMD5[:])
			SlcStrMD5 = append(SlcStrMD5, strMD5)
		} else {
			break
		}
	}
	return SlcStrMD5
}

func compare_Map(FileName []string, MD5 []string, GetFileName []string, GetMD5 []string, m2 map[string]string) {
	for i := range FileName {
		for j := range GetFileName {
			if FileName[i] == FileName[j] && MD5[i] != GetMD5[j] {
				fmt.Println("Syncing Fail")
			} else if FileName[i] == FileName[j] && MD5[i] == GetMD5[j] {
				delete(m2, MD5[i])
			}
		}
	}
}
