package proxy

import (
	logger "github.com/sirupsen/logrus"
	"github.com/tjfoc/gmsm/gmtls"
	"net"
)

type AppListener struct {
	net.Listener
	tlsConfig *gmtls.Config
	//tlsConfig *tls.Config
}

//func NewAppListener(listener net.Listener, tlsConfig *tls.Config) *AppListener {
//	return &AppListener{
//		Listener:  listener,
//		tlsConfig: tlsConfig,
//	}
//}

func NewAppListener(listener net.Listener, tlsConfig *gmtls.Config) *AppListener {
	return &AppListener{
		Listener:  listener,
		tlsConfig: tlsConfig,
	}
}

func (l *AppListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// 读取第一个字节，判断是否是TLS连接
	buffer := make([]byte, 1)
	var readNum int
	readNum, err = conn.Read(buffer)
	if err != nil {
		logger.WithField("readNum", readNum).Error(err)
		// 如果readNum为0，说明连接已经关闭
		if readNum == 0 {
			return conn, nil
		}
		// 如果readNum不为0，说明连接未关闭，但是读取数据失败
		return nil, err
	}

	conn = newBufferedConn(conn, buffer)

	if buffer[0] == 0x16 {
		// 是TLS连接
		//tlsConn := tls.Server(conn, l.tlsConfig)
		//if err = tlsConn.Handshake(); err != nil {
		//	return nil, err
		//}

		//sni := tlsConn.ConnectionState().ServerName
		//// sni即域名，根据域名找到对应的路由，并完成服务

		//return tlsConn, nil
		//return tls.Server(conn, l.tlsConfig), nil
		return gmtls.Server(conn, l.tlsConfig), nil
	}

	return conn, nil
}

// 自定义一个缓冲连接
type bufferedConn struct {
	net.Conn
	buf []byte
}

func newBufferedConn(conn net.Conn, buf []byte) net.Conn {
	return &bufferedConn{
		Conn: conn,
		buf:  buf,
	}
}

//	func (bc *bufferedConn) Read(p []byte) (int, error) {
//		if len(bc.buf) > 0 {
//			n := copy(p, bc.buf)
//			bc.buf = bc.buf[n:]
//			return n, nil
//		}
//		n, err := bc.Conn.Read(p)
//		if err != nil {
//			logger.Error(err)
//		}
//		return n, err
//	}
func (bc *bufferedConn) Read(p []byte) (int, error) {
	totalN := 0
	if len(bc.buf) > 0 {
		n := copy(p, bc.buf)
		bc.buf = bc.buf[n:]
		totalN += n
	}
	if totalN < len(p) {
		n, err := bc.Conn.Read(p[totalN:])
		if err != nil {
			logger.Error(err)
			return totalN, err
		}
		totalN += n
	}
	return totalN, nil
}
