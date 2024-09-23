package main

import (
	"net"
	"os"
	"syscall"
)

func socketReuseAddr(sock int) error {
	return syscall.SetsockoptInt(
		sock,
		syscall.SOL_SOCKET,
		syscall.SO_REUSEADDR,
		1,
	)
}

func socketBindToDevice(sock int, ifi *net.Interface) error {
	return syscall.SetsockoptString(
		sock,
		syscall.SOL_SOCKET,
		syscall.SO_BINDTODEVICE,
		ifi.Name,
	)
}

func socketMulticastIf4(sock int, mreq *syscall.IPMreqn) error {
	return syscall.SetsockoptIPMreqn(
		sock,
		syscall.IPPROTO_IP,
		syscall.IP_MULTICAST_IF,
		mreq,
	)
}

func socketMulticastJoin4(sock int, ifi *net.Interface, ip net.IP) error {
	mreq := syscall.IPMreqn{}

	if ifi != nil {
		mreq.Ifindex = int32(ifi.Index)
	}
	copy(mreq.Multiaddr[:], ip.To4())

	if err := socketMulticastIf4(sock, &mreq); err != nil {
		return err
	}

	return syscall.SetsockoptIPMreqn(
		sock,
		syscall.IPPROTO_IP,
		syscall.IP_ADD_MEMBERSHIP,
		&mreq,
	)
}

func openSocket4(ifi *net.Interface, ip net.IP, port int) (net.PacketConn, error) {
	sock, err := syscall.Socket(
		syscall.AF_INET,
		syscall.SOCK_DGRAM,
		syscall.IPPROTO_UDP,
	)
	if err != nil {
		return nil, err
	}

	if err := socketReuseAddr(sock); err != nil {
		syscall.Close(sock)
		return nil, err
	}

	if ifi != nil {
		if err := socketBindToDevice(sock, ifi); err != nil {
			syscall.Close(sock)
			return nil, err
		}
	}

	addr := syscall.SockaddrInet4{}
	addr.Port = port
	copy(addr.Addr[:], ip.To4())

	if err := syscall.Bind(sock, &addr); err != nil {
		syscall.Close(sock)
		return nil, err
	}

	if ip.IsMulticast() {
		if err := socketMulticastJoin4(sock, ifi, ip); err != nil {
			syscall.Close(sock)
			return nil, err
		}
	}

	file := os.NewFile(uintptr(sock), "")
	conn, err := net.FilePacketConn(file)
	file.Close()
	if err != nil {
		return nil, err
	}

	return conn, nil
}
