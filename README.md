# LIBP2P


### Libp2p là gì ?

libp2p là tập hợp các module về protocols, specifications và libraries cho phép phát triển các ứng dụng peer-to-peer.
### Những vấn đề libp2p giải quyết

##### Transport

Thực ra nền tảng của libP2P nhắm vào transport layer, chịu trách nhiệm truyền vào nhận dữ liệu giữa các peer trong mạng p2p. Libp2p cung cấp một interface đơn giản có thể điều chỉnh được để hỗ trợ các giao thức ở thiện tại và trong tương lai , cho phép các ứng dụng libp2p hoạt động trên nhiều môi trường bất kể thời gian .
##### Identity

Libp2p sử dụng public key cryptography làm cơ sở nhận dạng trong mạng ngang hàng phục vụ với hai mục đích.

Đầu tiên, nó cung cấp cho mỗi peer một tên gọi duy nhất dưới dạng PeerId .
Thứ hai là cho phép mọi người trong mạng giao tiếp an toàn với nhau . ( người A gửi một tin nhắn mã hóa bằng public key của người B vì thế chỉ người B mới có private key của mình để giải mã và đọc được tin nhắn đó). 
##### Security

Libp2p cung cấp một kết nối nâng cao bằng cách truyền dữ liệu qua encrypted channel . Quá trình này linh hoạt và có thể hỗ trợ nhiều phương thức mã hóa truyền thống .Mặc định hiện tại là secio.
##### Peer Routing

Peer routing là quá trình khám phá các địa chỉ bằng cách vận dụng "kiến thức" của các peer khác.

Trong một hệ thống peer routing một peer có thể cung cấp cho chúng ta địa chỉ của peer chúng ta cần miễn là chúng có . Nếu không nó sẽ chuyển yêu cầu của chúng ta đến một peer khác mà nhiều khả năng sẽ có . Khi chúng ta giao tiếp với càng nhiều peer chúng ta không chỉ tăng cơ hội tìm thấy peer mà chúng ta đang tìm kiếm mà còn cập nhật được đầy đủ hơn về mạng , cho phép trả lời các truy vấn định tuyến của từ các peer khác . Ngoài ra libp2p sử dụng Kademlia routing algorithm.

##### Content Discovery

Libp2p cung cấp content routing interface để có thể giúp chúng ta tìm ,tải và verify tính toàn diện của dữ liệu mà chúng ta cần tìm.
##### Messaging/PubSub
Việc gửi các message từ peer này tới peer khác luôn là vấn đề quan trọng nhất đối với các hệ thống peer-to-peer và pub/sub là một pattern vô cùng quan trọng . Định nghĩa qua về pub/sub là một hệ thống mà trong đó các peer sẽ subscribe vào các chủ đề ( topic ) mà họ quan tâm đến và các peer gửi dữ liệu vào topic đó gọi là public

Trong một hệ thống peer-to-peer pub/sub có đôi chút khác biệt với các hệ thống pub/sub thông thường .Bao gồm:

* Reliability (độ tin cậy ) : Tất cả các message phải được gửi đến tất cả các peer đã đăng ký chủ đề.
* Speed ( tốc độ ) : Message phải được gửi một cách nhanh chóng
* Efficiency ( hiệu quả ) : Mạng sẽ ko tồn tại quá nhiều những bản copy của messages một cách thừa thãi
* Resilience ( khả năng phục hồi ) : Các peer có thể tham gia hoặc thoát mạng mà không làm gián đoạn nó.
* Scale ( khả năng mở rộng ) : Các topic có thể có xử lý một lượng lớn người dùng đăng ký và xử lý một lượng lớn message
* Simplicity ( sự đơn giản ) : Hệ thống này phải đủ đơn giản để có thể hiểu và thực hiện . Các peer cũng chỉ cần phải lưu ít giữ liệu.
  
Libp2p định nghĩa một pubsub interface để có thể gửi message đến tất cả các peer đã subscribe vào cùng một topic . Interface hiện tại có 2 triển khai .

* floodsub sử dụng chiến lược network flooding là truyền lần lượt đến tất cả các peer, tuy nhiên giải pháp này không hiệu quả .
* gossipsub là giao thức hiện tại đang được libp2p sửa dụng . Đúng như cái tên gossip là đồn thổi, các peer đồn với nhau về message mà chúng thấy được với các peer khác.
* episub: một  bản gossipsub mở rộng được tối ưu hóa cho single source multicast and scenarios với một số sửa đổi hướng tới 1 lượng lớn client trong 1 topic.

# Code node Libp2p 

##### Tạo 1 node với các cài đặt cơ bản.

>   package main
>
>   import (
> 
>	    "fmt"
>
>	    "github.com/libp2p/go-libp2p"
>   )
>
>   func main() {
> 
>	    // start a libp2p node with default settings
>	    node, err := libp2p.New()
>	    if err != nil {
>	    panic(err)
>	    }
>
>	    // print the node's listening addresses
>	    fmt.Println("Listen addresses:", node.Addrs())
>
>	    // shut the node down
>	    if err := node.Close(); err != nil {
>		    panic(err)
>	    }
>   }

