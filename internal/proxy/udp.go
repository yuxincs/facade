package proxy

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"
)

// UDP forwards datagrams to an already-running backend.
type UDP struct {
	BackendAddress string
	// IdleTimeout expires sessions with no traffic in either direction.
	// Zero uses one minute.
	IdleTimeout time.Duration
}

// Run forwards traffic until cancellation or a listener error.
// It owns listener and closes it and all backend connections before returning.
func (p *UDP) Run(ctx context.Context, listener *net.UDPConn) error {
	defer listener.Close()
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.IdleTimeout < 0 {
		return fmt.Errorf("UDP idle timeout must not be negative")
	}
	idleTimeout := p.IdleTimeout
	if idleTimeout == 0 {
		idleTimeout = time.Minute
	}
	backend, err := net.ResolveUDPAddr("udp", p.BackendAddress)
	if err != nil {
		return fmt.Errorf("resolve UDP backend: %w", err)
	}
	if backend.Port == 0 {
		return fmt.Errorf("UDP backend port must not be zero")
	}

	var mu sync.Mutex
	sessions := make(map[netip.AddrPort]*net.UDPConn)
	var workers sync.WaitGroup
	stop := context.AfterFunc(ctx, func() { listener.Close() })
	defer func() {
		stop()
		listener.Close()
		mu.Lock()
		for _, conn := range sessions {
			conn.Close()
		}
		mu.Unlock()
		workers.Wait()
	}()

	buffer := make([]byte, 65535)
	for {
		n, client, err := listener.ReadFromUDPAddrPort(buffer)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("read UDP client: %w", err)
		}
		mu.Lock()
		conn := sessions[client]
		if conn == nil {
			conn, err = net.DialUDP("udp", nil, backend)
			if err != nil {
				mu.Unlock()
				return fmt.Errorf("connect UDP backend: %w", err)
			}
			sessions[client] = conn
			workers.Add(1)
			go func(conn *net.UDPConn, client netip.AddrPort) {
				defer workers.Done()
				defer func() {
					mu.Lock()
					if sessions[client] == conn {
						delete(sessions, client)
					}
					conn.Close()
					mu.Unlock()
				}()
				response := make([]byte, 65535)
				for {
					n, err := conn.Read(response)
					if err != nil {
						return
					}
					conn.SetReadDeadline(time.Now().Add(idleTimeout))
					if _, err := listener.WriteToUDPAddrPort(response[:n], client); err != nil {
						return
					}
				}
			}(conn, client)
		}
		conn.SetDeadline(time.Now().Add(idleTimeout))
		mu.Unlock()
		_, err = conn.Write(buffer[:n])
		if err != nil {
			mu.Lock()
			if sessions[client] == conn {
				delete(sessions, client)
			}
			conn.Close()
			mu.Unlock()
		}
	}
}
