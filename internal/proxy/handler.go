package proxy

import "crypto/tls"

var CertificateManager = &certificateManager{
	certificates: make(map[uint16]map[string]*serviceCertificate),
}

type certificateManager struct {
	certificates map[uint16]map[string]*serviceCertificate
}

// serviceCertificate 服务对应的tls证书，包含证书和私钥
type serviceCertificate struct {
	port        uint16
	domain      string
	cert        []byte
	key         []byte
	Certificate tls.Certificate
}

func (m *certificateManager) updateServiceCertificate(port uint16, domain string, cert []byte, key []byte) (err error) {
	var certificate tls.Certificate
	certificate, err = tls.X509KeyPair(cert, key)
	if err != nil {
		return
	}

	if _, ok := m.certificates[port]; !ok {
		m.certificates[port] = make(map[string]*serviceCertificate)
	}

	m.certificates[port][domain] = &serviceCertificate{
		port:        port,
		domain:      domain,
		cert:        cert,
		key:         key,
		Certificate: certificate,
	}

	return
}

func (m *certificateManager) getServiceCertificate(port uint16, domain string) (certificate *serviceCertificate, ok bool) {
	if _, ok = m.certificates[port]; !ok {
		return
	}

	certificate, ok = m.certificates[port][domain]
	return
}

func (m *certificateManager) deleteServiceCertificate(port uint16, domain string) {
	if _, ok := m.certificates[port]; !ok {
		return
	}

	delete(m.certificates[port], domain)
}

func (m *certificateManager) generateDynamicTLSConfig(port uint16) (config *tls.Config) {
	config = &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			if certificate, ok := m.getServiceCertificate(port, info.ServerName); ok {
				return &certificate.Certificate, nil
			}

			return nil, nil
		},
	}
	return
}

//
//
//func handler() {
//	tlsConfig := &tls.Config{
//		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
//			switch info.ServerName {
//			case "example.com":
//				tls.X509KeyPair(certPEMBlock, keyPEMBlock)
//			}
//
//			return nil, nil
//		},
//	}
//}
