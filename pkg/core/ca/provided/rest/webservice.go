package rest

import (
	"context"

	"github.com/emicklei/go-restful"

	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/ca/provided"
	"github.com/Kong/kuma/pkg/core/ca/provided/rest/types"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	rest_errors "github.com/Kong/kuma/pkg/core/rest/errors"
	errors_types "github.com/Kong/kuma/pkg/core/rest/errors/types"
	"github.com/Kong/kuma/pkg/core/validators"
	"github.com/Kong/kuma/pkg/tls"
)

var logger = core.Log.WithName("ca-provided-ws")

type providedCAWebservice struct {
	providedCaManager provided.ProvidedCaManager
	resourceManager   manager.ResourceManager
}

func NewWebservice(providedCaManager provided.ProvidedCaManager, resourceManager manager.ResourceManager) *restful.WebService {
	caWs := providedCAWebservice{
		providedCaManager: providedCaManager,
		resourceManager:   resourceManager,
	}
	return caWs.createWs()
}

func (p *providedCAWebservice) createWs() *restful.WebService {
	ws := new(restful.WebService).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Path("/meshes/{mesh}/ca/provided").
		Route(ws.POST("/certificates").To(p.addSigningCertificate)).
		Route(ws.GET("/certificates").To(p.signingCertificates)).
		Route(ws.DELETE("/certificates/{id}").To(p.deleteSigningCertificate))
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
	signingCert, err := p.providedCaManager.AddSigningCert(request.Request.Context(), mesh, keyPair)
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
	if err := p.validateCaActive(request.Request.Context(), mesh); err != nil {
		handleError(response, err, "Could not delete signing cert")
		return
	}
	if err := p.providedCaManager.DeleteSigningCert(request.Request.Context(), mesh, id); err != nil {
		handleError(response, err, "Could not delete signing cert")
		return
	}
}

func (p *providedCAWebservice) validateCaActive(ctx context.Context, mesh string) error {
	meshRes := &core_mesh.MeshResource{}
	if err := p.resourceManager.Get(ctx, meshRes, store.GetByKey(mesh, mesh)); err != nil && !store.IsResourceNotFound(err) {
		// we ignore error if mesh is not found because the user may want to delete leftovers if something went wrong during mesh removal
		return err
	}
	if meshRes.Spec.GetMtls().GetEnabled() {
		return &providedCaActiveError{}
	}
	return nil
}

type providedCaActiveError struct {
}

func (p *providedCaActiveError) Error() string {
	return "Cannot delete signing certificate when the mTLS in the mesh is active and type of CA is Provided"
}

func (p *providedCAWebservice) signingCertificates(request *restful.Request, response *restful.Response) {
	mesh := request.PathParameter("mesh")
	certs, err := p.providedCaManager.GetSigningCerts(request.Request.Context(), mesh)
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
		kumaErr := errors_types.Error{
			Title:   title,
			Details: "Not found",
		}
		if err := response.WriteHeaderAndJson(404, kumaErr, "application/json"); err != nil {
			logger.Error(err, "Could not write the error response")
		}
	case *providedCaActiveError:
		kumaErr := errors_types.Error{
			Title:   title,
			Details: err.Error(),
		}
		if err := response.WriteHeaderAndJson(400, kumaErr, "application/json"); err != nil {
			logger.Error(err, "Could not write the error response")
		}
	default:
		rest_errors.HandleError(response, err, title)
	}
}