##### Tạo 1 node với các cài đặt khác
###### Tạo 1 keypair 
>priv, _, err := crypto.GenerateKeyPair(
>
>		crypto.Ed25519, // Chọn kiểu mã khóa. VD: Ed25519
>
>		-1,             // Chọn độ dài của key. VD: RSA .
>	)
>	if err != nil {
>
>		panic(err)
>	}
###### Cau hinh 1 node
>h2, err := libp2p.New(
>
>		// Sử dụng keypair vừa tạo 
>		libp2p.Identity(priv),
>
>		// Multiple listen addresses
>		libp2p.ListenAddrStrings(
>			"/ip4/0.0.0.0/tcp/9000",      // kết nối tcp thông thường
>			"/ip4/0.0.0.0/udp/9000/quic", // UDP endpoint cho QUIC transport
>		),
>
>		// support TLS connections
>		libp2p.Security(libp2ptls.ID, libp2ptls.New),
>
>		// support noise connections
>		libp2p.Security(noise.ID, noise.New),
>
>		// support any other default transports (TCP)
>		libp2p.DefaultTransports,
>
>		// Giới hạn số kết nói tới node 
>       // Khi số kết nối vượt quá high, một số kết nối sẽ bị buộc chấm dứt
>          cho tới khi đạt mức low
>
>		libp2p.ConnectionManager(connmgr.NewConnManager(
>			100,         // Lowwater
>			400,         // HighWater,
>			time.Minute, // GracePeriod
>		)),
>
>		// mở các cổng bằng uPNP cho các máy chủ NAT.
>		libp2p.NATPortMap(),
>		// Sử dụng DHT để tìm các node khác 
>		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
>			idht, err = dht.New(ctx, h)
>			return idht, err
>		}),
>
>		// host use relays and advertise itself on relays if
>		// it finds it is behind NAT.
>		libp2p.EnableAutoRelay(),
>		// launch the server-side of AutoNAT to help other peers to figure 
>       //out if they are behind NATs.
>		// This service is highly rate-limited and should not cause any
>		// performance issues.
>		libp2p.EnableNATService(),
>	)

##### Connect 2 nodes
>   package main

>   import (
> 	    "context"
> 	    "fmt"
> 	    "os"
> 	    "os/signal"
> 	    "syscall"
> 
> 	    "github.com/libp2p/go-libp2p"
> 	    peerstore "github.com/libp2p/go-libp2p-core/peer"
> 	    "github.com/libp2p/go-libp2p/p2p/protocol/ping"
> 	    multiaddr "github.com/multiformats/go-multiaddr"
>   )
> 
>   func main() {
>	    // start a libp2p node that listens on a random local TCP port,
> 	    // but without running the built-in ping protocol
> 	    node, err := libp2p.New(
> 		    libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
> 		    libp2p.Ping(false),
> 	    )
> 	    if err != nil {
> 		    panic(err)
> 	    }
> 
> 	    // configure our own ping protocol
> 	    pingService := &ping.PingService{Host: node}
> 	    node.SetStreamHandler(ping.ID, pingService.PingHandler)
> 
> 	    // print the node's PeerInfo in multiaddr format
> 	    peerInfo := peerstore.AddrInfo{
>		    ID:    node.ID(),
> 		    Addrs: node.Addrs(),
> 	    }
> 	    addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
>	    if err != nil {
>		    panic(err)
>	    }
>	    fmt.Println("libp2p node address:", addrs[0])
>
>	    // if a remote peer has been passed on the command line, connect to it
>	    // and send it 5 ping messages, otherwise wait for a signal to stop
>	    if len(os.Args) > 1 {
>		    addr, err := multiaddr.NewMultiaddr(os.Args[1])
>		    if err != nil {
>			    panic(err)
>           }
>		    peer, err := peerstore.AddrInfoFromP2pAddr(addr)
>		    if err != nil {
>			    panic(err)
>		    }
>		    if err := node.Connect(context.Background(), *peer); err != nil {
>			    panic(err)
>		    }
>		    fmt.Println("sending 5 ping messages to", addr)
>		    ch := pingService.Ping(context.Background(), peer.ID)
>		    for i := 0; i < 5; i++ {
>			    res := <-ch
>			    fmt.Println("pinged", addr, "in", res.RTT)
>		    }
>	    } else {
>		    // wait for a SIGINT or SIGTERM signal
>		    ch := make(chan os.Signal, 1)
>		    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
>		    <-ch
>		    fmt.Println("Received signal, shutting down...")
>	    }
>
>	    // shut the node down
>	    if err := node.Close(); err != nil {
>		    panic(err)
>	    }
>   }
-----------------------------------
