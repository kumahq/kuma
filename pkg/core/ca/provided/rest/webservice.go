package rest

import (
	api_server "github.com/Kong/kuma/pkg/api-server"
	api_server_types "github.com/Kong/kuma/pkg/api-server/types"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/ca/provided"
	"github.com/Kong/kuma/pkg/core/ca/provided/rest/types"
	"github.com/Kong/kuma/pkg/core/validators"
	"github.com/Kong/kuma/pkg/tls"
	"github.com/emicklei/go-restful"
)

var logger = core.Log.WithName("ca-provided-ws")

type providedCAWebservice struct {
	manager provided.ProvidedCaManager
}

func NewWebservice(manager provided.ProvidedCaManager) *restful.WebService {
	caWs := providedCAWebservice{manager}
	return caWs.createWs()
}

func (p *providedCAWebservice) createWs() *restful.WebService {
	ws := new(restful.WebService).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Path("/meshes/{mesh}/ca/provided").
		Route(ws.POST("/certificates").To(p.addSigningCertificate)).
		Route(ws.GET("/certificates").To(p.signingCertificates)).
		Route(ws.DELETE("/certificates/{id}").To(p.deleteSigningCertificate)).
		Route(ws.DELETE("").To(p.deleteCa))
	return ws
}

func (p *providedCAWebservice) addSigningCertificate(request *restful.Request, response *restful.Response) {
	reqPair := types.KeyPair{}
	if err := request.ReadEntity(&reqPair); err != nil {
		handleError(response, err, "Could not process the key pair")
		return
	}
	verr := validators.ValidationError{}
	if reqPair.Cert == "" {
		verr.AddViolation("cert", "must not be empty")
	}
	if reqPair.Key == "" {
		verr.AddViolation("key", "must not be empty")
	}
	if verr.HasViolations() {
		handleError(response, verr.OrNil(), "Could not add signing cert")
		return
	}

	keyPair := tls.KeyPair{
		CertPEM: []byte(reqPair.Cert),
		KeyPEM:  []byte(reqPair.Key),
	}
	mesh := request.PathParameter("mesh")
	signingCert, err := p.manager.AddSigningCert(request.Request.Context(), mesh, keyPair)
	if err != nil {
		handleError(response, err, "Could not add signing cert")
		return
	}

	certResp := types.SigningCert{
		Id:   signingCert.Id,
		Cert: string(signingCert.Cert),
	}
	if err := response.WriteAsJson(certResp); err != nil {
		handleError(response, err, "Could not add signing cert")
	}
}

func (p *providedCAWebservice) deleteSigningCertificate(request *restful.Request, response *restful.Response) {
	mesh := request.PathParameter("mesh")
	id := request.PathParameter("id")
	if err := p.manager.DeleteSigningCert(request.Request.Context(), mesh, id); err != nil {
		handleError(response, err, "Could not delete signing cert")
		return
	}
}

func (p *providedCAWebservice) deleteCa(request *restful.Request, response *restful.Response) {
	mesh := request.PathParameter("mesh")
	if err := p.manager.DeleteCa(request.Request.Context(), mesh); err != nil {
		handleError(response, err, "Could not delete CA")
		return
	}
}

func (p *providedCAWebservice) signingCertificates(request *restful.Request, response *restful.Response) {
	mesh := request.PathParameter("mesh")
	certs, err := p.manager.GetSigningCerts(request.Request.Context(), mesh)
	if err != nil {
		handleError(response, err, "Could not retrieve signing certs")
		return
	}
	signingCerts := []types.SigningCert{}
	for _, cert := range certs {
		signingCerts = append(signingCerts, types.SigningCert{
			Id:   cert.Id,
			Cert: string(cert.Cert),
		})
	}
	if err := response.WriteAsJson(signingCerts); err != nil {
		logger.Error(err, "Could not write the error response")
	}
}

func handleError(response *restful.Response, err error, title string) {
	switch err.(type) {
	case *provided.SigningCertNotFound:
		kumaErr := api_server_types.Error{
			Title:   title,
			Details: "Not found",
		}
		if err := response.WriteHeaderAndJson(404, kumaErr, "application/json"); err != nil {
			logger.Error(err, "Could not write the error response")
		}
	default:
		api_server.HandleError(response, err, title)
	}
}
