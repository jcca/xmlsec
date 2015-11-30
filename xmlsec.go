/*
Package xmlsec is a Go binding for XML Security Library (https://www.aleksey.com/xmlsec/index.html)

In order to use this library, you need go-libxml2 (https://github.com/lestrrat/go-libxml2)
*/

package xmlsec

/*
#cgo pkg-config: xmlsec1
#include <stdlib.h>
#include <libxml/tree.h>
#include <libxml/xmlerror.h>
#include <libxml/xmlstring.h>
#include <libxslt/xslt.h>
#include <libxslt/security.h>
#include <xmlsec/xmlsec.h>
#include <xmlsec/xmltree.h>
#include <xmlsec/xmldsig.h>
#include <xmlsec/templates.h>
#include <xmlsec/transforms.h>
#include <xmlsec/crypto.h>

static inline xmlChar* to_xmlcharptr(const char *s) {
  return (xmlChar *) s;
}

static inline char * to_charptr(const xmlChar *s) {
  return (char *) s;
}

static xsltSecurityPrefsPtr xsltSecPrefs = NULL;
static int
go_xmlsec_init() {
  xmlIndentTreeOutput = 1;
  xsltSecPrefs = xsltNewSecurityPrefs();
  xsltSetSecurityPrefs(xsltSecPrefs,  XSLT_SECPREF_READ_FILE,        xsltSecurityForbid);
  xsltSetSecurityPrefs(xsltSecPrefs,  XSLT_SECPREF_WRITE_FILE,       xsltSecurityForbid);
  xsltSetSecurityPrefs(xsltSecPrefs,  XSLT_SECPREF_CREATE_DIRECTORY, xsltSecurityForbid);
  xsltSetSecurityPrefs(xsltSecPrefs,  XSLT_SECPREF_READ_NETWORK,     xsltSecurityForbid);
  xsltSetSecurityPrefs(xsltSecPrefs,  XSLT_SECPREF_WRITE_NETWORK,    xsltSecurityForbid);
  xsltSetDefaultSecurityPrefs(xsltSecPrefs);

  if(xmlSecInit() < 0) {
    return -1;
  }

  if(xmlSecCheckVersion() != 1) {
    return -1;
  }

  if(xmlSecCryptoAppInit(NULL) < 0) {
    return -1;
  }

  if(xmlSecCryptoInit() < 0) {
    fprintf(stderr, "Error: xmlsec-crypto initialization failed.\n");
    return -1;
  }

	return 0;
}

static void
go_xmlsec_shutdown() {
  xmlSecCryptoShutdown();
  xmlSecCryptoAppShutdown();
  xmlSecShutdown();
  xsltFreeSecurityPrefs(xsltSecPrefs);
  xsltCleanupGlobals();
}

*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/lestrrat/go-libxml2"
)

type KeyDataFormat int

const (
	KeyDataFormatUnknown  KeyDataFormat = C.xmlSecKeyDataFormatUnknown
	KeyDataFormatBinary                 = C.xmlSecKeyDataFormatBinary
	KeyDataFormatPem                    = C.xmlSecKeyDataFormatPem
	KeyDataFormatDer                    = C.xmlSecKeyDataFormatDer
	KeyDataFormatPkcs8Pem               = C.xmlSecKeyDataFormatPkcs8Pem
	KeyDataFormatPkcs8Der               = C.xmlSecKeyDataFormatPkcs8Der
	KeyDataFormatPkcs12                 = C.xmlSecKeyDataFormatPkcs12
	KeyDataFormatCertPem                = C.xmlSecKeyDataFormatCertPem
	KeyDataFormatCertDer                = C.xmlSecKeyDataFormatCertDer
)

func Init() error {
	if C.go_xmlsec_init() < C.int(0) {
		return errors.New("failed to initialize")
	}
	return nil
}

func Shutdown() error {
	C.go_xmlsec_shutdown()
	return nil
}

func xmlCharToString(s *C.xmlChar) string {
	return C.GoString(C.to_charptr(s))
}

func stringToXMLChar(s string) *C.xmlChar {
	return C.to_xmlcharptr(C.CString(s))
}

func validNodePtr(n libxml2.Node) (*C.xmlNode, error) {
	if n == nil {
		return nil, libxml2.ErrInvalidNode
	}

	nptr := (*C.xmlNode)(n.Pointer())
	if nptr == nil {
		return nil, libxml2.ErrInvalidNode
	}

	return nptr, nil
}

func xmlSecDSigCtxCreate() (*DSigCtx, error) {
	ctx := C.xmlSecDSigCtxCreate(nil)
	if ctx == nil {
		return nil, errors.New("failed to create DSigCtx")
	}
	return &DSigCtx{ptr: ctx}, nil
}

func validDSigCtxPtr(ctx *DSigCtx) (*C.xmlSecDSigCtx, error) {
	if ctx == nil {
		return nil, ErrInvalidDSigCtx
	}

	if ptr := ctx.ptr; ptr != nil {
		return ptr, nil
	}

	return nil, ErrInvalidDSigCtx
}

func xmlSecDSigCtxDestroy(ctx *DSigCtx) error {
	ctxptr, err := validDSigCtxPtr(ctx)
	if err != nil {
		return err
	}

	C.xmlSecDSigCtxDestroy(ctxptr)
	return nil
}

func xmlSecCryptoAppKeyLoadMemory(buf []byte, format KeyDataFormat) (*Key, error) {
	key := C.xmlSecCryptoAppKeyLoadMemory(
		(*C.xmlSecByte)(unsafe.Pointer(&buf[0])),
		(C.xmlSecSize)(C.int(len(buf))),
		(C.xmlSecKeyDataFormat)(format),
		nil,
		nil,
		nil,
	)
	if key == nil {
		return nil, errors.New("failed to load key")
	}
	return &Key{ptr: key}, nil
}

func xmlSecCryptoAppKeyLoad(file string, format KeyDataFormat) (*Key, error) {
	cfile := C.CString(file)
	defer C.free(unsafe.Pointer(cfile))
	key := C.xmlSecCryptoAppKeyLoad(cfile, (C.xmlSecKeyDataFormat)(format), nil, nil, nil)
	if key == nil {
		return nil, errors.New("failed to load key")
	}

	if C.xmlSecKeySetName(key, (*C.xmlChar)(unsafe.Pointer(cfile))) < C.int(0) {
		return nil, errors.New("failed to set key name")
	}

	return &Key{ptr: key}, nil
}

func (ctx *DSigCtx) SetKey(key *Key) error {
	ctxptr, err := validDSigCtxPtr(ctx)
	if err != nil {
		return err
	}

	ctxptr.signKey = key.ptr
	return nil
}

func xmlSecDSigCtxSignRaw(ctxptr *C.xmlSecDSigCtx, nodeptr *C.xmlNode) error {
	if C.xmlSecDSigCtxSign(ctxptr, nodeptr) < C.int(0) {
		return errors.New("failed to sign node")
	}
	return nil
}

func xmlSecDSigCtxSignNode(ctx *DSigCtx, n libxml2.Node) error {
	ctxptr, err := validDSigCtxPtr(ctx)
	if err != nil {
		return err
	}

	nodeptr, err := validNodePtr(n)
	if err != nil {
		return err
	}

	return xmlSecDSigCtxSignRaw(ctxptr, nodeptr)
}

func xmlSecDSigCtxSignDocument(ctx *DSigCtx, doc *libxml2.Document) error {
	ctxptr, err := validDSigCtxPtr(ctx)
	if err != nil {
		return err
	}

	root, err := doc.DocumentElement()
	if err != nil {
		return err
	}

	rootptr, err := validNodePtr(root)
	if err != nil {
		return err
	}

	cname := stringToXMLChar(SignatureNode)
	cns := stringToXMLChar(DSigNs)
	defer C.free(unsafe.Pointer(cname))
	defer C.free(unsafe.Pointer(cns))

	nodeptr := C.xmlSecFindNode(rootptr, cname, cns)
	if nodeptr == nil {
		return errors.New("failed to find start node")
	}

	return xmlSecDSigCtxSignRaw(ctxptr, nodeptr)
}

func xmlSecDSigCtxVerifyRaw(ctxptr *C.xmlSecDSigCtx, nodeptr *C.xmlNode) error {
	if C.xmlSecDSigCtxVerify(ctxptr, nodeptr) < C.int(0) {
		return errors.New("failed to verify node")
	}

	if ctxptr.status != C.xmlSecDSigStatusSucceeded {
		return errors.New("signature verification failed")
	}
	return nil
}

func xmlSecDSigCtxVerifyNode(ctx *DSigCtx, n libxml2.Node) error {
	ctxptr, err := validDSigCtxPtr(ctx)
	if err != nil {
		return err
	}

	nodeptr, err := validNodePtr(n)
	if err != nil {
		return err
	}

	return xmlSecDSigCtxVerifyRaw(ctxptr, nodeptr)
}

func xmlSecDSigCtxVerifyDocument(ctx *DSigCtx, doc *libxml2.Document) error {
	ctxptr, err := validDSigCtxPtr(ctx)
	if err != nil {
		return err
	}

	root, err := doc.DocumentElement()
	if err != nil {
		return err
	}

	rootptr, err := validNodePtr(root)
	if err != nil {
		return err
	}

	cname := stringToXMLChar(SignatureNode)
	cns := stringToXMLChar(DSigNs)
	defer C.free(unsafe.Pointer(cname))
	defer C.free(unsafe.Pointer(cns))

	nodeptr := C.xmlSecFindNode(rootptr, cname, cns)
	if nodeptr == nil {
		return errors.New("failed to find start node")
	}

	return xmlSecDSigCtxVerifyRaw(ctxptr, nodeptr)
}

func xmlSecTmplSignatureCreate(doc *libxml2.Document, c14nMethod TransformID, signMethod TransformID, id string) (libxml2.Node, error) {
	docptr := (*C.xmlDoc)(doc.Pointer())
	if docptr == nil {
		return nil, libxml2.ErrInvalidDocument
	}

	var idxml *C.xmlChar
	if id != "" {
		idxml = stringToXMLChar(id)
		defer C.free(unsafe.Pointer(idxml))
	}
	ptr := C.xmlSecTmplSignatureCreate(
		docptr,
		c14nMethod.ptr,
		signMethod.ptr,
		idxml,
	)
	if ptr == nil {
		return nil, errors.New("failed to create signature template")
	}

	return libxml2.WrapToNodeUnsafe(unsafe.Pointer(ptr))
}

func xmlSecTmplSignatureAddReference(signode libxml2.Node, digestMethod TransformID, id, uri, nodeType string) (libxml2.Node, error) {
	nptr, err := validNodePtr(signode)
	if err != nil {
		return nil, err
	}

	var idxml, urixml, typexml *C.xmlChar
	if id != "" {
		idxml = stringToXMLChar(id)
		defer C.free(unsafe.Pointer(idxml))
	}
	if uri != "" {
		urixml = stringToXMLChar(uri)
		defer C.free(unsafe.Pointer(urixml))
	}
	if nodeType != "" {
		typexml = stringToXMLChar(nodeType)
		defer C.free(unsafe.Pointer(typexml))
	}

	ptr := C.xmlSecTmplSignatureAddReference(
		nptr,
		digestMethod.ptr,
		idxml,
		urixml,
		typexml,
	)
	if ptr == nil {
		return nil, errors.New("failed to add reference")
	}

	return libxml2.WrapToNodeUnsafe(unsafe.Pointer(ptr))
}

func xmlSecTmplReferenceAddTransform(n libxml2.Node, transformId TransformID) (libxml2.Node, error) {
	nptr, err := validNodePtr(n)
	if err != nil {
		return nil, err
	}

	ptr := C.xmlSecTmplReferenceAddTransform(
		nptr,
		transformId.ptr,
	)
	if ptr == nil {
		return nil, errors.New("failed to add transform")
	}

	return libxml2.WrapToNodeUnsafe(unsafe.Pointer(ptr))
}

func xmlSecTmplSignatureEnsureKeyInfo(n libxml2.Node, id string) (libxml2.Node, error) {
	nptr, err := validNodePtr(n)
	if err != nil {
		return nil, err
	}
	var idc *C.xmlChar
	if id != "" {
		idc = stringToXMLChar(id)
		defer C.free(unsafe.Pointer(idc))
	}

	ret := C.xmlSecTmplSignatureEnsureKeyInfo(nptr, idc)
	if ret == nil {
		return nil, errors.New("failed to add KeyInfo node")
	}

	return libxml2.WrapToNodeUnsafe(unsafe.Pointer(ret))
}

func xmlSecTmplKeyInfoAddX509Data(n libxml2.Node) (libxml2.Node, error) {
	nptr, err := validNodePtr(n)
	if err != nil {
		return nil, err
	}

	ret := C.xmlSecTmplKeyInfoAddX509Data(nptr)
	if ret == nil {
		return nil, errors.New("failed to add X509Data node")
	}

	return libxml2.WrapToNodeUnsafe(unsafe.Pointer(ret))
}

//func xmlSecCryptoAppKeyCertLoad(ctx *DSigCtx, certFile string, format KeyDataFormat) error {
