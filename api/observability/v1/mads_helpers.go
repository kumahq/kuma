package v1

import "fmt"

// GetName returns a name including the mesh which is used for ResourceName in MADS snapshot cache
func (x *MonitoringAssignment) GetName() string {
	if x != nil {
		dpName := ""
		if len(x.Targets) > 0 {
			dpName = x.Targets[0].GetName()
		}
		return fmt.Sprintf("/meshes/%s/dataplanes/%s", x.GetMesh(), dpName)
	}
	return ""
}
