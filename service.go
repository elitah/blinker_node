package blinker

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type blinkerPayloadPowerSet struct {
	Value     bool  `json:"value"`
	ConfirmID int64 `json:"confirm_id"`
}

type blinkerMsg struct {
	Key     string          `json:"key"`
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

type blinkerNode struct {
	sync.Mutex

	flag uint32

	exit chan struct{}

	timeout time.Duration

	loglevel int
	logger   *log.Logger

	network, address string

	conn net.Conn

	cb_resolve_fail func(string, string)
	cb_power_set    func(bool)
	cb_update       func() bool
}

func NewBlinkerNode(options ...option) *blinkerNode {
	//
	node := &blinkerNode{
		exit: make(chan struct{}),

		timeout: time.Minute,
	}
	//
	for _, item := range options {
		//
		if nil != item {
			//
			item(node)
		}
	}
	//
	if 5*time.Second > node.timeout {
		//
		node.timeout = time.Minute
	} else if 3*time.Minute < node.timeout {
		//
		node.timeout = 3 * time.Minute
	}
	//
	return node
}

func (this *blinkerNode) Close() {
	//
	if 0x0 != atomic.LoadUint32(&this.flag) {
		//
		select {
		case <-this.exit:
		default:
			//
			close(this.exit)
		}
		//
		this.Lock()
		//
		if nil != this.conn {
			//
			this.conn.Close()
			this.conn = nil
		}
		//
		this.Unlock()
	}
}

func (this *blinkerNode) IsRunning() bool {
	//
	return 0x0 != atomic.LoadUint32(&this.flag)
}

func (this *blinkerNode) IsConnected() bool {
	//
	return 0x2 == atomic.LoadUint32(&this.flag)
}

func (this *blinkerNode) Loop(key string) error {
	//
	if "" != key {
		//
		if ("" != this.network && "" != this.address) || nil != this.conn {
			//
			if atomic.CompareAndSwapUint32(&this.flag, 0x0, 0x1) {
				//
				if nil != this.conn {
					//
					go func() {
						//
						if err := this.handleConn(this.conn, key); nil != err {
							//
							this.logwarn(err)
						}
						//
						atomic.StoreUint32(&this.flag, 0x0)
					}()
					//
					return nil
				} else {
					//
					go func(network, address string) {
						//
						var failed_cnt int
						//
						defer atomic.StoreUint32(&this.flag, 0x0)
						//
						for {
							//
							switch network {
							case "tcp", "tcp4", "tcp6":
								//
								if addr, err := net.ResolveTCPAddr(network, address); nil == err {
									//
									if conn, err := net.DialTCP(network, nil, addr); nil == err {
										//
										failed_cnt = 0
										//
										this.Lock()
										//
										this.conn = conn
										//
										this.Unlock()
										//
										if err := this.handleConn(conn, key); nil != err {
											//
											this.logwarn(err)
										}
										//
										this.Lock()
										//
										this.conn = nil
										//
										this.Unlock()
									} else {
										//
										this.logwarn(err)
									}
								} else {
									//
									this.logwarn(err)
									//
									if nil != this.cb_resolve_fail {
										//
										this.cb_resolve_fail(network, address)
									}
								}
							case "udp", "udp4", "udp6":
								//
								if addr, err := net.ResolveUDPAddr(network, address); nil == err {
									//
									if conn, err := net.DialUDP(network, nil, addr); nil == err {
										//
										failed_cnt = 0
										//
										this.Lock()
										//
										this.conn = conn
										//
										this.Unlock()
										//
										if err := this.handleConn(conn, key); nil != err {
											//
											this.logwarn(err)
										}
										//
										this.Lock()
										//
										this.conn = nil
										//
										this.Unlock()
									} else {
										//
										this.logwarn(err)
									}
								} else {
									//
									this.logwarn(err)
									//
									if nil != this.cb_resolve_fail {
										//
										this.cb_resolve_fail(network, address)
									}
								}
							}
							//
							select {
							case <-this.exit:
								//
								return
							default:
							}
							//
							if 3 > failed_cnt {
								//
								time.Sleep(time.Second)
							} else if 30 > failed_cnt {
								//
								time.Sleep(time.Duration(failed_cnt/3+1) * time.Second)
							} else {
								//
								time.Sleep(10 * time.Second)
								//
								if 40 <= failed_cnt {
									//
									failed_cnt = 0
									//
									continue
								}
							}
							//
							failed_cnt++
						}
					}(this.network, this.address)
					//
					return nil
				}
			} else {
				//
				return syscall.EBUSY
			}
		}
	}
	//
	return syscall.EINVAL
}

func (this *blinkerNode) logout(level int, v interface{}, args ...interface{}) {
	//
	if (this.loglevel >= level) && nil != this.logger {
		//
		prefix := ""
		//
		switch level {
		case LogLevelDebug:
			//
			prefix = "\033[33;1m[D]\033[0m "
		case LogLevelWarn:
			//
			prefix = "\033[31;1m[W]\033[0m "
		case LogLevelErr:
			//
			prefix = "\033[30;41m[E]\033[0m "
		case LogLevelFatal:
			//
			prefix = "\033[30;43m[F]\033[0m "
		default:
			//
			prefix = "[I] "
		}
		//
		if format, ok := v.(string); ok {
			//
			this.logger.Output(3, fmt.Sprintf(prefix+format, args...))
			//
			return
		}
		//
		this.logger.Output(3, fmt.Sprint(prefix, v)+fmt.Sprint(args...))
	}
}

func (this *blinkerNode) loginfo(v interface{}, args ...interface{}) {
	//
	this.logout(LogLevelInfo, v, args...)
}

func (this *blinkerNode) logdebug(v interface{}, args ...interface{}) {
	//
	this.logout(LogLevelDebug, v, args...)
}

func (this *blinkerNode) logwarn(v interface{}, args ...interface{}) {
	//
	this.logout(LogLevelWarn, v, args...)
}

func (this *blinkerNode) logerr(v interface{}, args ...interface{}) {
	//
	this.logout(LogLevelErr, v, args...)
}

func (this *blinkerNode) logfatal(v interface{}, args ...interface{}) {
	//
	this.logout(LogLevelFatal, v, args...)
}

func (this *blinkerNode) handleConn(conn net.Conn, key string) error {
	//
	var buffer [1024]byte
	//
	var heartbeat bytes.Buffer
	//
	defer conn.Close()
	//
	atomic.StoreUint32(&this.flag, 0x2)
	//
	defer func() {
		//
		atomic.StoreUint32(&this.flag, 0x1)
		//
		conn.Close()
	}()
	//
	if data, err := json.Marshal(blinkerMsg{
		Key:    key,
		Action: "keepalive",
	}); nil == err {
		//
		heartbeat.WriteString(base64.StdEncoding.EncodeToString(data))
	} else {
		//
		return err
	}
	//
	lastLoopTime := time.Time{}
	lastRecvTime := time.Now()
	lastHBTime := time.Time{}
	lastUpdateTime := time.Time{}
	//
	for {
		//
		lastLoopTime = time.Now()
		//
		if 5*time.Minute < lastLoopTime.Sub(lastRecvTime) {
			//
			return syscall.ETIMEDOUT
		}
		//
		if d := lastLoopTime.Sub(lastHBTime); this.timeout-5*time.Second <= d {
			//
			if _, err := conn.Write(heartbeat.Bytes()); nil != err {
				//
				this.logwarn(err)
			}
			//
			lastHBTime = lastLoopTime
			//
			conn.SetReadDeadline(lastLoopTime.Add(this.timeout))
		} else {
			//
			conn.SetReadDeadline(time.Now().Add(this.timeout - d))
		}
		//
		if 2*time.Minute <= lastLoopTime.Sub(lastUpdateTime) {
			//
			if err := this.sendUpdate(conn, key, 0); nil == err {
				//
				lastUpdateTime = lastLoopTime
			} else {
				//
				this.logwarn(err)
			}
		}
		//
		if n, err := conn.Read(buffer[:]); nil == err {
			//
			this.loginfo("recv: %s", string(buffer[:n]))
			//
			lastRecvTime = lastLoopTime
			//
			if 2 == n && "ok" == string(buffer[:2]) {
			} else if data, err := base64.StdEncoding.DecodeString(
				string(buffer[:n]),
			); nil == err {
				//
				var result blinkerMsg
				//
				if err := json.Unmarshal(data, &result); nil == err {
					//
					if key == result.Key {
						//
						switch result.Action {
						case "powerset":
							//
							var payload blinkerPayloadPowerSet
							//
							if err := json.Unmarshal(result.Payload, &payload); nil == err {
								//
								if nil != this.cb_power_set {
									//
									this.cb_power_set(payload.Value)
								}
								//
								if err := this.sendUpdate(conn, key, payload.ConfirmID); nil == err {
									//
									lastUpdateTime = lastLoopTime
								} else {
									//
									this.logwarn(err)
								}
							} else {
								//
								this.logwarn(err)
							}
						default:
							//
							this.logwarn(err)
						}
					} else {
						//
						this.logwarn(err)
					}
				} else {
					//
					this.logwarn(err)
				}
			} else {
				//
				this.logwarn(err)
			}
		} else if _err, ok := err.(net.Error); ok && _err.Timeout() {
		} else {
			//
			return err
		}
	}
}

func (this *blinkerNode) sendUpdate(conn net.Conn, key string, cid int64) error {
	//
	var value bool
	//
	if nil != this.cb_update {
		//
		value = this.cb_update()
	}
	//
	if _, err := conn.Write([]byte(base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(
		`{"key":"%s","action":"update","payload":{"status":%v,"confirm_id":%d}}`,
		key,
		value,
		cid,
	))))); nil != err {
		//
		return err
	}
	//
	return nil
}
