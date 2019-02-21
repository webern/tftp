// Copyright (c) 2019 by Matthew James Briggs, https://github.com/webern

package tftpsrv

import (
	"github.com/webern/flog"
	"github.com/webern/tftp/tftplib/wire"
	"net"
)

func transfer(hndshk handshake) error {
	conn, err := net.DialUDP("udp", &hndshk.server, &hndshk.client)

	if err != nil {
		return flog.Wrap(err)
	}

	if err := sendHandshakeAck(conn); err != nil {
		return flog.Wrap(err)
	}

	return nil
}

func sendHandshakeAck(conn *net.UDPConn) error {
	ack := wire.PacketAck{}
	_, err := conn.Write(ack.Serialize())

	if err != nil {
		return err
	}

	return nil
}
