package proxy

import (
	"github.com/tjfoc/gmsm/gmtls"
	"security-gateway/internal/model"
)

type certificateManager struct {
	certificates map[uint16]map[string]*serviceCertificate
}

// serviceCertificate 服务对应的tls证书，包含证书和私钥
type serviceCertificate struct {
	port             uint16
	domain           string
	RsaCertificate   *gmtls.Certificate
	SmSigCertificate *gmtls.Certificate
	SmEncCertificate *gmtls.Certificate
}

func (m *certificateManager) UpdateServiceCertificate(port uint16, domain string, cert *model.Certificate) (err error) {
	var rsaCert gmtls.Certificate
	if cert.CertPem != "" && cert.KeyPem != "" {
		rsaCert, err = gmtls.X509KeyPair([]byte(cert.CertPem), []byte(cert.KeyPem))
		if err != nil {
			return
		}
	}
	var smSigCert gmtls.Certificate
	if cert.SignKeyPem != "" && cert.SignCertPem != "" {
		smSigCert, err = gmtls.X509KeyPair([]byte(cert.SignCertPem), []byte(cert.SignKeyPem))
		if err != nil {
			return
		}
	}
	var smEncCert gmtls.Certificate
	if cert.EncKeyPem != "" && cert.EncCertPem != "" {
		smEncCert, err = gmtls.X509KeyPair([]byte(cert.EncCertPem), []byte(cert.EncKeyPem))
		if err != nil {
			return
		}
	}

	if _, ok := m.certificates[port]; !ok {
		m.certificates[port] = make(map[string]*serviceCertificate)
	}

	m.certificates[port][domain] = &serviceCertificate{
		port:             port,
		domain:           domain,
		RsaCertificate:   &rsaCert,
		SmSigCertificate: &smSigCert,
		SmEncCertificate: &smEncCert,
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

func (m *certificateManager) generateDynamicTLSConfig(port uint16) (config *gmtls.Config) {
	gmSupport := gmtls.NewGMSupport()
	gmSupport.EnableMixMode()

	config = &gmtls.Config{
		GMSupport: gmSupport,
		GetCertificate: func(info *gmtls.ClientHelloInfo) (*gmtls.Certificate, error) {
			gmFlag := false
			// 检查支持协议中是否包含GMSSL
			for _, v := range info.SupportedVersions {
				if v == gmtls.VersionGMSSL {
					gmFlag = true
					break
				}
			}

			if certificate, ok := m.getServiceCertificate(port, info.ServerName); ok {
				if gmFlag && certificate.SmSigCertificate != nil {
					return certificate.SmSigCertificate, nil
				} else {
					return certificate.RsaCertificate, nil
				}
			}
			return nil, nil
		},
		GetKECertificate: func(info *gmtls.ClientHelloInfo) (*gmtls.Certificate, error) {
			if certificate, ok := m.getServiceCertificate(port, info.ServerName); ok {
				return certificate.SmEncCertificate, nil
			}
			return nil, nil
		},
	}
	return
}
