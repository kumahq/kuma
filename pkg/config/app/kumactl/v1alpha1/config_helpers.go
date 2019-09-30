package v1alpha1

import (
	"encoding/json"
	api_server "github.com/Kong/kuma/pkg/api-server"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

func (cfg *Configuration) GetCurrent() *Context {
	_, c := cfg.GetContext(cfg.CurrentContext)
	return c
}

func (cfg *Configuration) GetContext(name string) (int, *Context) {
	for i, c := range cfg.Contexts {
		if c.Name == name {
			return i, c
		}
	}
	return -1, nil
}

func (cfg *Configuration) AddContext(c *Context) bool {
	_, old := cfg.GetContext(c.Name)
	if old != nil {
		return false
	}
	cfg.Contexts = append(cfg.Contexts, c)
	return true
}

func (cfg *Configuration) RemoveContext(name string) bool {
	i, old := cfg.GetContext(name)
	if old == nil {
		return false
	}
	cfg.Contexts = append(cfg.Contexts[:i], cfg.Contexts[i+1:]...)
	if cfg.CurrentContext == name {
		cfg.CurrentContext = ""
		if len(cfg.Contexts) > 0 {
			cfg.CurrentContext = cfg.Contexts[0].Name
		}
	}
	return true
}

func (cfg *Configuration) SwitchContext(name string) bool {
	_, new := cfg.GetContext(name)
	if new == nil {
		return false
	}
	cfg.CurrentContext = new.Name
	return true
}

func (cfg *Configuration) GetControlPlane(name string) (int, *ControlPlane) {
	for i, p := range cfg.ControlPlanes {
		if p.Name == name {
			return i, p
		}
	}
	return -1, nil
}

func (cfg *Configuration) AddControlPlane(cp *ControlPlane) error {
	_, old := cfg.GetControlPlane(cp.Name)
	if old != nil {
		return errors.Errorf("Control Plane with name %q already exists", cp.Name)
	}
	if err := validateCpCoordinates(cp); err != nil {
		return err
	}

	cfg.ControlPlanes = append(cfg.ControlPlanes, cp)
	return nil
}

func validateCpCoordinates(cp *ControlPlane) error {
	resp, err := http.Get(cp.Coordinates.ApiServer.Url)
	if err != nil {
		return errors.Wrap(err, "could not connect to the Control Plane API Server")
	}
	if resp.StatusCode != 200 {
		return errors.New("Control Plane API Server is not responding")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "could not read body from the Control Plane API Server")
	}
	response := api_server.IndexResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return errors.Wrap(err, "could not unmarshal body from the Control Plane API Server. Provided address is not valid Kuma Control Plane API Server")
	}
	if response.Tagline != api_server.TaglineKuma {
		return errors.New("provided address is not valid Kuma Control Plane API Server")
	}
	return nil
}

func (cfg *Configuration) RemoveControlPlane(name string) bool {
	i, old := cfg.GetControlPlane(name)
	if old == nil {
		return false
	}
	for _, context := range cfg.Contexts {
		if context.ControlPlane == name {
			cfg.RemoveContext(context.Name)
		}
	}
	cfg.ControlPlanes = append(cfg.ControlPlanes[:i], cfg.ControlPlanes[i+1:]...)
	return true
}
