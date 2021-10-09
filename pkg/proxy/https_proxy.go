package proxy

import (
	"bufio"
	"bytes"
	crypto_rand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"github.com/proxy-server/internal/pkg/models"
	"github.com/proxy-server/pkg/request_utils"
	"io"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func (p *ProxyManager) ProxyHttpsRequest(w http.ResponseWriter, r *http.Request, save bool) error {
	requestUrl, _ := url.Parse(r.RequestURI)
	var scheme string
	if requestUrl.Scheme == "" {
		scheme = r.URL.Host
	} else {
		scheme = requestUrl.Scheme
	}

	rootDir, _ := os.Getwd()
	certsPath := rootDir + "/certificates"
	if _, errStat := os.Stat(certsPath); os.IsNotExist(errStat) {
		if err := os.MkdirAll(certsPath, 0700); err != nil {
			return err
		}
	}

	hijackedConn, err := p.interceptConnection(w)
	if err != nil {
		return err
	}
	defer hijackedConn.Close()

	tcpClientConn, err := p.initTcpClient(hijackedConn, scheme)
	if err != nil {
		return err
	}
	defer tcpClientConn.Close()

	clientReader := bufio.NewReader(tcpClientConn)
	tcpClientRequest, err := http.ReadRequest(clientReader)
	if err != nil {
		return err
	}

	var request *models.Request
	if save {
		request, err = request_utils.ParseRequest(tcpClientRequest, "https")
		if err != nil {
			return err
		}

		err = p.proxyRepository.InsertRequest(request)
		if err != nil {
			return err
		}
	}

	tcpServerConn, err := p.initTcpServer(r.Host, scheme)
	if err != nil {
		return err
	}
	defer tcpServerConn.Close()

	if err = p.makeHttpsRequest(tcpClientConn, tcpServerConn, tcpClientRequest, request.Id, save); err != nil {
		return err
	}

	return nil
}

func (p *ProxyManager) interceptConnection(w http.ResponseWriter) (net.Conn, error) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("creating hijacker failed")
	}

	hijackedConn, _, err := hijacker.Hijack()
	if err != nil {
		return nil, err
	}

	_, err = hijackedConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	if err != nil {
		hijackedConn.Close()
		return nil, err
	}

	return hijackedConn, nil
}

func (p *ProxyManager) initTcpClient(hijackedConn net.Conn, scheme string) (*tls.Conn, error) {
	cert, err := p.generateCertificate(scheme)
	if err != nil {
		return nil, err
	}

	tcpClientConn := tls.Server(hijackedConn, &tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   scheme,
	})

	if err = tcpClientConn.Handshake(); err != nil {
		tcpClientConn.Close()
		return nil, err
	}

	return tcpClientConn, nil
}

func (p *ProxyManager) initTcpServer(host, scheme string) (*tls.Conn, error) {
	file, err := os.Open("./keys/ca.key")
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	privPem, _ := pem.Decode(b)

	priv, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{scheme},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(crypto_rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})
	cert, err := tls.X509KeyPair(caPEM.Bytes(), b)
	conf := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	tcpServerConn, err := tls.Dial("tcp", host, conf)
	if err != nil {
		return nil, err
	}

	return tcpServerConn, nil
}

func (p *ProxyManager) generateCertificate(scheme string) (tls.Certificate, error) {
	rootDir, err := os.Getwd()
	if err != nil {
		return tls.Certificate{}, err
	}

	cmdGenDir := rootDir + "/keys"
	certsDir := rootDir + "/certificates/"

	certFilename := certsDir + scheme + ".crt"

	if _, errStat := os.Stat(certFilename); os.IsNotExist(errStat) {
		genCommand := exec.Command("bash", rootDir+"/scripts/bash/generate_cert.sh",
			scheme, strconv.Itoa(rand.Intn(100000000)))

		if _, err = genCommand.CombinedOutput(); err != nil {
			return tls.Certificate{}, err
		}
	}

	cert, err := tls.LoadX509KeyPair(certFilename, cmdGenDir+"/cert.key")
	if err != nil {
		return tls.Certificate{}, err
	}

	return cert, nil
}

func (p *ProxyManager) makeHttpsRequest(tcpClientConn, tcpServerConn *tls.Conn, request *http.Request,
	requestId int64, save bool) error {
	if err := request.Write(tcpServerConn); err != nil {
		return err
	}

	serverReader := bufio.NewReader(tcpServerConn)
	tcpServerResponse, err := http.ReadResponse(serverReader, request)
	if err != nil {
		return err
	}

	rawResp, err := httputil.DumpResponse(tcpServerResponse, true)
	if _, err = tcpClientConn.Write(rawResp); err != nil {
		return err
	}

	var response *models.Response
	if save {
		response, err = request_utils.ParseResponse(tcpServerResponse, requestId)
		if err != nil {
			return err
		}

		err = p.proxyRepository.InsertResponse(response)
		if err != nil {
			return err
		}
	}

	go func(destination, source *tls.Conn) {
		defer destination.Close()
		defer source.Close()
		io.Copy(destination, source)
	}(tcpClientConn, tcpServerConn)

	go func(destination, source *tls.Conn) {
		defer destination.Close()
		defer source.Close()
		io.Copy(destination, source)
	}(tcpServerConn, tcpClientConn)

	return nil
}
